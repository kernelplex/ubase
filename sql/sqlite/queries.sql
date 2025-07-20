-- ---------------------------------------------------------------------------
--
-- Organization Management
--
-- ---------------------------------------------------------------------------

-- name: AddOrganization :exec
INSERT INTO organizations (id, name, system_name, status) 
VALUES (sqlc.arg(id), sqlc.arg(name), sqlc.arg(system_name), sqlc.arg(status));

-- name: UpdateOrganization :exec
UPDATE organizations SET 
name = sqlc.arg(name), system_name = sqlc.arg(system_name), status = sqlc.arg(status)
WHERE id = sqlc.arg(id);

-- name: GetOrganization :one
SELECT id, name, system_name, status FROM organizations WHERE id = sqlc.arg(id);

-- name: GetOrganizationBySystemName :one
SELECT id, name, system_name, status FROM organizations WHERE system_name = sqlc.arg(system_name);

-- name: GetOrganizations :many
SELECT id, name, system_name FROM organizations;

-- ---------------------------------------------------------------------------
--
-- Permission Management
--
-- ---------------------------------------------------------------------------

-- name: AddPermission :exec
INSERT INTO permissions (id, system_name) 
VALUES (sqlc.arg(id), sqlc.arg(system_name));

-- name: GetPermissions :many
SELECT id, system_name FROM permissions;

-- ---------------------------------------------------------------------------
--
-- Role Management
--
-- ---------------------------------------------------------------------------

-- name: AddRole :exec
INSERT INTO roles (id, organization_id, name, system_name) 
VALUES (sqlc.arg(id), sqlc.arg(organization_id), sqlc.arg(name), sqlc.arg(system_name));

-- name: UpdateRole :exec
UPDATE roles SET 
name = sqlc.arg(name), system_name = sqlc.arg(system_name) WHERE id = sqlc.arg(id);

-- name: DeleteRole :exec
DELETE FROM roles WHERE id = sqlc.arg(id);

-- name: GetOrganizationRoles :many
SELECT r.id, r.name, r.system_name FROM roles r
WHERE r.organization_id = sqlc.arg(organization_id);

-- name: AddUserToRole :exec
INSERT INTO user_roles (user_id, role_id) 
VALUES (sqlc.arg(user_id), sqlc.arg(role_id));

-- name: RemoveUserFromRole :exec
DELETE FROM user_roles WHERE user_id = sqlc.arg(user_id) AND role_id = sqlc.arg(role_id);

-- name: RemoveAllRolesFromUser :exec
DELETE FROM user_roles WHERE user_id = sqlc.arg(user_id);

-- name: AddRoleToUser :exec
INSERT INTO user_roles (user_id, role_id) 
VALUES (sqlc.arg(user_id), sqlc.arg(role_id));

-- name: GetUserOrganizationRoles :many
SELECT r.id, r.name, r.system_name FROM user_roles ur
LEFT JOIN roles r ON r.id = ur.role_id
WHERE ur.user_id = sqlc.arg(user_id) AND r.organization_id = sqlc.arg(organization_id);

-- name: AddPermissionToRole :exec
INSERT INTO role_permissions (role_id, permission_id) 
VALUES (sqlc.arg(role_id), sqlc.arg(permission_id));

-- name: RemovePermissionFromRole :exec
DELETE FROM role_permissions 
WHERE role_id = sqlc.arg(role_id) AND permission_id = sqlc.arg(permission_id);

-- name: GetRolePermissions :many
SELECT p.id, p.system_name FROM role_permissions rp
LEFT JOIN permissions p ON p.id = rp.permission_id
WHERE rp.role_id = sqlc.arg(role_id);

-- name: GetUserOrganizations :many
SELECT o.id, o.name, o.system_name 
FROM user_roles ur
LEFT JOIN roles r ON r.id = ur.role_id
LEFT JOIN organizations o ON o.id = r.organization_id
WHERE ur.user_id = sqlc.arg(user_id);

-- name: GetUserOrganizationPermissions :many
SELECT p.id, p.system_name FROM user_roles ur
LEFT JOIN role_permissions rp ON rp.role_id = ur.role_id
LEFT JOIN permissions p ON p.id = rp.permission_id
WHERE ur.user_id = sqlc.arg(user_id) AND ur.organization_id = sqlc.arg(organization_id);

-- ---------------------------------------------------------------------------
--
-- User Management
--
-- ---------------------------------------------------------------------------

-- name: AddUser :exec
INSERT INTO users (id, first_name, last_name, display_name, email) 
VALUES (sqlc.arg(id), sqlc.arg(first_name), sqlc.arg(last_name), sqlc.arg(display_name), sqlc.arg(email));

-- name: GetUser :one
SELECT id, first_name, last_name, display_name, email FROM users WHERE id = sqlc.arg(id);

-- name: GetUserByEmail :one
SELECT id, first_name, last_name, display_name, email FROM users WHERE email = sqlc.arg(email);

-- name: UpdateUser :exec
UPDATE users SET first_name = sqlc.arg(first_name), last_name = sqlc.arg(last_name), display_name = sqlc.arg(display_name), email = sqlc.arg(email) WHERE id = sqlc.arg(id);


