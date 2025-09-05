-- Create course_completions table
CREATE TABLE course_completions (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    completed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completion_rate DECIMAL(5,2) NOT NULL DEFAULT 100.00,
    PRIMARY KEY (user_id, course_id)
);

-- Create indexes for performance
CREATE INDEX idx_course_completions_user_id ON course_completions(user_id);
CREATE INDEX idx_course_completions_course_id ON course_completions(course_id);
CREATE INDEX idx_course_completions_completed_at ON course_completions(completed_at);

-- Create trigger to update updated_at timestamp
CREATE TRIGGER update_course_completions_updated_at 
    BEFORE UPDATE ON course_completions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
