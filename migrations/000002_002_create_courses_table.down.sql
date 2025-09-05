-- Drop courses table
DROP TRIGGER IF EXISTS update_courses_updated_at ON courses;
DROP INDEX IF EXISTS idx_courses_title;
DROP INDEX IF EXISTS idx_courses_created_at;
DROP INDEX IF EXISTS idx_courses_status;
DROP INDEX IF EXISTS idx_courses_instructor_id;
DROP TABLE IF EXISTS courses;
