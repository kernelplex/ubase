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

-- name: ListOrganizations :many
SELECT id, name, system_name, status FROM organizations;

-- name: GetOrganization :one
SELECT id, name, system_name, status FROM organizations WHERE id = sqlc.arg(id);

-- name: GetOrganizationBySystemName :one
SELECT id, name, system_name, status FROM organizations WHERE system_name = sqlc.arg(system_name);

-- name: GetOrganizations :many
SELECT id, name, system_name FROM organizations;

-- name: GetUsersInRole :many
SELECT u.id, u.first_name, u.last_name, u.display_name, u.email, u.verified
FROM user_roles ur
JOIN users u ON u.id = ur.user_id
WHERE ur.role_id = sqlc.arg(role_id);

-- name: GetRolesForUser :many
SELECT r.id, r.name, r.system_name
FROM user_roles ur
JOIN roles r ON r.id = ur.role_id
WHERE ur.user_id = sqlc.arg(user_id);


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
JOIN roles r ON r.id = ur.role_id
WHERE ur.user_id = sqlc.arg(user_id) AND r.organization_id = sqlc.arg(organization_id);

-- name: GetAllUserOrganizationRoles :many
SELECT r.id, r.name, r.system_name, o.id as organization_id, o.name as organization_name, o.system_name as organization_system_name
FROM user_roles ur
JOIN roles r ON r.id = ur.role_id
JOIN organizations o ON o.id = r.organization_id
WHERE ur.user_id = sqlc.arg(user_id);

-- name: AddPermissionToRole :exec
INSERT INTO role_permissions (role_id, permission) 
VALUES (sqlc.arg(role_id), sqlc.arg(permission));

-- name: RemovePermissionFromRole :exec
DELETE FROM role_permissions 
WHERE role_id = sqlc.arg(role_id) AND permission = sqlc.arg(permission);

-- name: GetRolePermissions :many
SELECT rp.permission FROM role_permissions rp
WHERE rp.role_id = sqlc.arg(role_id);

-- name: GetUserOrganizations :many
SELECT o.id, o.name, o.system_name 
FROM user_roles ur
LEFT JOIN roles r ON r.id = ur.role_id
LEFT JOIN organizations o ON o.id = r.organization_id
WHERE ur.user_id = sqlc.arg(user_id);

-- name: GetUserOrganizationPermissions :many
SELECT rp.permission FROM user_roles ur
JOIN roles r ON r.id = ur.role_id
JOIN role_permissions rp ON rp.role_id = ur.role_id
WHERE ur.user_id = sqlc.arg(user_id) AND r.organization_id = sqlc.arg(organization_id);

-- ---------------------------------------------------------------------------
--
-- User Management
--
-- ---------------------------------------------------------------------------

-- name: AddUser :exec
INSERT INTO users (id, first_name, last_name, display_name, email, verified, created_at, updated_at) 
VALUES (sqlc.arg(id), sqlc.arg(first_name), sqlc.arg(last_name), sqlc.arg(display_name), sqlc.arg(email), sqlc.arg(verified), sqlc.arg(created_at), sqlc.arg(updated_at));


-- name: GetUser :one
SELECT id, first_name, last_name, display_name, email, verified FROM users WHERE id = sqlc.arg(id);

-- name: GetUserByEmail :one
SELECT id, first_name, last_name, display_name, email, verified FROM users WHERE email = sqlc.arg(email);

-- name: UpdateUser :exec
UPDATE users SET first_name = sqlc.arg(first_name), last_name = sqlc.arg(last_name), display_name = sqlc.arg(display_name), email = sqlc.arg(email), verified = sqlc.arg(verified), updated_at = sqlc.arg(updated_at) WHERE id = sqlc.arg(id);


-- name: ListUserOrganizationRoles :many
SELECT o.id as organization_id, o.name as organization, 
	o.system_name as organization_system_name,
	r.id as role_id, r.name as role_name, r.system_name as role_system_name 
FROM user_roles ur
JOIN roles r ON r.id = ur.role_id
JOIN organizations o ON o.id = r.organization_id
WHERE ur.user_id = sqlc.arg(user_id);

-- name: UsersCount :one
SELECT COUNT(*) AS count FROM users;

-- name: OrganizationsCount :one
SELECT COUNT(*) AS count FROM organizations;


-- name: UpdateUserLoginStats :exec
UPDATE users 
SET last_login = sqlc.arg(last_login), login_count = sqlc.arg(login_count) 
WHERE id = sqlc.arg(id);

-- name: ListRolesWithUserCounts :many
SELECT r.id, r.name, r.system_name, count(ur.user_id) AS user_count
FROM roles r
LEFT JOIN user_roles ur ON ur.role_id=r.id
WHERE organization_id=sqlc.arg(organization_id)
GROUP BY r.id;

-- name: UserSearch :many
SELECT id, first_name, last_name, display_name, email, verified
FROM users
WHERE email LIKE sqlc.arg(query) OR display_name LIKE sqlc.arg(query)
LIMIT sqlc.arg(count) OFFSET sqlc.arg(start);

-- name: UserAddApiKey :exec
INSERT INTO user_api_keys (id, secret_hash, user_id, organization_id, name, created_at, expires_at)
VALUES (sqlc.arg(id), sqlc.arg(secret_hash), sqlc.arg(user_id), sqlc.arg(organization_id), sqlc.arg(name), sqlc.arg(created_at), sqlc.arg(expires_at));


-- name: UserDeleteApiKey :exec
DELETE FROM user_api_keys
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);

-- name: UserGetApiKey :one
SELECT id, secret_hash, user_id, organization_id, name, created_at, expires_at
FROM user_api_keys
WHERE id = sqlc.arg(api_key_hash);

-- name: UserListApiKeys :many
SELECT id, user_id, organization_id, name, created_at, expires_at
FROM user_api_keys
WHERE user_id = sqlc.arg(user_id);


