-- +goose Up
-- +goose StatementBegin


CREATE TABLE organizations (
	id BIGINT PRIMARY KEY,
	name VARCHAR(255) NOT NULL,
	system_name VARCHAR(255) NOT NULL,
	status VARCHAR(255) NOT NULL,
	CONSTRAINT organizations_name_key UNIQUE (name),
	CONSTRAINT organizations_system_name_key UNIQUE (system_name)
);

CREATE TABLE roles (
	id BIGINT PRIMARY KEY,
	organization_id BIGINT NOT NULL,
	name VARCHAR(255) NOT NULL,
	system_name VARCHAR(255) NOT NULL,
	CONSTRAINT roles_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES organizations(id),
	CONSTRAINT roles_name_key UNIQUE (name),
	CONSTRAINT roles_system_name_key UNIQUE (system_name)
);

CREATE TABLE users (
	id BIGINT PRIMARY KEY,
	first_name VARCHAR(255) NOT NULL,
	last_name VARCHAR(255) NOT NULL,
	display_name VARCHAR(255) NOT NULL,
	email VARCHAR(255) NOT NULL,
	CONSTRAINT users_email_key UNIQUE (email)
);

CREATE TABLE resource_types (
	id BIGINT PRIMARY KEY,
	name VARCHAR(255) NOT NULL,
	system_name VARCHAR(255) NOT NULL
);

CREATE TABLE user_roles (
	user_id BIGINT NOT NULL,
	role_id BIGINT NOT NULL,
	CONSTRAINT user_roles_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id),
	CONSTRAINT user_roles_role_id_fkey FOREIGN KEY (role_id) REFERENCES roles(id),
	CONSTRAINT user_roles_user_id_role_id_key UNIQUE (user_id, role_id)
	
);

CREATE TABLE role_permissions (
	role_id BIGINT NOT NULL,
	permission VARCHAR(255) NOT NULL,
	CONSTRAINT role_permissions_role_id_fkey FOREIGN KEY (role_id) REFERENCES roles(id),
	CONSTRAINT role_permissions_role_id_fkey UNIQUE (role_id, permission)
);

-- Create indexes for foreign keys
CREATE INDEX idx_roles_organization_id ON roles(organization_id);
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_organization_id ON user_roles(organization_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE user_roles;
DROP TABLE roles;
DROP TABLE users;
DROP TABLE resource_types;
DROP TABLE organizations;

-- +goose StatementEnd
