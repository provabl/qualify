// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package training

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/lib/pq"
	"github.com/provabl/qualify/internal/database"
)

// iamTagWriter is the interface for writing IAM role tags.
// Defined as an interface to enable mocking in tests.
type iamTagWriter interface {
	TagRole(ctx context.Context, params *iam.TagRoleInput, optFns ...func(*iam.Options)) (*iam.TagRoleOutput, error)
}

// moduleTagMap maps qualify training module IDs to attest:* IAM tag keys.
// When a researcher completes a module, qualify writes these tags to their IAM role
// so attest's principal resolver can evaluate them in Cedar policies.
var moduleTagMap = map[string]string{
	"cui-fundamentals":       "attest:cui-training",
	"hipaa-privacy-security": "attest:hipaa-training",
	"security-awareness":     "attest:awareness-training",
	"ferpa-basics":           "attest:ferpa-training",
	"itar-export-control":    "attest:itar-training",
	"data-classification":    "attest:data-class-training",
	"nih-research-security":  "attest:research-security-training",
}

// defaultTrainingExpiry is how long a training certification is valid.
const defaultTrainingExpiry = 365 * 24 * time.Hour

// Service provides training policy enforcement functionality
type Service struct {
	db        *database.DB
	iamTagger iamTagWriter // optional; nil = no IAM tagging
	awsRegion string
}

// NewService creates a new training service
func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

// NewServiceWithIAM creates a training service that writes attest:* tags to IAM roles
// on training completion. The userRoleARNFunc is called to resolve a user's IAM role ARN.
func NewServiceWithIAM(ctx context.Context, db *database.DB, region string) *Service {
	svc := &Service{db: db, awsRegion: region}
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		log.Printf("ark/training: could not init IAM client for attest tagging: %v", err)
		return svc
	}
	svc.iamTagger = iam.NewFromConfig(cfg)
	return svc
}

// CheckTrainingGate evaluates if a user meets training requirements for an action
func (s *Service) CheckTrainingGate(ctx context.Context, userID string, action string) (*PolicyDecision, error) {
	// Query for policies that apply to this action
	query := `
		SELECT rules
		FROM policies
		WHERE policy_type = 'training_gate'
		  AND rules->'actions' @> $1::jsonb
		  AND status = 'active'
	`

	actionJSON := fmt.Sprintf(`["%s"]`, action)
	rows, err := s.db.QueryContext(ctx, query, actionJSON)
	if err != nil {
		return nil, fmt.Errorf("query policies: %w", err)
	}
	defer rows.Close()

	// Collect required modules from all matching policies
	requiredModules := make(map[string]bool)
	for rows.Next() {
		var rulesJSON []byte
		if err := rows.Scan(&rulesJSON); err != nil {
			return nil, fmt.Errorf("scan rules: %w", err)
		}

		var rules struct {
			RequiredModules []string `json:"required_modules"`
		}
		if err := json.Unmarshal(rulesJSON, &rules); err != nil {
			return nil, fmt.Errorf("unmarshal rules: %w", err)
		}

		for _, moduleName := range rules.RequiredModules {
			requiredModules[moduleName] = true
		}
	}

	// If no policies match, allow the action
	if len(requiredModules) == 0 {
		return &PolicyDecision{
			Action:  "allow",
			Message: "No training requirements for this action",
		}, nil
	}

	// Check if user has completed all required modules
	moduleNames := make([]string, 0, len(requiredModules))
	for name := range requiredModules {
		moduleNames = append(moduleNames, name)
	}

	// Query for incomplete modules
	incompleteQuery := `
		SELECT tm.id, tm.name, tm.title, tm.estimated_minutes
		FROM training_modules tm
		LEFT JOIN user_training_progress utp
			ON tm.id = utp.module_id AND utp.user_id = $1 AND utp.status = 'completed'
		WHERE tm.name = ANY($2)
		  AND utp.id IS NULL
	`

	incompleteRows, err := s.db.QueryContext(ctx, incompleteQuery, userID, pq.Array(moduleNames))
	if err != nil {
		return nil, fmt.Errorf("query incomplete modules: %w", err)
	}
	defer incompleteRows.Close()

	var incomplete []Module
	for incompleteRows.Next() {
		var module Module
		if err := incompleteRows.Scan(&module.ID, &module.Name, &module.Title, &module.EstimatedMinutes); err != nil {
			return nil, fmt.Errorf("scan module: %w", err)
		}
		incomplete = append(incomplete, module)
	}

	// If any modules are incomplete, block the action
	if len(incomplete) > 0 {
		return &PolicyDecision{
			Action:          "block",
			Reason:          "training_required",
			RequiredModules: incomplete,
			Message:         "Complete required training modules to perform this operation",
		}, nil
	}

	// All required modules completed - allow
	return &PolicyDecision{
		Action:  "allow",
		Message: "Training requirements met",
	}, nil
}

