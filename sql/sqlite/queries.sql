-- ---------------------------------------------------------------------------
--
-- User and Role management
--
-- ---------------------------------------------------------------------------


-- name: AddUser :one
INSERT INTO users (user_id, first_name, last_name, display_name, email) 
	VALUES (sqlc.arg(user_id), sqlc.arg(first_name), sqlc.arg(last_name), sqlc.arg(display_name), sqlc.arg(email))
	RETURNING user_id;


-- name: UpdateUser :exec
UPDATE users 
SET last_name = sqlc.arg(last_name), 
    first_name = sqlc.arg(first_name), 
    display_name = sqlc.arg(display_name), 
    email = sqlc.arg(email) 
WHERE user_id = sqlc.arg(user_id);


-- name: GetUser :one
select * from users where user_id = sqlc.arg(user_id);

-- name: GetUserByEmail :one
select * from users where email = sqlc.arg(email);

-- name: AddRole :one
insert into roles (role_id, name) 
	VALUES (sqlc.arg(role_id), sqlc.arg(name))
	RETURNING role_id;


-- name: UpdateRole :exec
UPDATE roles SET name = sqlc.arg(name) WHERE role_id = sqlc.arg(role_id);

-- AddUserToRole :exec
INSERT INTO user_roles (user_id, role_id) 
	VALUES (sqlc.arg(user_id), sqlc.arg(role_id));

-- RemoveUserFromRole :exec
DELETE FROM user_roles WHERE user_id = sqlc.arg(user_id) AND role_id = sqlc.arg(role_id);

-- name: RemoveAllRolesFromUser :exec
DELETE FROM user_roles WHERE user_id = sqlc.arg(user_id);

-- name: AddRoleToUser :exec
INSERT INTO user_roles (user_id, role_id) 
	VALUES (sqlc.arg(user_id), sqlc.arg(role_id));

-- name: GetRoles :many
SELECT role_id, name FROM roles;

-- name: GetUserRoles :many
SELECT r.role_id, r.name FROM user_roles ur
LEFT JOIN roles r ON r.role_id = ur.role_id
WHERE ur.user_id = sqlc.arg(user_id);

-- name: AddPermissionToRole :exec
INSERT INTO role_permissions (role_id, permission_id) 
	VALUES (sqlc.arg(role_id), sqlc.arg(permission_id));

-- name: RemovePermissionFromRole :exec
DELETE FROM role_permissions WHERE role_id = sqlc.arg(role_id) AND permission_id = sqlc.arg(permission_id);


-- name: CreatePermission :one
INSERT INTO permissions (name) 
	VALUES (sqlc.arg(name))
	RETURNING permission_id;

-- name: GetPermissions :many
SELECT permission_id, name FROM permissions;

-- name: GetRolePermissions :many
SELECT p.permission_id, p.name FROM role_permissions rp
LEFT JOIN permissions p ON p.permission_id = rp.permission_id
WHERE rp.role_id = sqlc.arg(role_id);

-- name: GetUserPermissions :many
SELECT p.permission_id, p.name
FROM user_roles up
LEFT JOIN role_permissions rp ON rp.role_id = up.role_id
LEFT JOIN permissions p ON p.permission_id = up.permission_id
WHERE up.user_id = sqlc.arg(user_id);

