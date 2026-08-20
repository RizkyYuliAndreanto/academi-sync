CREATE TABLE annotation_snapshots (
    id UUID PRIMARY KEY,
    guidance_session_id UUID NOT NULL REFERENCES guidance_sessions(id) ON DELETE CASCADE,
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    page_number INT NOT NULL CHECK (page_number > 0),
    room_sequence BIGINT NOT NULL CHECK (room_sequence >= 0),
    state_version INT NOT NULL CHECK (state_version >= 0),
    snapshot JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_doc_page_state_version UNIQUE (document_id, page_number, state_version)
);

CREATE INDEX idx_annotation_snapshots_session ON annotation_snapshots (guidance_session_id);
