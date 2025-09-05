-- +goose Up
-- +goose StatementBegin

CREATE TABLE user_api_keys (
    id VARCHAR(10) NOT NULL PRIMARY KEY,
    secret_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    user_id BIGINT NOT NULL,
    organization_id BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE user_api_keys;
-- +goose StatementEnd
