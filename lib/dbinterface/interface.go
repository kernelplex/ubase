package dbinterface

import (
	"context"
	"github.com/kernelplex/ubase/lib/ubstate"
)

// DataAdapter defines the common interface for user/role/permission operations
type DataAdapter interface {
	// User operations
	AddUser(ctx context.Context, userID int64, firstName, lastName, displayName, email string) error
	GetUser(ctx context.Context, userID int64) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	UpdateUser(ctx context.Context, userID int64, firstName, lastName, displayName, email string) error

	// Role operations
	AddRole(ctx context.Context, roleID int64, name string) error
	UpdateRole(ctx context.Context, roleID int64, name string) error
	GetRoles(ctx context.Context) ([]Role, error)

	// Permission operations
	CreatePermission(ctx context.Context, name string) (int64, error)
	GetPermissions(ctx context.Context) ([]Permission, error)

	// Role-Permission operations
	AddPermissionToRole(ctx context.Context, roleID, permissionID int64) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID int64) error
	GetRolePermissions(ctx context.Context, roleID int64) ([]Permission, error)

	// User-Role operations
	AddRoleToUser(ctx context.Context, userID, roleID int64) error
	RemoveAllRolesFromUser(ctx context.Context, userID int64) error
	GetUserRoles(ctx context.Context, userID int64) ([]Role, error)
	GetUserPermissions(ctx context.Context, userID int64) ([]Permission, error)

	ProjectUser(ctx context.Context, userID int64, userState ubstate.UserState) error
	ProjectUserRoles(ctx context.Context, userID int64, stateRoles []int64) error
}

// User represents a user in the system
type User struct {
	UserID      int64
	FirstName   string
	LastName    string
	DisplayName string
	Email       string
}

// Role represents a role in the system
type Role struct {
	RoleID int64
	Name   string
}

// Permission represents a permission in the system
type Permission struct {
	PermissionID int64
	Name         string
}
