package ubdata

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kernelplex/ubase/internal/dbsqlite"
)

type SQLiteAdapter struct {
	db      *sql.DB
	queries *dbsqlite.Queries
}

func NewSQLiteAdapter(db *sql.DB) *SQLiteAdapter {
	return &SQLiteAdapter{
		db:      db,
		queries: dbsqlite.New(db),
	}
}

func (a *SQLiteAdapter) DeleteRole(ctx context.Context, roleID int64) error {
	err := a.queries.DeleteRole(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	return nil
}

func (a *SQLiteAdapter) AddUser(ctx context.Context, userID int64, firstName, lastName, displayName, email string, createdAt int64, updatedAt int64) error {
	createdTime := sql.NullTime{
		Time:  time.Unix(createdAt, 0),
		Valid: true,
	}
	updatedTime := sql.NullTime{
		Time:  time.Unix(updatedAt, 0),
		Valid: true,
	}

	return a.queries.AddUser(ctx, dbsqlite.AddUserParams{
		ID:          userID,
		FirstName:   firstName,
		LastName:    lastName,
		DisplayName: displayName,
		Email:       email,
		CreatedAt:   createdTime,
		UpdatedAt:   updatedTime,
	})
}

func (a *SQLiteAdapter) GetUser(ctx context.Context, userID int64) (User, error) {
	user, err := a.queries.GetUser(ctx, userID)
	if err != nil {
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return User{
		UserID:      user.ID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		DisplayName: user.DisplayName,
		Email:       user.Email,
	}, nil
}

func (a *SQLiteAdapter) GetUserByEmail(ctx context.Context, email string) (User, error) {
	user, err := a.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return User{}, fmt.Errorf("failed to get user by email: %w", err)
	}
	return User{
		UserID:      user.ID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		DisplayName: user.DisplayName,
		Email:       user.Email,
	}, nil
}

func (a *SQLiteAdapter) UpdateUser(ctx context.Context, userID int64, firstName, lastName, displayName, email string, updatedAt int64) error {
	updatedTime := sql.NullTime{
		Time:  time.Unix(updatedAt, 0),
		Valid: true,
	}

	return a.queries.UpdateUser(ctx, dbsqlite.UpdateUserParams{
		ID:          userID,
		FirstName:   firstName,
		LastName:    lastName,
		DisplayName: displayName,
		Email:       email,
		UpdatedAt:   updatedTime,
	})
}

func (a *SQLiteAdapter) AddRole(ctx context.Context, roleID int64, organizationID int64, name string, systemName string) error {
	return a.queries.AddRole(ctx, dbsqlite.AddRoleParams{
		ID:             roleID,
		OrganizationID: organizationID,
		Name:           name,
		SystemName:     systemName,
	})
}

func (a *SQLiteAdapter) UpdateRole(ctx context.Context, roleID int64, name string, systemName string) error {
	params := dbsqlite.UpdateRoleParams{
		Name:       name,
		SystemName: systemName,
		ID:         roleID,
	}

	err := a.queries.UpdateRole(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}
	return nil
}

func (a *SQLiteAdapter) AddOrganization(ctx context.Context, id int64, name string, systemName string, status string) error {

	err := a.queries.AddOrganization(ctx, dbsqlite.AddOrganizationParams{
		ID:         id,
		Name:       name,
		SystemName: systemName,
		Status:     status,
	})
	if err != nil {
		return fmt.Errorf("failed to add organization: %w", err)
	}
	return nil
}

func (a *SQLiteAdapter) GetOrganization(ctx context.Context, organizationID int64) (Organization, error) {
	org, err := a.queries.GetOrganization(ctx, organizationID)
	if err != nil {
		return Organization{}, fmt.Errorf("failed to get organization: %w", err)
	}
	return Organization(org), nil
}

func (a *SQLiteAdapter) GetOrganizationBySystemName(ctx context.Context, systemName string) (Organization, error) {
	org, err := a.queries.GetOrganizationBySystemName(ctx, systemName)
	if err != nil {
		return Organization{}, fmt.Errorf("failed to get organization by system name: %w", err)
	}
	return Organization(org), nil
}

func (a *SQLiteAdapter) GetOrganizationRoles(ctx context.Context, organizationID int64) ([]RoleRow, error) {
	roles, err := a.queries.GetOrganizationRoles(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization roles: %w", err)
	}
	result := make([]RoleRow, len(roles))
	for i, r := range roles {
		result[i] = RoleRow(r)
	}
	return result, nil
}

func (a *SQLiteAdapter) UpdateOrganization(ctx context.Context, id int64, name string, systemName string, status string) error {
	err := a.queries.UpdateOrganization(ctx, dbsqlite.UpdateOrganizationParams{
		ID:         id,
		Name:       name,
		SystemName: systemName,
		Status:     status,
	})
	if err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}
	return nil
}

