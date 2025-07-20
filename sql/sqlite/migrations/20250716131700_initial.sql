-- +goose Up
-- +goose StatementBegin

CREATE TABLE organizations (
	id INTEGER PRIMARY KEY,
	name VARCHAR(255) NOT NULL,
	system_name VARCHAR(255) NOT NULL,
	status VARCHAR(255) NOT NULL,
	unique(name),
	unique(system_name)
);

CREATE TABLE roles (
	id INTEGER PRIMARY KEY,
	organization_id INTEGER NOT NULL,
	name VARCHAR(255) NOT NULL,
	system_name VARCHAR(255) NOT NULL,
	FOREIGN KEY(organization_id) REFERENCES organizations(id),
	unique(name),
	unique(system_name)
);

CREATE TABLE users (
	id INTEGER PRIMARY KEY,
	first_name VARCHAR(255) NOT NULL,
	last_name VARCHAR(255) NOT NULL,
	display_name VARCHAR(255) NOT NULL,
	email VARCHAR(255) NOT NULL,
	UNIQUE (email)
);

CREATE TABLE resource_types (
	id INTEGER PRIMARY KEY,
	name VARCHAR(255) NOT NULL,
	system_name VARCHAR(255) NOT NULL
);

CREATE TABLE user_roles (
	user_id INTEGER NOT NULL,
	organization_id INTEGER NOT NULL,
	role_id INTEGER NOT NULL,
	FOREIGN KEY(user_id) REFERENCES users(id),
	FOREIGN KEY(organization_id) REFERENCES organizations(id),
	FOREIGN KEY(role_id) REFERENCES roles(id)
);

CREATE TABLE permissions (
	id INTEGER PRIMARY KEY,
	system_name VARCHAR(255) NOT NULL,
	unique(system_name)
);

CREATE TABLE role_permissions (
	role_id INTEGER NOT NULL,
	permission_id INTEGER NOT NULL,
	FOREIGN KEY(role_id) REFERENCES roles(id),
	FOREIGN KEY(permission_id) REFERENCES permissions(id)
);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE role_permissions;
DROP TABLE permissions;
DROP TABLE user_roles;
DROP TABLE roles;
DROP TABLE resource_types;
DROP TABLE organizations;

-- +goose StatementEnd
