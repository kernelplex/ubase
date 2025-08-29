-- +goose Up
-- +goose StatementBegin

ALTER TABLE users
	ADD COLUMN created_at TIMESTAMP;

ALTER TABLE users
	ADD COLUMN updated_at TIMESTAMP;

ALTER TABLE users
	ADD COLUMN last_login TIMESTAMP;

ALTER TABLE users
	ADD COLUMN login_count INT NOT NULL DEFAULT 0;

CREATE INDEX idx_users_last_login ON users(last_login);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE users
	DROP INDEX idx_users_last_login;

ALTER TABLE users
	DROP COLUMN created_at;

ALTER TABLE users
	DROP COLUMN updated_at;

ALTER TABLE users
	DROP COLUMN last_login;

ALTER TABLE users
	DROP COLUMN login_count;

-- +goose StatementEnd