func (a *SQLiteAdapter) AddPermissionToRole(ctx context.Context, roleID int64, permission string) error {
	err := a.queries.AddPermissionToRole(ctx, dbsqlite.AddPermissionToRoleParams{
		RoleID:     roleID,
		Permission: permission,
	})
	if err != nil {
		return fmt.Errorf("failed to add permission to role: %w", err)
	}
	return nil
}

func (a *SQLiteAdapter) RemovePermissionFromRole(ctx context.Context, roleID int64, permission string) error {
	err := a.queries.RemovePermissionFromRole(ctx, dbsqlite.RemovePermissionFromRoleParams{
		RoleID:     roleID,
		Permission: permission,
	})
	if err != nil {
		return fmt.Errorf("failed to remove permission from role: %w", err)
	}
	return nil
}

func (a *SQLiteAdapter) GetRolePermissions(ctx context.Context, roleID int64) ([]string, error) {
	perms, err := a.queries.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	return perms, nil
}

func (a *SQLiteAdapter) AddUserToRole(ctx context.Context, userID int64, roleID int64) error {
	err := a.queries.AddUserToRole(ctx, dbsqlite.AddUserToRoleParams{
		UserID: userID,
		RoleID: roleID,
	})
	if err != nil {
		return fmt.Errorf("failed to add user to role: %w", err)
	}
	return nil
}

func (a *SQLiteAdapter) RemoveUserFromRole(ctx context.Context, userID int64, roleID int64) error {
	err := a.queries.RemoveUserFromRole(ctx, dbsqlite.RemoveUserFromRoleParams{
		UserID: userID,
		RoleID: roleID,
	})
	if err != nil {
		return fmt.Errorf("failed to remove user from role: %w", err)
	}
	return nil
}

func (a *SQLiteAdapter) RemoveAllRolesFromUser(ctx context.Context, userID int64) error {
	err := a.queries.RemoveAllRolesFromUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to remove all roles from user: %w", err)
	}
	return nil
}

func (a *SQLiteAdapter) GetUserOrganizationRoles(ctx context.Context, userID int64, organizationId int64) ([]RoleRow, error) {
	roles, err := a.queries.GetUserOrganizationRoles(ctx, dbsqlite.GetUserOrganizationRolesParams{
		UserID:         userID,
		OrganizationID: organizationId,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get user organization roles: %w", err)
	}

	result := make([]RoleRow, len(roles))
	for i, r := range roles {
		result[i] = RoleRow(r)
	}

	return result, nil
}

func (a *SQLiteAdapter) ListOrganizations(ctx context.Context) ([]Organization, error) {
	orgs, err := a.queries.ListOrganizations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}
	result := make([]Organization, len(orgs))
	for i, o := range orgs {
		result[i] = Organization(o)
	}
	return result, nil
}

func (a *SQLiteAdapter) ListUserOrganizationRoles(ctx context.Context, userID int64) ([]ListUserOrganizationRolesRow, error) {
	roles, err := a.queries.ListUserOrganizationRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user organization roles: %w", err)
	}
	result := make([]ListUserOrganizationRolesRow, len(roles))
	for i, r := range roles {
		result[i] = ListUserOrganizationRolesRow(r)
	}
	return result, nil
}

func (a *SQLiteAdapter) GetAllUserOrganizationRoles(ctx context.Context, userID int64) ([]ListUserOrganizationRolesRow, error) {
	roles, err := a.queries.GetAllUserOrganizationRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all user organization roles: %w", err)
	}
	result := make([]ListUserOrganizationRolesRow, len(roles))
	for i, r := range roles {
		res := ListUserOrganizationRolesRow{
			OrganizationID:         r.OrganizationID,
			Organization:           r.OrganizationName,
			OrganizationSystemName: r.OrganizationSystemName,
			RoleID:                 r.ID,
			RoleName:               r.Name,
			RoleSystemName:         r.SystemName,
		}
		result[i] = res
	}
	return result, nil
}
