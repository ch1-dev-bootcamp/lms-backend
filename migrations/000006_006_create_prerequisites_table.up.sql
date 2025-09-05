-- Create prerequisites table
CREATE TABLE prerequisites (
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    required_course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    PRIMARY KEY (course_id, required_course_id),
    -- Prevent self-referencing prerequisites
    CONSTRAINT check_no_self_prerequisite CHECK (course_id != required_course_id)
);

-- Create indexes for performance
CREATE INDEX idx_prerequisites_course_id ON prerequisites(course_id);
CREATE INDEX idx_prerequisites_required_course_id ON prerequisites(required_course_id);
