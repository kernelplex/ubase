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

type ListUserOrganizationRolesRow struct {
	OrganizationID         int64
	Organization           string
	OrganizationSystemName string
	RoleID                 int64
	RoleName               string
	RoleSystemName         string
}

type ListRolesWithUserCountsRow struct {
	ID         int64
	Name       string
	SystemName string
	UserCount  int64
}

// DataAdapter defines the common interface for user/role/permission operations
type DataAdapter interface {

	// User operations
	AddUser(ctx context.Context, userID int64, firstName, lastName, displayName, email string, createdAt int64, updatedAt int64) error
	GetUser(ctx context.Context, userID int64) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	SearchUsers(ctx context.Context, searchTerm string, limit, offset int) ([]User, error)
	UpdateUser(ctx context.Context, userID int64, firstName, lastName, displayName, email string, updatedAt int64) error
	AddOrganization(ctx context.Context, id int64, name string, systemName string, status string) error
	GetOrganization(ctx context.Context, organizationID int64) (Organization, error)
	ListOrganizations(ctx context.Context) ([]Organization, error)
	GetOrganizationBySystemName(ctx context.Context, systemName string) (Organization, error)
	UpdateOrganization(ctx context.Context, id int64, name string, systemName string, status string) error

	// Role operations
	AddRole(ctx context.Context, roleID int64, organizationID int64, name string, systemName string) error
	UpdateRole(ctx context.Context, roleID int64, name string, systemName string) error
	DeleteRole(ctx context.Context, roleID int64) error
	GetOrganizationRoles(ctx context.Context, organizationID int64) ([]RoleRow, error)

	AddPermissionToRole(ctx context.Context, roleID int64, permission string) error
	RemovePermissionFromRole(ctx context.Context, roleID int64, permission string) error
	GetRolePermissions(ctx context.Context, roleID int64) ([]string, error)

	// User-Role operations
	AddUserToRole(ctx context.Context, userID int64, roleID int64) error
	RemoveUserFromRole(ctx context.Context, userID int64, roleID int64) error
	RemoveAllRolesFromUser(ctx context.Context, userID int64) error
	GetUserOrganizationRoles(ctx context.Context, userID int64, organizationId int64) ([]RoleRow, error)
	GetAllUserOrganizationRoles(ctx context.Context, userID int64) ([]ListUserOrganizationRolesRow, error)
	ListUserOrganizationRoles(ctx context.Context, userID int64) ([]ListUserOrganizationRolesRow, error)

	OrganizationsCount(ctx context.Context) (int64, error)
	UsersCount(ctx context.Context) (int64, error)
	UpdateUserLoginStats(ctx context.Context, userID int64, lastLogin int64, loginCount int64) error

	ListOrganizationsRolesWithUserCounts(ctx context.Context, organizationId int64) ([]ListRolesWithUserCountsRow, error)
	GetUsersInRole(ctx context.Context, roleID int64) ([]User, error)
	GetRolesForUser(ctx context.Context, userID int64) ([]Role, error)
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
	ID         int64
	Name       string
	SystemName string
	Status     string
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
