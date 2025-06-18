-- +goose Up
-- +goose StatementBegin

CREATE TABLE "users" (
	"user_id" BIGINT PRIMARY KEY,
	"last_name" VARCHAR(255) NOT NULL,
	"first_name" VARCHAR(255) NOT NULL,
	"display_name" VARCHAR(255) NOT NULL,
	"email" VARCHAR(255) NOT NULL,
	CONSTRAINT users_email_unique UNIQUE ("email")
);

CREATE INDEX idx_users_email ON users (email);

CREATE table "roles" (
	"role_id" BIGINT PRIMARY KEY,
	"name" VARCHAR(255) NOT NULL
);

CREATE TABLE "user_roles" (
	"user_id" BIGINT NOT NULL,
	"role_id" BIGINT NOT NULL,
	PRIMARY KEY ("user_id", "role_id"),
	FOREIGN KEY ("user_id") REFERENCES "users" ("user_id"),
	FOREIGN KEY ("role_id") REFERENCES "roles" ("role_id")
);

CREATE TABLE "permissions" (
	"permission_id" BIGINT PRIMARY KEY,
	"name" VARCHAR(255) NOT NULL
);

CREATE TABLE "role_permissions" (
	"role_id" BIGINT NOT NULL,
	"permission_id" BIGINT NOT NULL,
	PRIMARY KEY ("role_id", "permission_id"),
	FOREIGN KEY ("role_id") REFERENCES "roles" ("role_id"),
	FOREIGN KEY ("permission_id") REFERENCES "permissions" ("permission_id")
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS "role_permissions";
DROP TABLE IF EXISTS "permissions";
DROP TABLE IF EXISTS "user_roles";
DROP TABLE IF EXISTS "roles";
DROP TABLE IF EXISTS "users";

-- +goose StatementEnd
