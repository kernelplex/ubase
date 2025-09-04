-- +goose Up
-- +goose StatementBegin
ALTER TABLE USERS
    ADD COLUMN verified BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE USERS
    DROP COLUMN verified;
-- +goose StatementEnd