// GetUserProgress retrieves training progress for a user
func (s *Service) GetUserProgress(ctx context.Context, userID string) ([]Progress, error) {
	query := `
		SELECT
			tm.id,
			tm.name,
			COALESCE(utp.status, 'not_started') as status,
			utp.completed_at
		FROM training_modules tm
		LEFT JOIN user_training_progress utp
			ON tm.id = utp.module_id AND utp.user_id = $1
		ORDER BY tm.name
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query progress: %w", err)
	}
	defer rows.Close()

	var progress []Progress
	for rows.Next() {
		var p Progress
		var completedAt sql.NullTime

		if err := rows.Scan(&p.ModuleID, &p.ModuleName, &p.Status, &completedAt); err != nil {
			return nil, fmt.Errorf("scan progress: %w", err)
		}

		if completedAt.Valid {
			p.CompletedAt = completedAt.Time.Format("2006-01-02T15:04:05Z")
		}

		progress = append(progress, p)
	}

	return progress, nil
}

// ListModules returns all active training modules
func (s *Service) ListModules(ctx context.Context) ([]Module, error) {
	query := `
		SELECT id, name, title, description, category, difficulty, estimated_minutes, status
		FROM training_modules
		WHERE status = 'active'
		ORDER BY category, difficulty, name
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query modules: %w", err)
	}
	defer rows.Close()

	var modules []Module
	for rows.Next() {
		var m Module
		var description, category, difficulty, status sql.NullString

		if err := rows.Scan(&m.ID, &m.Name, &m.Title, &description, &category, &difficulty, &m.EstimatedMinutes, &status); err != nil {
			return nil, fmt.Errorf("scan module: %w", err)
		}

		if description.Valid {
			m.Description = description.String
		}
		if category.Valid {
			m.Category = category.String
		}
		if difficulty.Valid {
			m.Difficulty = difficulty.String
		}
		if status.Valid {
			m.Status = status.String
		}

		modules = append(modules, m)
	}

	return modules, nil
}

// GetModule returns a specific module with full content
func (s *Service) GetModule(ctx context.Context, moduleID string) (*Module, error) {
	query := `
		SELECT id, name, title, description, category, difficulty, estimated_minutes, content, status, prerequisites
		FROM training_modules
		WHERE name = $1 OR (id::text = $1)
	`

	var m Module
	var description, category, difficulty, status sql.NullString
	var content []byte
	var prerequisites pq.StringArray

	err := s.db.QueryRowContext(ctx, query, moduleID).Scan(
		&m.ID, &m.Name, &m.Title, &description, &category, &difficulty, &m.EstimatedMinutes, &content, &status, &prerequisites,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("module not found: %s", moduleID)
	}
	if err != nil {
		return nil, fmt.Errorf("query module: %w", err)
	}

	if description.Valid {
		m.Description = description.String
	}
	if category.Valid {
		m.Category = category.String
	}
	if difficulty.Valid {
		m.Difficulty = difficulty.String
	}
	if status.Valid {
		m.Status = status.String
	}
	if len(content) > 0 {
		m.Content = content
	}
	if len(prerequisites) > 0 {
		m.Prerequisites = prerequisites
	}

	return &m, nil
}

