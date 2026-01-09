-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS repository_sync_audit (
    id TEXT PRIMARY KEY,
    repo_name TEXT NOT NULL,
    workspace_slug TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    successful_sync_time TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    success BOOLEAN NOT NULL,
    error_context TEXT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE IF EXISTS repository_sync_audit;
-- +goose StatementEnd
