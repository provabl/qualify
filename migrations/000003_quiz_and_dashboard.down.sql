-- Rollback quiz and dashboard support

-- Remove user preferences
DROP INDEX IF EXISTS idx_users_preferences;
ALTER TABLE users DROP COLUMN IF EXISTS preferences;

-- Remove quiz answers from progress tracking
ALTER TABLE user_training_progress DROP COLUMN IF EXISTS quiz_answers;

-- Remove activity tracking table
DROP INDEX IF EXISTS idx_training_activities_type;
DROP INDEX IF EXISTS idx_training_activities_created_at;
DROP INDEX IF EXISTS idx_training_activities_user_id;
DROP TABLE IF EXISTS training_activities;
