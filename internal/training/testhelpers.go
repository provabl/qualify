package training

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestHelper provides utilities for training service tests
type TestHelper struct {
	Mock sqlmock.Sqlmock
}

// NewTestHelper creates a new test helper
func NewTestHelper(mock sqlmock.Sqlmock) *TestHelper {
	return &TestHelper{Mock: mock}
}

// MockModuleQuery mocks a training module query
func (h *TestHelper) MockModuleQuery(moduleID string, content map[string]interface{}) {
	contentJSON, _ := json.Marshal(content)
	rows := sqlmock.NewRows([]string{"id", "name", "title", "description", "content", "required_score", "created_at", "updated_at"}).
		AddRow(moduleID, "test-module", "Test Module", "Test Description", contentJSON, 70, time.Now(), time.Now())

	h.Mock.ExpectQuery(`SELECT (.+) FROM training_modules WHERE id`).
		WithArgs(moduleID).
		WillReturnRows(rows)
}

// MockModuleNotFound mocks a module not found scenario
func (h *TestHelper) MockModuleNotFound(moduleID string) {
	h.Mock.ExpectQuery(`SELECT (.+) FROM training_modules WHERE id`).
		WithArgs(moduleID).
		WillReturnError(sqlmock.ErrCancelled)
}

// MockUserProgressQuery mocks a user progress query
func (h *TestHelper) MockUserProgressQuery(userID, moduleID string, status string, score *int) {
	if score != nil {
		rows := sqlmock.NewRows([]string{"user_id", "module_id", "status", "score", "started_at", "completed_at", "updated_at"}).
			AddRow(userID, moduleID, status, *score, time.Now(), time.Now(), time.Now())

		h.Mock.ExpectQuery(`SELECT (.+) FROM user_progress WHERE user_id`).
			WithArgs(userID, moduleID).
			WillReturnRows(rows)
	} else {
		rows := sqlmock.NewRows([]string{"user_id", "module_id", "status", "score", "started_at", "completed_at", "updated_at"}).
			AddRow(userID, moduleID, status, nil, time.Now(), nil, time.Now())

		h.Mock.ExpectQuery(`SELECT (.+) FROM user_progress WHERE user_id`).
			WithArgs(userID, moduleID).
			WillReturnRows(rows)
	}
}

// MockUserProgressNotFound mocks no user progress found
func (h *TestHelper) MockUserProgressNotFound(userID, moduleID string) {
	h.Mock.ExpectQuery(`SELECT (.+) FROM user_progress WHERE user_id`).
		WithArgs(userID, moduleID).
		WillReturnError(sqlmock.ErrCancelled)
}

// MockUserProgressUpdate mocks updating user progress
func (h *TestHelper) MockUserProgressUpdate(userID, moduleID string) {
	h.Mock.ExpectExec(`UPDATE user_progress SET`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), userID, moduleID).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

// MockUserProgressInsert mocks inserting new user progress
func (h *TestHelper) MockUserProgressInsert(userID, moduleID string) {
	h.Mock.ExpectExec(`INSERT INTO user_progress`).
		WithArgs(userID, moduleID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

// MockActivityInsert mocks inserting activity log
func (h *TestHelper) MockActivityInsert(userID string) {
	h.Mock.ExpectExec(`INSERT INTO activity_log`).
		WithArgs(userID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

// MockActivityQuery mocks querying activity log
func (h *TestHelper) MockActivityQuery(userID string, count int) {
	rows := sqlmock.NewRows([]string{"id", "activity_type", "module_id", "module_name", "score", "metadata", "created_at"})

	for i := 0; i < count; i++ {
		rows.AddRow(
			i+1,
			"module_completed",
			"mod-1",
			"Test Module",
			85,
			nil,
			time.Now(),
		)
	}

	h.Mock.ExpectQuery(`SELECT (.+) FROM activity_log WHERE user_id`).
		WithArgs(userID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)
}

// MockDashboardStatsQuery mocks dashboard statistics query
func (h *TestHelper) MockDashboardStatsQuery(userID string, total, completed, inProgress, notStarted int) {
	// Mock user progress summary
	summaryRows := sqlmock.NewRows([]string{"status", "count"}).
		AddRow("completed", completed).
		AddRow("in_progress", inProgress).
		AddRow("not_started", notStarted)

	h.Mock.ExpectQuery(`SELECT status, COUNT`).
		WithArgs(userID).
		WillReturnRows(summaryRows)

	// Mock average score
	avgRows := sqlmock.NewRows([]string{"avg"}).
		AddRow(85.5)

	h.Mock.ExpectQuery(`SELECT AVG`).
		WithArgs(userID).
		WillReturnRows(avgRows)

	// Mock total modules
	totalRows := sqlmock.NewRows([]string{"count"}).
		AddRow(total)

	h.Mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(totalRows)

	// Mock recent activity
	h.MockActivityQuery(userID, 5)

	// Mock unlocked operations
	unlockedRows := sqlmock.NewRows([]string{"operation"}).
		AddRow("s3:CreateBucket").
		AddRow("s3:ListBuckets")

	h.Mock.ExpectQuery(`SELECT operation FROM unlocked_operations WHERE user_id`).
		WithArgs(userID).
		WillReturnRows(unlockedRows)

	// Mock all required operations
	allOpsRows := sqlmock.NewRows([]string{"operation"}).
		AddRow("s3:CreateBucket").
		AddRow("s3:ListBuckets").
		AddRow("ec2:RunInstance").
		AddRow("iam:CreateUser")

	h.Mock.ExpectQuery(`SELECT operation FROM required_operations`).
		WillReturnRows(allOpsRows)
}

// CreateMockQuizContent creates mock quiz content for testing
func CreateMockQuizContent(questionCount int) map[string]interface{} {
	questions := make([]map[string]interface{}, questionCount)

	for i := 0; i < questionCount; i++ {
		questions[i] = map[string]interface{}{
			"id":            i,
			"question":      "Test question",
			"options":       []string{"A", "B", "C", "D"},
			"correctAnswer": 1,
			"explanation":   "Test explanation",
		}
	}

	return map[string]interface{}{
		"quiz": questions,
	}
}

// CreateMockQuizSubmission creates a mock quiz submission
func CreateMockQuizSubmission(userID string, questionCount int, correctCount int) QuizSubmission {
	answers := make([]QuizAnswer, questionCount)

	for i := 0; i < questionCount; i++ {
		answer := 0
		if i < correctCount {
			answer = 1 // Correct answer
		}

		answers[i] = QuizAnswer{
			QuestionID:     fmt.Sprintf("q%d", i+1),
			SelectedAnswer: answer,
		}
	}

	return QuizSubmission{
		UserID:  userID,
		Answers: answers,
	}
}
