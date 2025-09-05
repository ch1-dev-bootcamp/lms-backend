-- Drop lessons table
DROP TRIGGER IF EXISTS update_lessons_updated_at ON lessons;
DROP INDEX IF EXISTS idx_lessons_course_order;
DROP INDEX IF EXISTS idx_lessons_created_at;
DROP INDEX IF EXISTS idx_lessons_order_number;
DROP INDEX IF EXISTS idx_lessons_course_id;
DROP TABLE IF EXISTS lessons;
