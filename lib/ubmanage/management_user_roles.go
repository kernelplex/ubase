package ubmanage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	evercore "github.com/kernelplex/evercore/base"
	"github.com/kernelplex/ubase/lib/ubdata"
	r "github.com/kernelplex/ubase/lib/ubresponse"
	"github.com/kernelplex/ubase/lib/ubstatus"
)

func (m *ManagementImpl) UserGetOrganizationRoles(ctx context.Context, userId int64, organizationId int64) (r.Response[[]ubdata.RoleRow], error) {
	roles, err := m.dbadapter.GetUserOrganizationRoles(ctx, userId, organizationId)
	if err != nil {
		return r.Response[[]ubdata.RoleRow]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error getting user organization roles",
		}, err
	}

	return r.Response[[]ubdata.RoleRow]{
		Status: ubstatus.Success,
		Data:   roles,
	}, nil
}

func (m *ManagementImpl) UserAddToRole(ctx context.Context,
	command UserAddToRoleCommand,
	agent string) (r.Response[any], error) {

	// Validation
	ok, issues := command.Validate()
	if !ok {
		return r.Response[any]{
			Status:           ubstatus.ValidationError,
			Message:          "Validation issues",
			ValidationIssues: issues,
		}, nil
	}

	err := m.store.WithContext(
		ctx,
		func(etx evercore.EventStoreContext) error {
			aggregate := UserRolesAggregate{}
			// Identity aggregate
			_, err := etx.LoadOrCreateAggregate(&aggregate, "UserRolesAggregate")
			if err != nil {
				return fmt.Errorf("failed to load user by ID: %w", err)
			}

			event := UserAddedToRoleEvent{
				UserId: command.UserId,
				RoleId: command.RoleId,
			}
			err = etx.ApplyEventTo(&aggregate, event, time.Now(), agent)
			if err != nil {
				return fmt.Errorf("failed to apply user added to role event: %w", err)
			}

			err = m.dbadapter.AddUserToRole(ctx, command.UserId, command.RoleId)
			if err != nil {
				return fmt.Errorf("failed to add user to role in database: %w", err)
			}

			return nil
		})

	if err != nil {
		slog.Error("Error adding user to role", "error", err)
		return r.Response[any]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error adding user to role",
		}, err
	}

	return r.Response[any]{
		Status: ubstatus.Success,
	}, nil
}

func (m *ManagementImpl) UserRemoveFromRole(ctx context.Context,
	command UserRemoveFromRoleCommand,
	agent string) (r.Response[any], error) {

	// Validation
	ok, issues := command.Validate()
	if !ok {
		return r.Response[any]{
			Status:           ubstatus.ValidationError,
			Message:          "Validation issues",
			ValidationIssues: issues,
		}, nil
	}

	err := m.store.WithContext(
		ctx,
		func(etx evercore.EventStoreContext) error {
			aggregate := UserRolesAggregate{}
			_, err := etx.LoadOrCreateAggregate(&aggregate, "UserRolesAggregate")
			if err != nil {
				return fmt.Errorf("failed to load user by ID: %w", err)
			}

			event := UserRemovedFromRoleEvent{
				UserId: command.UserId,
				RoleId: command.RoleId,
			}
			err = etx.ApplyEventTo(&aggregate, event, time.Now(), agent)
			if err != nil {
				return fmt.Errorf("failed to apply user removed from role event: %w", err)
			}

			err = m.dbadapter.RemoveUserFromRole(ctx, command.UserId, command.RoleId)
			if err != nil {
				return fmt.Errorf("failed to remove user from role in database: %w", err)
			}

			return nil
		})

	if err != nil {
		slog.Error("Error removing user from role", "error", err)
		return r.Response[any]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error removing user from role",
		}, err
	}

	return r.Response[any]{
		Status: ubstatus.Success,
	}, nil
}
