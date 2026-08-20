CREATE TABLE guidance_sessions (
    id UUID PRIMARY KEY,
    lecturer_id UUID NOT NULL REFERENCES users(id),
    student_id UUID NOT NULL REFERENCES users(id),
    topic VARCHAR(255) NOT NULL CHECK (length(trim(topic)) > 0),
    description TEXT NOT NULL DEFAULT '',
    scheduled_start_at TIMESTAMPTZ NOT NULL,
    scheduled_end_at TIMESTAMPTZ NOT NULL,
    actual_start_at TIMESTAMPTZ,
    actual_end_at TIMESTAMPTZ,
    status VARCHAR(50) NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'ongoing', 'completed', 'cancelled')),
    created_by UUID NOT NULL REFERENCES users(id),
    version INT NOT NULL DEFAULT 1 CHECK (version >= 1),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_lecturer_student_different CHECK (lecturer_id != student_id),
    CONSTRAINT check_scheduled_times CHECK (scheduled_end_at > scheduled_start_at)
);

CREATE INDEX idx_guidance_sessions_lecturer ON guidance_sessions (lecturer_id, scheduled_start_at);
CREATE INDEX idx_guidance_sessions_student ON guidance_sessions (student_id, scheduled_start_at);
