-- Add quiz and dashboard support

-- Activity tracking table for dashboard recent activity
CREATE TABLE training_activities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activity_type VARCHAR(100) NOT NULL, -- module_started, module_completed, quiz_passed, quiz_failed, operation_blocked
    module_id UUID REFERENCES training_modules(id) ON DELETE SET NULL,
    module_name VARCHAR(255),
    score INTEGER, -- Quiz score percentage (0-100)
    metadata JSONB DEFAULT '{}'::jsonb, -- Additional activity details
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_training_activities_user_id ON training_activities(user_id);
CREATE INDEX idx_training_activities_created_at ON training_activities(created_at DESC);
CREATE INDEX idx_training_activities_type ON training_activities(activity_type);

-- Add quiz answers to user training progress
ALTER TABLE user_training_progress
ADD COLUMN quiz_answers JSONB DEFAULT NULL;

-- Add user preferences for onboarding and settings
ALTER TABLE users
ADD COLUMN preferences JSONB DEFAULT '{"has_completed_onboarding": false, "show_training_reminders": true}'::jsonb;

CREATE INDEX idx_users_preferences ON users USING GIN (preferences);

-- Add comment documentation
COMMENT ON TABLE training_activities IS 'Tracks user training activities for dashboard and analytics';
COMMENT ON COLUMN training_activities.activity_type IS 'Type of activity: module_started, module_completed, quiz_passed, quiz_failed, operation_blocked';
COMMENT ON COLUMN training_activities.score IS 'Quiz score percentage (0-100), NULL for non-quiz activities';
COMMENT ON COLUMN user_training_progress.quiz_answers IS 'Stores user quiz answers as JSONB array of {question_id, selected_answer}';
COMMENT ON COLUMN users.preferences IS 'User preferences including has_completed_onboarding, show_training_reminders, default_aws_region';
