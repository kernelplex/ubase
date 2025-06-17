-- +goose Up
-- +goose StatementBegin

CREATE TABLE "users" (
	"user_id" INTEGER NOT NULL PRIMARY KEY,
	"last_name" VARCHAR(255) NOT NULL,
	"first_name" VARCHAR(255) NOT NULL,
	"display_name" VARCHAR(255) NOT NULL,
	"email" VARCHAR(255) NOT NULL,
	UNIQUE ("email")
);

CREATE table "roles" (
	"role_id" INTEGER NOT NULL PRIMARY KEY,
	"name" VARCHAR(255) NOT NULL
);

CREATE TABLE "user_roles" (
	"user_id" INTEGER NOT NULL,
	"role_id" INTEGER NOT NULL,
	PRIMARY KEY ("user_id", "role_id"),
	FOREIGN KEY ("user_id") REFERENCES "users" ("user_id"),
	FOREIGN KEY ("role_id") REFERENCES "roles" ("role_id")
);

CREATE TABLE "permissions" (
	"permission_id" INTEGER NOT NULL PRIMARY KEY,
	"name" VARCHAR(255) NOT NULL
);

CREATE TABLE "role_permissions" (
	"role_id" INTEGER NOT NULL,
	"permission_id" INTEGER NOT NULL,
	PRIMARY KEY ("role_id", "permission_id"),
	FOREIGN KEY ("role_id") REFERENCES "roles" ("role_id"),
	FOREIGN KEY ("permission_id") REFERENCES "permissions" ("permission_id")
);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE "role_permissions";
DROP TABLE "permissions";
DROP TABLE "user_roles";
DROP TABLE "roles";

DROP TABLE "users";
-- +goose StatementEnd
