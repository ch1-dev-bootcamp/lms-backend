-- Drop progress table
DROP INDEX IF EXISTS idx_progress_completed_at;
DROP INDEX IF EXISTS idx_progress_lesson_id;
DROP INDEX IF EXISTS idx_progress_user_id;
DROP TABLE IF EXISTS progress;
