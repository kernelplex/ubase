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

CREATE TABLE user_roles (
	user_id INTEGER NOT NULL,
	role_id INTEGER NOT NULL,
	FOREIGN KEY(user_id) REFERENCES users(id),
	FOREIGN KEY(role_id) REFERENCES roles(id),
	CONSTRAINT user_roles_user_id_role_id_key UNIQUE (user_id, role_id)
);

CREATE TABLE role_permissions (
	role_id INTEGER NOT NULL,
	permission VARCHAR(255) NOT NULL,
	FOREIGN KEY(role_id) REFERENCES roles(id),
	CONSTRAINT role_permissions_role_id_fkey UNIQUE (role_id, permission)
);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE role_permissions;
DROP TABLE user_roles;
DROP TABLE users;
DROP TABLE roles;
DROP TABLE organizations;

-- +goose StatementEnd
