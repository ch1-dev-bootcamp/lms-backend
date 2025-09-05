-- Drop enrollments table
DROP INDEX IF EXISTS idx_enrollments_enrolled_at;
DROP INDEX IF EXISTS idx_enrollments_course_id;
DROP INDEX IF EXISTS idx_enrollments_user_id;
DROP TABLE IF EXISTS enrollments;
