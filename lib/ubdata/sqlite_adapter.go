package ubdata

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/kernelplex/ubase/lib/dbsqlite"
	"github.com/kernelplex/ubase/lib/ubstate"
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

func (a *SQLiteAdapter) AddUser(ctx context.Context, userID int64, firstName, lastName, displayName, email string) error {
	return a.queries.AddUser(ctx, dbsqlite.AddUserParams{
		UserID:      userID,
		FirstName:   firstName,
		LastName:    lastName,
		DisplayName: displayName,
		Email:       email,
	})
}

func (a *SQLiteAdapter) GetUser(ctx context.Context, userID int64) (User, error) {
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

func (a *SQLiteAdapter) GetUserByEmail(ctx context.Context, email string) (User, error) {
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

func (a *SQLiteAdapter) UpdateUser(ctx context.Context, userID int64, firstName, lastName, displayName, email string) error {
	return a.queries.UpdateUser(ctx, dbsqlite.UpdateUserParams{
		UserID:      userID,
		FirstName:   firstName,
		LastName:    lastName,
		DisplayName: displayName,
		Email:       email,
	})
}

func (a *SQLiteAdapter) AddRole(ctx context.Context, roleID int64, name string) error {
	return a.queries.AddRole(ctx, dbsqlite.AddRoleParams{
		RoleID: roleID,
		Name:   name,
	})
}

func (a *SQLiteAdapter) UpdateRole(ctx context.Context, roleID int64, name string) error {
	return a.queries.UpdateRole(ctx, dbsqlite.UpdateRoleParams{
		Name:   name,
		RoleID: roleID,
	})
}

func (a *SQLiteAdapter) GetRoles(ctx context.Context) ([]Role, error) {
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

func (a *SQLiteAdapter) CreatePermission(ctx context.Context, name string) (int64, error) {
	id, err := a.queries.CreatePermission(ctx, name)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (a *SQLiteAdapter) GetPermissions(ctx context.Context) ([]Permission, error) {
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

func (a *SQLiteAdapter) AddPermissionToRole(ctx context.Context, roleID, permissionID int64) error {
	return a.queries.AddPermissionToRole(ctx, dbsqlite.AddPermissionToRoleParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	})
}

func (a *SQLiteAdapter) RemovePermissionFromRole(ctx context.Context, roleID, permissionID int64) error {
	return a.queries.RemovePermissionFromRole(ctx, dbsqlite.RemovePermissionFromRoleParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	})
}

func (a *SQLiteAdapter) GetRolePermissions(ctx context.Context, roleID int64) ([]Permission, error) {
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

func (a *SQLiteAdapter) AddRoleToUser(ctx context.Context, userID, roleID int64) error {
	return a.queries.AddRoleToUser(ctx, dbsqlite.AddRoleToUserParams{
		UserID: userID,
		RoleID: roleID,
	})
}

func (a *SQLiteAdapter) RemoveAllRolesFromUser(ctx context.Context, userID int64) error {
	return a.queries.RemoveAllRolesFromUser(ctx, userID)
}

func (a *SQLiteAdapter) GetUserRoles(ctx context.Context, userID int64) ([]Role, error) {
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

func (a *SQLiteAdapter) GetUserPermissions(ctx context.Context, userID int64) ([]Permission, error) {
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

func (a *SQLiteAdapter) ProjectUser(ctx context.Context, userID int64, userState ubstate.UserState) error {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	queries := dbsqlite.New(tx)
	_, err = queries.GetUser(ctx, userID)

	// If the user doesn't exist, create it
	if err != nil {
		if err != sql.ErrNoRows {
			slog.Error("Failed to get user", "error", err)
			return err
		}

		addUserParams := dbsqlite.AddUserParams{
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
		updateUserParams := dbsqlite.UpdateUserParams{
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

func (a *SQLiteAdapter) projectUserRoles(ctx context.Context, queries *dbsqlite.Queries, userID int64, stateRoles []int64) error {
	// Remove all existing roles
	err := queries.RemoveAllRolesFromUser(ctx, userID)
	if err != nil {
		slog.Error("Failed to remove all roles from user", "error", err)
		return err
	}

	// Add roles
	for _, roleId := range stateRoles {
		addRoleToUserParams := dbsqlite.AddRoleToUserParams{
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

func (a *SQLiteAdapter) ProjectUserRoles(ctx context.Context, userID int64, stateRoles []int64) error {
	return a.projectUserRoles(ctx, a.queries, userID, stateRoles)
}
