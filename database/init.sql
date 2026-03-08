CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Users Table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    email VARCHAR(255) NOT NULL,
    CONSTRAINT uni_users_email UNIQUE (email), -- Tên tường minh cho GORM
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Documents Table
CREATE TABLE IF NOT EXISTS documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    original_filename VARCHAR(255) NOT NULL,
    storage_bucket VARCHAR(100) NOT NULL,
    storage_path VARCHAR(512) NOT NULL,
    mime_type VARCHAR(50),
    file_size BIGINT,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_documents_user_id ON documents (user_id);

-- 3. Jobs Table
CREATE TABLE IF NOT EXISTS jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    -- Bổ sung user_id để khớp với Struct Go và dễ truy vấn
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    document_id UUID NOT NULL REFERENCES documents (id) ON DELETE CASCADE,
    state VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    result JSONB,
    retry_count INT DEFAULT 0,
    error_message TEXT,
    trace_id VARCHAR(100),
    started_at TIMESTAMP WITH TIME ZONE,
    finished_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_jobs_user_id ON jobs (user_id);

CREATE INDEX idx_jobs_state ON jobs (state);

CREATE INDEX idx_jobs_document_id ON jobs (document_id);