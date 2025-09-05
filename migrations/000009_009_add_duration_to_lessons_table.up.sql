-- Add duration column to lessons table
ALTER TABLE lessons ADD COLUMN duration INTEGER NOT NULL DEFAULT 30;

-- Add comment to explain the column
COMMENT ON COLUMN lessons.duration IS 'Duration of the lesson in minutes';
