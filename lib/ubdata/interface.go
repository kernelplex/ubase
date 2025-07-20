package ubdata

import (
	"context"
	// "github.com/kernelplex/ubase/lib/ubstate"
)

type RoleRow struct {
	ID         int64
	Name       string
	SystemName string
}

// DataAdapter defines the common interface for user/role/permission operations
type DataAdapter interface {

	// User operations
	AddUser(ctx context.Context, userID int64, firstName, lastName, displayName, email string) error
	GetUser(ctx context.Context, userID int64) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	UpdateUser(ctx context.Context, userID int64, firstName, lastName, displayName, email string) error
	AddOrganization(ctx context.Context, id int64, name string, systemName string, status string) error
	GetOrganization(ctx context.Context, organizationID int64) (Organization, error)
	GetOrganizationBySystemName(ctx context.Context, systemName string) (Organization, error)
	UpdateOrganization(ctx context.Context, id int64, name string, systemName string, status string) error

	// Role operations
	AddRole(ctx context.Context, roleID int64, organizationID int64, name string, systemName string) error
	UpdateRole(ctx context.Context, roleID int64, name string, systemName string) error
	DeleteRole(ctx context.Context, roleID int64) error
	GetOrganizationRoles(ctx context.Context, organizationID int64) ([]RoleRow, error)

	// Permission operations
	AddPermission(ctx context.Context, id int64, name string) error
	GetPermissions(ctx context.Context) ([]Permission, error)

	AddPermissionToRole(ctx context.Context, roleID int64, permissionID int64) error
	RemovePermissionFromRole(ctx context.Context, roleID int64, permissionID int64) error

	// User-Role operations
	AddUserToRole(ctx context.Context, userID int64, roleID int64) error
	RemoveUserFromRole(ctx context.Context, userID int64, roleID int64) error
	RemoveAllRolesFromUser(ctx context.Context, userID int64) error
}

// User represents a user in the system
type User struct {
	UserID      int64
	FirstName   string
	LastName    string
	DisplayName string
	Email       string
}

type Organization struct {
	OrganizationID int64
	Name           string
	SystemName     string
	Status         string
}

// Role represents a role in the system
type Role struct {
	RoleID     int64
	Name       string
	SystemName string
}

// Permission represents a permission in the system
type Permission struct {
	PermissionID int64
	SystemName   string
}