// StartModule marks a module as started for a user
func (s *Service) StartModule(ctx context.Context, userID, moduleID string) error {
	// Check if progress record already exists
	checkQuery := `
		SELECT id FROM user_training_progress
		WHERE user_id = $1 AND module_id = $2
	`

	var existingID string
	err := s.db.QueryRowContext(ctx, checkQuery, userID, moduleID).Scan(&existingID)

	if err == sql.ErrNoRows {
		// Insert new progress record
		insertQuery := `
			INSERT INTO user_training_progress (user_id, module_id, status, started_at)
			VALUES ($1, $2, 'in_progress', NOW())
		`
		_, err = s.db.ExecContext(ctx, insertQuery, userID, moduleID)
		if err != nil {
			return fmt.Errorf("insert progress: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("check existing progress: %w", err)
	} else {
		// Update existing record to in_progress if not already completed
		updateQuery := `
			UPDATE user_training_progress
			SET status = 'in_progress', started_at = COALESCE(started_at, NOW())
			WHERE user_id = $1 AND module_id = $2 AND status != 'completed'
		`
		_, err = s.db.ExecContext(ctx, updateQuery, userID, moduleID)
		if err != nil {
			return fmt.Errorf("update progress: %w", err)
		}
	}

	return nil
}

// CompleteModule marks a module as completed for a user.
// If the service was initialized with IAM credentials and the user has a roleARN
// registered, it writes attest:* tags to enable Cedar policy enforcement in attest.
func (s *Service) CompleteModule(ctx context.Context, userID, moduleID string, score int) error {
	// Upsert progress record.
	query := `
		INSERT INTO user_training_progress (user_id, module_id, status, started_at, completed_at, score)
		VALUES ($1, $2, 'completed', NOW(), NOW(), $3)
		ON CONFLICT (user_id, module_id)
		DO UPDATE SET
			status = 'completed',
			completed_at = NOW(),
			score = $3
	`
	_, err := s.db.ExecContext(ctx, query, userID, moduleID, score)
	if err != nil {
		return fmt.Errorf("complete module: %w", err)
	}

	// Record completion activity.
	_ = s.RecordActivity(ctx, userID, "module_completed", moduleID, map[string]interface{}{
		"score": score,
	})

	// Write attest:* tags to the user's IAM role (non-fatal if it fails).
	// This enables attest's Cedar policies to evaluate principal.cui_training_current etc.
	if s.iamTagger != nil {
		roleARN := s.getUserRoleARN(ctx, userID)
		if roleARN != "" {
			expiresAt := time.Now().Add(defaultTrainingExpiry)
			if tagErr := s.writeAttestTags(ctx, roleARN, moduleID, expiresAt); tagErr != nil {
				// Non-fatal: log but don't fail training completion.
				log.Printf("ark/training: could not write attest tags for user %s module %s: %v", userID, moduleID, tagErr)
			}
		}
	}

	return nil
}

// getUserRoleARN retrieves the IAM role ARN associated with a user.
// The role ARN is stored in user metadata (attest:role-arn tag on the user record).
func (s *Service) getUserRoleARN(ctx context.Context, userID string) string {
	var roleARN string
	row := s.db.QueryRowContext(ctx,
		`SELECT metadata->>'role_arn' FROM users WHERE id = $1`, userID)
	_ = row.Scan(&roleARN)
	return roleARN
}

// writeAttestTags writes attest:* IAM role tags when training is completed.
// Enables attest's principal resolver to read training status for Cedar evaluation.
func (s *Service) writeAttestTags(ctx context.Context, roleARN, moduleID string, expiresAt time.Time) error {
	tagKey, ok := moduleTagMap[moduleID]
	if !ok {
		return nil // no attest mapping for this module
	}

	roleName := extractRoleName(roleARN)
	if roleName == "" {
		return fmt.Errorf("could not extract role name from ARN: %s", roleARN)
	}

	tags := []iamtypes.Tag{
		{Key: aws.String(tagKey), Value: aws.String("true")},
		{Key: aws.String(tagKey + "-expiry"), Value: aws.String(expiresAt.Format(time.RFC3339))},
	}

	_, err := s.iamTagger.TagRole(ctx, &iam.TagRoleInput{
		RoleName: aws.String(roleName),
		Tags:     tags,
	})
	if err != nil {
		return fmt.Errorf("tagging IAM role %s: %w", roleName, err)
	}
	return nil
}

// extractRoleName extracts the role name from an IAM role ARN.
// "arn:aws:iam::123456789012:role/my-role" → "my-role"
func extractRoleName(arn string) string {
	const prefix = ":role/"
	idx := strings.LastIndex(arn, prefix)
	if idx == -1 {
		return ""
	}
	name := arn[idx+len(prefix):]
	if i := strings.LastIndex(name, "/"); i != -1 {
		name = name[i+1:]
	}
	return name
}

// SubmitQuiz evaluates quiz answers and returns results
func (s *Service) SubmitQuiz(ctx context.Context, moduleID string, submission QuizSubmission) (*QuizResponse, error) {
	// Fetch module content to get correct answers
	module, err := s.GetModule(ctx, moduleID)
	if err != nil {
		return nil, fmt.Errorf("get module: %w", err)
	}

	// Parse module content to extract quiz questions
	var content struct {
		Sections []struct {
			Type      string `json:"type"`
			Questions []struct {
				ID            string   `json:"id"`
				Question      string   `json:"question"`
				Options       []string `json:"options"`
				CorrectAnswer int      `json:"correct_answer"`
				Explanation   string   `json:"explanation,omitempty"`
			} `json:"questions,omitempty"`
		} `json:"sections"`
	}

	if err := json.Unmarshal(module.Content, &content); err != nil {
		return nil, fmt.Errorf("parse module content: %w", err)
	}

	// Build map of question ID to correct answer
	questionMap := make(map[string]struct {
		correctAnswer int
		explanation   string
	})

	for _, section := range content.Sections {
		if section.Type == "quiz" {
			for _, q := range section.Questions {
				questionMap[q.ID] = struct {
					correctAnswer int
					explanation   string
				}{
					correctAnswer: q.CorrectAnswer,
					explanation:   q.Explanation,
				}
			}
		}
	}

	// Evaluate answers
	var results []QuizResult
	correctCount := 0

	for _, answer := range submission.Answers {
		question, exists := questionMap[answer.QuestionID]
		if !exists {
			return nil, fmt.Errorf("invalid question ID: %s", answer.QuestionID)
		}

		isCorrect := answer.SelectedAnswer == question.correctAnswer
		if isCorrect {
			correctCount++
		}

		results = append(results, QuizResult{
			QuestionID:     answer.QuestionID,
			Correct:        isCorrect,
			SelectedAnswer: answer.SelectedAnswer,
			CorrectAnswer:  question.correctAnswer,
			Explanation:    question.explanation,
		})
	}

	// Calculate score percentage
	totalQuestions := len(submission.Answers)
	score := 0
	if totalQuestions > 0 {
		score = (correctCount * 100) / totalQuestions
	}

	// Save quiz answers to user_training_progress
	answersJSON, _ := json.Marshal(submission.Answers)
	updateQuery := `
		UPDATE user_training_progress
		SET quiz_answers = $1
		WHERE user_id = $2 AND module_id = $3
	`
	_, err = s.db.ExecContext(ctx, updateQuery, answersJSON, submission.UserID, moduleID)
	if err != nil {
		return nil, fmt.Errorf("save quiz answers: %w", err)
	}

	// Record quiz activity
	activityType := "quiz_passed"
	if score < 70 { // Passing threshold
		activityType = "quiz_failed"
	}
	_ = s.RecordActivity(ctx, submission.UserID, activityType, moduleID, map[string]interface{}{
		"score": score,
	})

	return &QuizResponse{
		Score:          score,
		TotalQuestions: totalQuestions,
		CorrectAnswers: correctCount,
		Results:        results,
	}, nil
}

// GetDashboardStats aggregates user training data for dashboard display
func (s *Service) GetDashboardStats(ctx context.Context, userID string) (*DashboardStats, error) {
	// Get training summary statistics
	summaryQuery := `
		SELECT
			COUNT(*) as total_modules,
			COUNT(CASE WHEN utp.status = 'completed' THEN 1 END) as completed,
			COUNT(CASE WHEN utp.status = 'in_progress' THEN 1 END) as in_progress,
			COUNT(CASE WHEN utp.status IS NULL OR utp.status = 'not_started' THEN 1 END) as not_started,
			AVG(CASE WHEN utp.status = 'completed' AND utp.score IS NOT NULL THEN utp.score END) as avg_score
		FROM training_modules tm
		LEFT JOIN user_training_progress utp
			ON tm.id = utp.module_id AND utp.user_id = $1
		WHERE tm.status = 'active'
	`

	var totalModules, completed, inProgress, notStarted int
	var avgScore sql.NullFloat64

	err := s.db.QueryRowContext(ctx, summaryQuery, userID).Scan(
		&totalModules, &completed, &inProgress, &notStarted, &avgScore,
	)
	if err != nil {
		return nil, fmt.Errorf("query training summary: %w", err)
	}

	completionPercentage := 0
	if totalModules > 0 {
		completionPercentage = (completed * 100) / totalModules
	}

	var averageScore *int
	if avgScore.Valid {
		score := int(avgScore.Float64)
		averageScore = &score
	}

	summary := TrainingSummary{
		TotalModules:         totalModules,
		Completed:            completed,
		InProgress:           inProgress,
		NotStarted:           notStarted,
		CompletionPercentage: completionPercentage,
		AverageScore:         averageScore,
	}

	// Get recent activity
	recentActivity, err := s.GetUserActivity(ctx, userID, 10, 0)
	if err != nil {
		return nil, fmt.Errorf("get recent activity: %w", err)
	}

	// Get available operations (mock for now - would integrate with policy service)
	// For MVP, return static operations based on completed modules
	unlockedOps := []string{}
	lockedOps := []string{"ec2:RunInstance", "iam:CreateUser"}

	// Check if S3 training is complete
	s3Query := `
		SELECT COUNT(*)
		FROM training_modules tm
		JOIN user_training_progress utp ON tm.id = utp.module_id
		WHERE tm.category = 's3'
		  AND utp.user_id = $1
		  AND utp.status = 'completed'
	`
	var s3Count int
	err = s.db.QueryRowContext(ctx, s3Query, userID).Scan(&s3Count)
	if err == nil && s3Count > 0 {
		unlockedOps = append(unlockedOps, "s3:CreateBucket", "s3:ListBuckets")
	} else {
		lockedOps = append([]string{"s3:CreateBucket", "s3:ListBuckets"}, lockedOps...)
	}

	return &DashboardStats{
		UserID:          userID,
		TrainingSummary: summary,
		RecentActivity:  recentActivity,
		AvailableOps: AvailableOps{
			Unlocked: unlockedOps,
			Locked:   lockedOps,
		},
	}, nil
}

// GetUserActivity retrieves paginated training activity log
func (s *Service) GetUserActivity(ctx context.Context, userID string, limit, offset int) ([]ActivityItem, error) {
	query := `
		SELECT id, activity_type, module_id, module_name, score, metadata, created_at
		FROM training_activities
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query activities: %w", err)
	}
	defer rows.Close()

	var activities []ActivityItem
	for rows.Next() {
		var activity ActivityItem
		var moduleID, moduleName sql.NullString
		var score sql.NullInt64
		var metadataJSON []byte
		var createdAt sql.NullTime

		if err := rows.Scan(&activity.ID, &activity.Type, &moduleID, &moduleName, &score, &metadataJSON, &createdAt); err != nil {
			return nil, fmt.Errorf("scan activity: %w", err)
		}

		if moduleID.Valid {
			activity.ModuleID = moduleID.String
		}
		if moduleName.Valid {
			activity.ModuleName = moduleName.String
		}
		if score.Valid {
			scoreInt := int(score.Int64)
			activity.Score = &scoreInt
		}
		if createdAt.Valid {
			activity.Timestamp = createdAt.Time.Format("2006-01-02T15:04:05Z")
		}
		if len(metadataJSON) > 0 {
			var metadata map[string]interface{}
			if err := json.Unmarshal(metadataJSON, &metadata); err == nil {
				activity.Metadata = metadata
			}
		}

		activities = append(activities, activity)
	}

	if activities == nil {
		activities = []ActivityItem{}
	}

	return activities, nil
}

// RecordActivity logs a training activity event
func (s *Service) RecordActivity(ctx context.Context, userID, activityType, moduleID string, metadata map[string]interface{}) error {
	// Get module name if moduleID provided
	var moduleName string
	if moduleID != "" {
		nameQuery := `SELECT name FROM training_modules WHERE id = $1 OR name = $1`
		_ = s.db.QueryRowContext(ctx, nameQuery, moduleID).Scan(&moduleName)
	}

	// Extract score from metadata if present
	var score *int
	if metadata != nil {
		if scoreVal, ok := metadata["score"]; ok {
			if scoreInt, ok := scoreVal.(int); ok {
				score = &scoreInt
			}
		}
	}

	// Serialize metadata
	metadataJSON, _ := json.Marshal(metadata)

	query := `
		INSERT INTO training_activities (user_id, activity_type, module_id, module_name, score, metadata)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), $5, $6)
	`

	_, err := s.db.ExecContext(ctx, query, userID, activityType, moduleID, moduleName, score, metadataJSON)
	if err != nil {
		return fmt.Errorf("insert activity: %w", err)
	}

	return nil
}
