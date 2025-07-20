package ubdata

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kernelplex/ubase/internal/dbpostgres"
)

type PostgresAdapter struct {
	db      *sql.DB
	queries *dbpostgres.Queries
}

func NewPostgresAdapter(db *sql.DB) *PostgresAdapter {
	return &PostgresAdapter{
		db:      db,
		queries: dbpostgres.New(db),
	}
}

func (a *PostgresAdapter) DeleteRole(ctx context.Context, roleID int64) error {
	err := a.queries.DeleteRole(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	return nil
}

func (a *PostgresAdapter) AddUser(ctx context.Context, userID int64, firstName, lastName, displayName, email string) error {
	return a.queries.AddUser(ctx, dbpostgres.AddUserParams{
		ID:          userID,
		FirstName:   firstName,
		LastName:    lastName,
		DisplayName: displayName,
		Email:       email,
	})
}

func (a *PostgresAdapter) GetUser(ctx context.Context, userID int64) (User, error) {
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

func (a *PostgresAdapter) GetUserByEmail(ctx context.Context, email string) (User, error) {
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

func (a *PostgresAdapter) UpdateUser(ctx context.Context, userID int64, firstName, lastName, displayName, email string) error {
	return a.queries.UpdateUser(ctx, dbpostgres.UpdateUserParams{
		ID:          userID,
		FirstName:   firstName,
		LastName:    lastName,
		DisplayName: displayName,
		Email:       email,
	})
}

func (a *PostgresAdapter) AddRole(ctx context.Context, roleID int64, organizationID int64, name string, systemName string) error {
	return a.queries.AddRole(ctx, dbpostgres.AddRoleParams{
		ID:             roleID,
		OrganizationID: organizationID,
		Name:           name,
		SystemName:     systemName,
	})
}

func (a *PostgresAdapter) UpdateRole(ctx context.Context, roleID int64, name string, systemName string) error {
	params := dbpostgres.UpdateRoleParams{
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

func (a *PostgresAdapter) AddOrganization(ctx context.Context, id int64, name string, systemName string, status string) error {

	err := a.queries.AddOrganization(ctx, dbpostgres.AddOrganizationParams{
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

func (a *PostgresAdapter) GetOrganization(ctx context.Context, organizationID int64) (Organization, error) {
	org, err := a.queries.GetOrganization(ctx, organizationID)
	if err != nil {
		return Organization{}, fmt.Errorf("failed to get organization: %w", err)
	}
	return Organization{
		OrganizationID: org.ID,
		Name:           org.Name,
		SystemName:     org.SystemName,
		Status:         org.Status,
	}, nil
}

func (a *PostgresAdapter) GetOrganizationBySystemName(ctx context.Context, systemName string) (Organization, error) {
	org, err := a.queries.GetOrganizationBySystemName(ctx, systemName)
	if err != nil {
		return Organization{}, fmt.Errorf("failed to get organization by system name: %w", err)
	}
	return Organization{
		OrganizationID: org.ID,
		Name:           org.Name,
		SystemName:     org.SystemName,
		Status:         org.Status,
	}, nil
}

func (a *PostgresAdapter) GetOrganizationRoles(ctx context.Context, organizationID int64) ([]RoleRow, error) {
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

func (a *PostgresAdapter) UpdateOrganization(ctx context.Context, id int64, name string, systemName string, status string) error {
	err := a.queries.UpdateOrganization(ctx, dbpostgres.UpdateOrganizationParams{
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

func (a *PostgresAdapter) AddPermissionToRole(ctx context.Context, roleID int64, permission string) error {
	err := a.queries.AddPermissionToRole(ctx, dbpostgres.AddPermissionToRoleParams{
		RoleID:     roleID,
		Permission: permission,
	})
	if err != nil {
		return fmt.Errorf("failed to add permission to role: %w", err)
	}
	return nil
}

func (a *PostgresAdapter) GetRolePermissions(ctx context.Context, roleID int64) ([]string, error) {
	perms, err := a.queries.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	return perms, nil
}

func (a *PostgresAdapter) RemovePermissionFromRole(ctx context.Context, roleID int64, permission string) error {
	err := a.queries.RemovePermissionFromRole(ctx, dbpostgres.RemovePermissionFromRoleParams{
		RoleID:     roleID,
		Permission: permission,
	})
	if err != nil {
		return fmt.Errorf("failed to remove permission from role: %w", err)
	}
	return nil
}

func (a *PostgresAdapter) AddUserToRole(ctx context.Context, userID int64, roleID int64) error {
	err := a.queries.AddUserToRole(ctx, dbpostgres.AddUserToRoleParams{
		UserID: userID,
		RoleID: roleID,
	})
	if err != nil {
		return fmt.Errorf("failed to add user to role: %w", err)
	}
	return nil
}

func (a *PostgresAdapter) RemoveUserFromRole(ctx context.Context, userID int64, roleID int64) error {
	err := a.queries.RemoveUserFromRole(ctx, dbpostgres.RemoveUserFromRoleParams{
		UserID: userID,
		RoleID: roleID,
	})
	if err != nil {
		return fmt.Errorf("failed to remove user from role: %w", err)
	}
	return nil
}

func (a *PostgresAdapter) RemoveAllRolesFromUser(ctx context.Context, userID int64) error {
	err := a.queries.RemoveAllRolesFromUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to remove all roles from user: %w", err)
	}
	return nil
}
