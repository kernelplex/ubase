package ubdata

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

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

func (a *PostgresAdapter) AddUser(ctx context.Context, userID int64, firstName, lastName, displayName, email string, createdAt int64, updatedAt int64) error {

	createdTime := sql.NullTime{
		Time:  time.Unix(createdAt, 0),
		Valid: true,
	}

	updatedTime := sql.NullTime{
		Time:  time.Unix(updatedAt, 0),
		Valid: true,
	}

	return a.queries.AddUser(ctx, dbpostgres.AddUserParams{
		ID:          userID,
		FirstName:   firstName,
		LastName:    lastName,
		DisplayName: displayName,
		Email:       email,
		CreatedAt:   createdTime,
		UpdatedAt:   updatedTime,
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

func (a *PostgresAdapter) UpdateUser(ctx context.Context, userID int64, firstName, lastName, displayName, email string, updatedAt int64) error {

	updatedTime := sql.NullTime{
		Time:  time.Unix(updatedAt, 0),
		Valid: true,
	}

	return a.queries.UpdateUser(ctx, dbpostgres.UpdateUserParams{
		ID:          userID,
		FirstName:   firstName,
		LastName:    lastName,
		DisplayName: displayName,
		Email:       email,
		UpdatedAt:   updatedTime,
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

	return Organization(org), nil
}

func (a *PostgresAdapter) GetOrganizationBySystemName(ctx context.Context, systemName string) (Organization, error) {
	org, err := a.queries.GetOrganizationBySystemName(ctx, systemName)
	if err != nil {
		return Organization{}, fmt.Errorf("failed to get organization by system name: %w", err)
	}
	return Organization(org), nil
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

func (a *PostgresAdapter) GetUserOrganizationRoles(ctx context.Context, userID int64, organizationId int64) ([]RoleRow, error) {
	roles, err := a.queries.GetUserOrganizationRoles(ctx, dbpostgres.GetUserOrganizationRolesParams{
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

func (a *PostgresAdapter) ListOrganizations(ctx context.Context) ([]Organization, error) {
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

func (a *PostgresAdapter) ListUserOrganizationRoles(ctx context.Context, userID int64) ([]ListUserOrganizationRolesRow, error) {
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

func (a *PostgresAdapter) GetAllUserOrganizationRoles(ctx context.Context, userID int64) ([]ListUserOrganizationRolesRow, error) {
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
			RoleSystemName:         r.Name,
		}
		result[i] = res
	}
	return result, nil
}

func (a *PostgresAdapter) OrganizationsCount(ctx context.Context) (int64, error) {
	count, err := a.queries.OrganizationsCount(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get organization count: %w", err)
	}
	return count, nil
}

func (a *PostgresAdapter) UsersCount(ctx context.Context) (int64, error) {
	count, err := a.queries.UsersCount(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get user count: %w", err)
	}
	return count, nil
}

func (a *PostgresAdapter) UpdateUserLoginStats(ctx context.Context, userID int64, lastLogin int64, loginCount int64) error {
	lastLoginTime := sql.NullTime{
		Time:  time.Unix(lastLogin, 0),
		Valid: true,
	}

	err := a.queries.UpdateUserLoginStats(ctx, dbpostgres.UpdateUserLoginStatsParams{
		ID:         userID,
		LastLogin:  lastLoginTime,
		LoginCount: int32(loginCount),
	})
	if err != nil {
		return fmt.Errorf("failed to update user last login: %w", err)
	}
	return nil

}
func (a *PostgresAdapter) ListOrganizationsRolesWithUserCounts(ctx context.Context, organizationId int64) ([]ListRolesWithUserCountsRow, error) {
	orgs, err := a.queries.ListRolesWithUserCounts(ctx, organizationId)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations with user counts: %w", err)
	}
	result := make([]ListRolesWithUserCountsRow, len(orgs))
	for i, o := range orgs {
		result[i] = ListRolesWithUserCountsRow{
			ID:         o.ID,
			Name:       o.Name,
			SystemName: o.SystemName,
			UserCount:  o.UserCount,
		}
	}
	return result, nil
}

func sqlEscapeLike(input string) string {
	var result strings.Builder
	result.Grow(len(input) * 2) // Pre-allocate to avoid reallocations

	for _, char := range input {
		if char == '%' || char == '_' || char == '\\' {
			result.WriteRune('\\')
		}
		result.WriteRune(char)
	}
	return result.String()
}

func (a *PostgresAdapter) SearchUsers(ctx context.Context, searchTerm string, limit, offset int) ([]User, error) {

	searchTerm = sqlEscapeLike(searchTerm)

	users, err := a.queries.UserSearch(ctx, dbpostgres.UserSearchParams{
		Query: "%" + searchTerm + "%",
		Count: int32(limit),
		Start: int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	result := make([]User, len(users))
	for i, u := range users {
		result[i] = User{
			UserID:      u.ID,
			FirstName:   u.FirstName,
			LastName:    u.LastName,
			DisplayName: u.DisplayName,
			Email:       u.Email,
		}
	}
	return result, nil
}

func (a *PostgresAdapter) GetUsersInRole(ctx context.Context, roleID int64) ([]User, error) {
	users, err := a.queries.GetUsersInRole(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users in role: %w", err)
	}

	result := make([]User, len(users))
	for i, u := range users {
		result[i] = User{
			UserID:      u.ID,
			FirstName:   u.FirstName,
			LastName:    u.LastName,
			DisplayName: u.DisplayName,
			Email:       u.Email,
		}
	}
	return result, nil
}

func (a *PostgresAdapter) GetRolesForUser(ctx context.Context, userID int64) ([]RoleRow, error) {
	roles, err := a.queries.GetRolesForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles for user: %w", err)
	}

	result := make([]RoleRow, len(roles))
	for i, r := range roles {
		result[i] = RoleRow{
			ID:         r.ID,
			Name:       r.Name,
			SystemName: r.SystemName,
		}
	}
	return result, nil
}
