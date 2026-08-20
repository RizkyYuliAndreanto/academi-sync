CREATE TABLE documents (
    id UUID PRIMARY KEY,
    guidance_session_id UUID NOT NULL REFERENCES guidance_sessions(id) ON DELETE CASCADE,
    uploaded_by UUID NOT NULL REFERENCES users(id),
    original_filename VARCHAR(255) NOT NULL CHECK (length(trim(original_filename)) > 0),
    storage_bucket VARCHAR(100) NOT NULL CHECK (length(trim(storage_bucket)) > 0),
    storage_key VARCHAR(255) NOT NULL CHECK (length(trim(storage_key)) > 0),
    mime_type VARCHAR(100) NOT NULL CHECK (length(trim(mime_type)) > 0),
    size_bytes BIGINT NOT NULL CHECK (size_bytes > 0),
    checksum_sha256 VARCHAR(64) NOT NULL CHECK (length(checksum_sha256) = 64),
    page_count INT CHECK (page_count IS NULL OR page_count > 0),
    version INT NOT NULL DEFAULT 1 CHECK (version >= 1),
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'archived', 'deleted')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_storage_location UNIQUE (storage_bucket, storage_key)
);

CREATE INDEX idx_documents_guidance_session ON documents (guidance_session_id);
