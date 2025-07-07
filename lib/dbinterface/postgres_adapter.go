package dbinterface

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/kernelplex/ubase/lib/dbpostgres"
	"github.com/kernelplex/ubase/lib/ubstate"
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

func (a *PostgresAdapter) AddUser(ctx context.Context, userID int64, firstName, lastName, displayName, email string) error {
	return a.queries.AddUser(ctx, dbpostgres.AddUserParams{
		UserID:      userID,
		FirstName:   firstName,
		LastName:    lastName,
		DisplayName: displayName,
		Email:       email,
	})
}

func (a *PostgresAdapter) GetUser(ctx context.Context, userID int64) (User, error) {
	user, err := a.queries.GetUser(ctx, userID)
	if err != nil {
		return User{}, err
	}
	return User{
		UserID:      user.UserID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		DisplayName: user.DisplayName,
		Email:       user.Email,
	}, nil
}

func (a *PostgresAdapter) GetUserByEmail(ctx context.Context, email string) (User, error) {
	user, err := a.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return User{}, err
	}
	return User{
		UserID:      user.UserID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		DisplayName: user.DisplayName,
		Email:       user.Email,
	}, nil
}

func (a *PostgresAdapter) UpdateUser(ctx context.Context, userID int64, firstName, lastName, displayName, email string) error {
	return a.queries.UpdateUser(ctx, dbpostgres.UpdateUserParams{
		UserID:      userID,
		FirstName:   firstName,
		LastName:    lastName,
		DisplayName: displayName,
		Email:       email,
	})
}

func (a *PostgresAdapter) AddRole(ctx context.Context, roleID int64, name string) error {
	return a.queries.AddRole(ctx, dbpostgres.AddRoleParams{
		RoleID: roleID,
		Name:   name,
	})
}

func (a *PostgresAdapter) UpdateRole(ctx context.Context, roleID int64, name string) error {
	return a.queries.UpdateRole(ctx, dbpostgres.UpdateRoleParams{
		Name:   name,
		RoleID: roleID,
	})
}

func (a *PostgresAdapter) GetRoles(ctx context.Context) ([]Role, error) {
	roles, err := a.queries.GetRoles(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]Role, len(roles))
	for i, r := range roles {
		result[i] = Role{
			RoleID: r.RoleID,
			Name:   r.Name,
		}
	}
	return result, nil
}

func (a *PostgresAdapter) CreatePermission(ctx context.Context, name string) (int64, error) {
	var id int64

	id, err := a.queries.CreatePermission(ctx, name)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (a *PostgresAdapter) GetPermissions(ctx context.Context) ([]Permission, error) {
	perms, err := a.queries.GetPermissions(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]Permission, len(perms))
	for i, p := range perms {
		result[i] = Permission{
			PermissionID: p.PermissionID,
			Name:         p.Name,
		}
	}
	return result, nil
}

func (a *PostgresAdapter) AddPermissionToRole(ctx context.Context, roleID, permissionID int64) error {
	return a.queries.AddPermissionToRole(ctx, dbpostgres.AddPermissionToRoleParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	})
}

func (a *PostgresAdapter) RemovePermissionFromRole(ctx context.Context, roleID, permissionID int64) error {
	return a.queries.RemovePermissionFromRole(ctx, dbpostgres.RemovePermissionFromRoleParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	})
}

func (a *PostgresAdapter) GetRolePermissions(ctx context.Context, roleID int64) ([]Permission, error) {
	perms, err := a.queries.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}
	result := make([]Permission, 0, len(perms))
	for _, p := range perms {
		if p.PermissionID.Valid && p.Name.Valid {
			result = append(result, Permission{
				PermissionID: p.PermissionID.Int64,
				Name:         p.Name.String,
			})
		}
	}
	return result, nil
}

func (a *PostgresAdapter) AddRoleToUser(ctx context.Context, userID, roleID int64) error {
	return a.queries.AddRoleToUser(ctx, dbpostgres.AddRoleToUserParams{
		UserID: userID,
		RoleID: roleID,
	})
}

func (a *PostgresAdapter) RemoveAllRolesFromUser(ctx context.Context, userID int64) error {
	return a.queries.RemoveAllRolesFromUser(ctx, userID)
}

func (a *PostgresAdapter) GetUserRoles(ctx context.Context, userID int64) ([]Role, error) {
	roles, err := a.queries.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]Role, 0, len(roles))
	for _, r := range roles {
		if r.RoleID.Valid && r.Name.Valid {
			result = append(result, Role{
				RoleID: r.RoleID.Int64,
				Name:   r.Name.String,
			})
		}
	}
	return result, nil
}

func (a *PostgresAdapter) GetUserPermissions(ctx context.Context, userID int64) ([]Permission, error) {
	perms, err := a.queries.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]Permission, 0, len(perms))
	for _, p := range perms {
		if p.PermissionID.Valid && p.Name.Valid {
			result = append(result, Permission{
				PermissionID: p.PermissionID.Int64,
				Name:         p.Name.String,
			})
		}
	}
	return result, nil
}

func (a *PostgresAdapter) ProjectUser(ctx context.Context, userID int64, userState ubstate.UserState) error {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	queries := dbpostgres.New(tx)
	_, err = queries.GetUser(ctx, userID)

	// If the user doesn't exist, create it
	if err != nil {
		if err != sql.ErrNoRows {
			slog.Error("Failed to get user", "error", err)
			return err
		}

		addUserParams := dbpostgres.AddUserParams{
			UserID:      userID,
			FirstName:   userState.FirstName,
			LastName:    userState.LastName,
			DisplayName: userState.DisplayName,
			Email:       userState.Email,
		}

		err = queries.AddUser(ctx, addUserParams)
		if err != nil {
			slog.Error("Failed to project user", "error", err)
			return err
		}
	} else {
		// If the user exists, update it
		updateUserParams := dbpostgres.UpdateUserParams{
			LastName:    userState.LastName,
			FirstName:   userState.FirstName,
			DisplayName: userState.DisplayName,
			Email:       userState.Email,
			UserID:      userID,
		}
		err = queries.UpdateUser(ctx, updateUserParams)
		if err != nil {
			slog.Error("Failed to project user", "error", err)
			return err
		}
	}

	err = a.projectUserRoles(ctx, queries, userID, userState.Roles)
	if err != nil {
		slog.Error("Failed to project user roles", "error", err)
		return err
	}

	tx.Commit()

	return nil
}

func (a *PostgresAdapter) projectUserRoles(ctx context.Context, queries *dbpostgres.Queries, userID int64, stateRoles []int64) error {
	// Remove all existing roles
	err := queries.RemoveAllRolesFromUser(ctx, userID)
	if err != nil {
		slog.Error("Failed to remove all roles from user", "error", err)
		return err
	}

	// Add roles
	for _, roleId := range stateRoles {
		addRoleToUserParams := dbpostgres.AddRoleToUserParams{
			UserID: userID,
			RoleID: roleId,
		}
		err = queries.AddRoleToUser(ctx, addRoleToUserParams)
		if err != nil {
			slog.Error("Failed to add role to user", "error", err)
			return err
		}
	}
	return nil
}

func (a *PostgresAdapter) ProjectUserRoles(ctx context.Context, userID int64, stateRoles []int64) error {
	return a.projectUserRoles(ctx, a.queries, userID, stateRoles)
}
