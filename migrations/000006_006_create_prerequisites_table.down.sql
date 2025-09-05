-- Drop prerequisites table
DROP INDEX IF EXISTS idx_prerequisites_required_course_id;
DROP INDEX IF EXISTS idx_prerequisites_course_id;
DROP TABLE IF EXISTS prerequisites;
