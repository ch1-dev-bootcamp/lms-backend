-- Drop certificates table
DROP INDEX IF EXISTS idx_certificates_user_course;
DROP INDEX IF EXISTS idx_certificates_certificate_code;
DROP INDEX IF EXISTS idx_certificates_issued_at;
DROP INDEX IF EXISTS idx_certificates_course_id;
DROP INDEX IF EXISTS idx_certificates_user_id;
DROP TABLE IF EXISTS certificates;
