CREATE TABLE annotations (
    id UUID PRIMARY KEY,
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    document_version INT NOT NULL CHECK (document_version > 0),
    page_number INT NOT NULL CHECK (page_number > 0),
    created_by UUID NOT NULL REFERENCES users(id),
    tool VARCHAR(50) NOT NULL CHECK (length(trim(tool)) > 0),
    data JSONB NOT NULL,
    revision INT NOT NULL DEFAULT 1 CHECK (revision > 0),
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_annotations_doc_page ON annotations (document_id, page_number) WHERE deleted_at IS NULL;
