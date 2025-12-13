package training

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/scttfrdmn/ark/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_SubmitQuiz(t *testing.T) {
	tests := []struct {
		name          string
		moduleID      string
		submission    QuizSubmission
		moduleContent map[string]interface{}
		wantScore     int
		wantCorrect   int
		wantTotal     int
		wantErr       bool
	}{
		{
			name:     "all correct answers",
			moduleID: "module-1",
			submission: QuizSubmission{
				UserID: "user-1",
				Answers: []QuizAnswer{
					{QuestionID: "q1", SelectedAnswer: 0},
					{QuestionID: "q2", SelectedAnswer: 1},
				},
			},
			moduleContent: map[string]interface{}{
				"sections": []map[string]interface{}{
					{
						"type": "quiz",
						"questions": []map[string]interface{}{
							{
								"id":             "q1",
								"question":       "What is 1+1?",
								"options":        []string{"2", "3", "4"},
								"correct_answer": 0,
								"explanation":    "Basic math",
							},
							{
								"id":             "q2",
								"question":       "What is 2+2?",
								"options":        []string{"3", "4", "5"},
								"correct_answer": 1,
								"explanation":    "More math",
							},
						},
					},
				},
			},
			wantScore:   100,
			wantCorrect: 2,
			wantTotal:   2,
			wantErr:     false,
		},
		{
			name:     "partial correct answers",
			moduleID: "module-1",
			submission: QuizSubmission{
				UserID: "user-1",
				Answers: []QuizAnswer{
					{QuestionID: "q1", SelectedAnswer: 0},
					{QuestionID: "q2", SelectedAnswer: 0}, // Wrong
				},
			},
			moduleContent: map[string]interface{}{
				"sections": []map[string]interface{}{
					{
						"type": "quiz",
						"questions": []map[string]interface{}{
							{
								"id":             "q1",
								"question":       "What is 1+1?",
								"options":        []string{"2", "3", "4"},
								"correct_answer": 0,
								"explanation":    "Basic math",
							},
							{
								"id":             "q2",
								"question":       "What is 2+2?",
								"options":        []string{"3", "4", "5"},
								"correct_answer": 1,
								"explanation":    "More math",
							},
						},
					},
				},
			},
			wantScore:   50,
			wantCorrect: 1,
			wantTotal:   2,
			wantErr:     false,
		},
		{
			name:     "all wrong answers",
			moduleID: "module-1",
			submission: QuizSubmission{
				UserID: "user-1",
				Answers: []QuizAnswer{
					{QuestionID: "q1", SelectedAnswer: 1}, // Wrong
					{QuestionID: "q2", SelectedAnswer: 0}, // Wrong
				},
			},
			moduleContent: map[string]interface{}{
				"sections": []map[string]interface{}{
					{
						"type": "quiz",
						"questions": []map[string]interface{}{
							{
								"id":             "q1",
								"question":       "What is 1+1?",
								"options":        []string{"2", "3", "4"},
								"correct_answer": 0,
								"explanation":    "Basic math",
							},
							{
								"id":             "q2",
								"question":       "What is 2+2?",
								"options":        []string{"3", "4", "5"},
								"correct_answer": 1,
								"explanation":    "More math",
							},
						},
					},
				},
			},
			wantScore:   0,
			wantCorrect: 0,
			wantTotal:   2,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sqlDB, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer sqlDB.Close()

			db := &database.DB{DB: sqlDB}
			svc := NewService(db)

			// Mock GetModule
			contentJSON, _ := json.Marshal(tt.moduleContent)
			rows := sqlmock.NewRows([]string{"id", "name", "title", "description", "category", "difficulty", "estimated_minutes", "content", "status", "prerequisites"}).
				AddRow(tt.moduleID, "test-module", "Test Module", "Description", "test", "beginner", 10, contentJSON, "active", nil)
			mock.ExpectQuery("SELECT (.+) FROM training_modules WHERE").
				WithArgs(tt.moduleID).
				WillReturnRows(rows)

			// Mock quiz answers update
			answersJSON, _ := json.Marshal(tt.submission.Answers)
			mock.ExpectExec("UPDATE user_training_progress SET quiz_answers").
				WithArgs(answersJSON, tt.submission.UserID, tt.moduleID).
				WillReturnResult(sqlmock.NewResult(1, 1))

			// Mock activity recording (module name lookup)
			mock.ExpectQuery("SELECT name FROM training_modules WHERE id").
				WithArgs(tt.moduleID).
				WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("test-module"))

			// Mock activity insert
			mock.ExpectExec("INSERT INTO training_activities").
				WillReturnResult(sqlmock.NewResult(1, 1))

			// Execute
			result, err := svc.SubmitQuiz(context.Background(), tt.moduleID, tt.submission)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantScore, result.Score)
			assert.Equal(t, tt.wantCorrect, result.CorrectAnswers)
			assert.Equal(t, tt.wantTotal, result.TotalQuestions)
			assert.Len(t, result.Results, tt.wantTotal)

			// Verify all expectations met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestService_GetDashboardStats(t *testing.T) {
	tests := []struct {
		name                 string
		userID               string
		totalModules         int
		completed            int
		inProgress           int
		avgScore             sql.NullFloat64
		activityCount        int
		s3ModulesCompleted   int
		wantCompletionPct    int
		wantAvgScore         *int
		wantUnlockedOpsCount int
	}{
		{
			name:                 "no progress",
			userID:               "user-1",
			totalModules:         4,
			completed:            0,
			inProgress:           0,
			avgScore:             sql.NullFloat64{Valid: false},
			activityCount:        0,
			s3ModulesCompleted:   0,
			wantCompletionPct:    0,
			wantAvgScore:         nil,
			wantUnlockedOpsCount: 0,
		},
		{
			name:                 "50% complete with avg score",
			userID:               "user-1",
			totalModules:         4,
			completed:            2,
			inProgress:           1,
			avgScore:             sql.NullFloat64{Float64: 85.5, Valid: true},
			activityCount:        5,
			s3ModulesCompleted:   1,
			wantCompletionPct:    50,
			wantAvgScore:         intPtr(85),
			wantUnlockedOpsCount: 2,
		},
		{
			name:                 "100% complete",
			userID:               "user-1",
			totalModules:         4,
			completed:            4,
			inProgress:           0,
			avgScore:             sql.NullFloat64{Float64: 92.0, Valid: true},
			activityCount:        10,
			s3ModulesCompleted:   2,
			wantCompletionPct:    100,
			wantAvgScore:         intPtr(92),
			wantUnlockedOpsCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sqlDB, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer sqlDB.Close()

			db := &database.DB{DB: sqlDB}
			svc := NewService(db)

			// Mock training summary query
			notStarted := tt.totalModules - tt.completed - tt.inProgress
			rows := sqlmock.NewRows([]string{"total_modules", "completed", "in_progress", "not_started", "avg_score"}).
				AddRow(tt.totalModules, tt.completed, tt.inProgress, notStarted, tt.avgScore)
			mock.ExpectQuery("SELECT (.+) FROM training_modules tm LEFT JOIN user_training_progress").
				WithArgs(tt.userID).
				WillReturnRows(rows)

			// Mock recent activity query
			activityRows := sqlmock.NewRows([]string{"id", "activity_type", "module_id", "module_name", "score", "metadata", "created_at"})
			for i := 0; i < tt.activityCount; i++ {
				activityRows.AddRow(
					"activity-1",
					"module_completed",
					"module-1",
					"Test Module",
					sql.NullInt64{Int64: 90, Valid: true},
					[]byte("{}"),
					time.Now(),
				)
			}
			mock.ExpectQuery("SELECT (.+) FROM training_activities WHERE user_id").
				WithArgs(tt.userID, 10, 0).
				WillReturnRows(activityRows)

			// Mock S3 modules completed query
			s3Rows := sqlmock.NewRows([]string{"count"}).AddRow(tt.s3ModulesCompleted)
			mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM training_modules tm JOIN user_training_progress").
				WithArgs(tt.userID).
				WillReturnRows(s3Rows)

			// Execute
			stats, err := svc.GetDashboardStats(context.Background(), tt.userID)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.userID, stats.UserID)
			assert.Equal(t, tt.totalModules, stats.TrainingSummary.TotalModules)
			assert.Equal(t, tt.completed, stats.TrainingSummary.Completed)
			assert.Equal(t, tt.inProgress, stats.TrainingSummary.InProgress)
			assert.Equal(t, tt.wantCompletionPct, stats.TrainingSummary.CompletionPercentage)

			if tt.wantAvgScore == nil {
				assert.Nil(t, stats.TrainingSummary.AverageScore)
			} else {
				require.NotNil(t, stats.TrainingSummary.AverageScore)
				assert.Equal(t, *tt.wantAvgScore, *stats.TrainingSummary.AverageScore)
			}

			assert.Len(t, stats.RecentActivity, tt.activityCount)
			assert.Equal(t, tt.wantUnlockedOpsCount, len(stats.AvailableOps.Unlocked))

			// Verify all expectations met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestService_RecordActivity(t *testing.T) {
	tests := []struct {
		name         string
		userID       string
		activityType string
		moduleID     string
		metadata     map[string]interface{}
		moduleName   string
		wantErr      bool
	}{
		{
			name:         "module_completed with score",
			userID:       "user-1",
			activityType: "module_completed",
			moduleID:     "module-1",
			metadata:     map[string]interface{}{"score": 85},
			moduleName:   "Test Module",
			wantErr:      false,
		},
		{
			name:         "quiz_passed",
			userID:       "user-1",
			activityType: "quiz_passed",
			moduleID:     "module-1",
			metadata:     map[string]interface{}{"score": 90},
			moduleName:   "Test Module",
			wantErr:      false,
		},
		{
			name:         "operation_blocked",
			userID:       "user-1",
			activityType: "operation_blocked",
			moduleID:     "",
			metadata:     map[string]interface{}{"operation": "s3:CreateBucket"},
			moduleName:   "",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sqlDB, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer sqlDB.Close()

			db := &database.DB{DB: sqlDB}
			svc := NewService(db)

			// Mock module name lookup if moduleID provided
			if tt.moduleID != "" {
				rows := sqlmock.NewRows([]string{"name"}).AddRow(tt.moduleName)
				mock.ExpectQuery("SELECT name FROM training_modules WHERE id").
					WithArgs(tt.moduleID).
					WillReturnRows(rows)
			}

			// Mock activity insert
			mock.ExpectExec("INSERT INTO training_activities").
				WillReturnResult(sqlmock.NewResult(1, 1))

			// Execute
			err = svc.RecordActivity(context.Background(), tt.userID, tt.activityType, tt.moduleID, tt.metadata)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestService_GetUserActivity(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	db := &database.DB{DB: sqlDB}
	svc := NewService(db)

	// Setup mock data
	rows := sqlmock.NewRows([]string{"id", "activity_type", "module_id", "module_name", "score", "metadata", "created_at"}).
		AddRow("act-1", "module_completed", "mod-1", "S3 Basics", sql.NullInt64{Int64: 85, Valid: true}, []byte("{}"), time.Now()).
		AddRow("act-2", "quiz_passed", "mod-1", "S3 Basics", sql.NullInt64{Int64: 90, Valid: true}, []byte("{}"), time.Now().Add(-1*time.Hour))

	mock.ExpectQuery("SELECT (.+) FROM training_activities WHERE user_id").
		WithArgs("user-1", 10, 0).
		WillReturnRows(rows)

	// Execute
	activities, err := svc.GetUserActivity(context.Background(), "user-1", 10, 0)

	// Assert
	require.NoError(t, err)
	assert.Len(t, activities, 2)
	assert.Equal(t, "module_completed", activities[0].Type)
	assert.Equal(t, "quiz_passed", activities[1].Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Helper function
func intPtr(i int) *int {
	return &i
}
