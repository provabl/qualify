package training

import "encoding/json"

// Module represents a training module
type Module struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Title            string          `json:"title"`
	Description      string          `json:"description,omitempty"`
	Category         string          `json:"category,omitempty"`
	Difficulty       string          `json:"difficulty,omitempty"`
	EstimatedMinutes int             `json:"estimated_minutes"`
	Content          json.RawMessage `json:"content,omitempty"`
	Status           string          `json:"status,omitempty"`
	Prerequisites    []string        `json:"prerequisites,omitempty"`
}

// PolicyDecision represents the result of policy evaluation
type PolicyDecision struct {
	Action          string   `json:"action"` // "allow" or "block"
	Reason          string   `json:"reason,omitempty"`
	RequiredModules []Module `json:"required_modules,omitempty"`
	Message         string   `json:"message"`
}

// Progress represents user training progress
type Progress struct {
	ModuleID    string `json:"module_id"`
	ModuleName  string `json:"module_name"`
	Status      string `json:"status"` // not_started, in_progress, completed
	CompletedAt string `json:"completed_at,omitempty"`
}

// QuizAnswer represents a single answer in a quiz submission
type QuizAnswer struct {
	QuestionID     string `json:"question_id"`
	SelectedAnswer int    `json:"selected_answer"`
}

// QuizSubmission represents answers submitted by a user
type QuizSubmission struct {
	UserID  string       `json:"user_id"`
	Answers []QuizAnswer `json:"answers"`
}

// QuizResult represents the result of a single quiz question
type QuizResult struct {
	QuestionID     string `json:"question_id"`
	Correct        bool   `json:"correct"`
	SelectedAnswer int    `json:"selected_answer"`
	CorrectAnswer  int    `json:"correct_answer"`
	Explanation    string `json:"explanation,omitempty"`
}

// QuizResponse represents the complete quiz evaluation
type QuizResponse struct {
	Score          int          `json:"score"`
	TotalQuestions int          `json:"total_questions"`
	CorrectAnswers int          `json:"correct_answers"`
	Results        []QuizResult `json:"results"`
}

// ActivityItem represents a training activity entry
type ActivityItem struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"` // module_started, module_completed, quiz_passed, quiz_failed, operation_blocked
	ModuleName string                 `json:"module_name,omitempty"`
	ModuleID   string                 `json:"module_id,omitempty"`
	Timestamp  string                 `json:"timestamp"`
	Score      *int                   `json:"score,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// TrainingSummary represents user training completion statistics
type TrainingSummary struct {
	TotalModules         int     `json:"total_modules"`
	Completed            int     `json:"completed"`
	InProgress           int     `json:"in_progress"`
	NotStarted           int     `json:"not_started"`
	CompletionPercentage int     `json:"completion_percentage"`
	AverageScore         *int    `json:"average_score,omitempty"`
}

// AvailableOps represents available and locked AWS operations
type AvailableOps struct {
	Unlocked []string `json:"unlocked"`
	Locked   []string `json:"locked"`
}

// DashboardStats represents user dashboard data
type DashboardStats struct {
	UserID           string          `json:"user_id"`
	TrainingSummary  TrainingSummary `json:"training_summary"`
	RecentActivity   []ActivityItem  `json:"recent_activity"`
	AvailableOps     AvailableOps    `json:"available_operations"`
}

// UserPreferences represents user preferences and settings
type UserPreferences struct {
	HasCompletedOnboarding bool   `json:"has_completed_onboarding"`
	ShowTrainingReminders  bool   `json:"show_training_reminders"`
	DefaultAWSRegion       string `json:"default_aws_region,omitempty"`
}

// UserProfile represents a user profile with preferences
type UserProfile struct {
	UserID      string          `json:"user_id"`
	Email       string          `json:"email"`
	Name        string          `json:"name"`
	Institution string          `json:"institution"`
	Role        string          `json:"role"` // researcher, admin, instructor
	Preferences UserPreferences `json:"preferences"`
	CreatedAt   string          `json:"created_at"`
}
