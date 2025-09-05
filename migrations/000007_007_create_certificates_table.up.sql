-- Create certificates table
CREATE TABLE certificates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    issued_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    certificate_code VARCHAR(255) UNIQUE NOT NULL
);

-- Create indexes for performance
CREATE INDEX idx_certificates_user_id ON certificates(user_id);
CREATE INDEX idx_certificates_course_id ON certificates(course_id);
CREATE INDEX idx_certificates_issued_at ON certificates(issued_at);
CREATE INDEX idx_certificates_certificate_code ON certificates(certificate_code);

-- Create unique constraint for user_id and course_id (one certificate per user per course)
CREATE UNIQUE INDEX idx_certificates_user_course ON certificates(user_id, course_id);
