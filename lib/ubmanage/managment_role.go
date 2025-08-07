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

func (m *ManagementImpl) RoleList(ctx context.Context, OrganizationId int64) (r.Response[[]ubdata.RoleRow], error) {
	roles, err := m.dbadapter.GetOrganizationRoles(ctx, OrganizationId)
	if err != nil {
		return r.Response[[]ubdata.RoleRow]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error listing roles",
		}, err
	}

	result := make([]ubdata.RoleRow, len(roles))
	for i, r := range roles {
		result[i] = ubdata.RoleRow(r)
	}

	return r.Response[[]ubdata.RoleRow]{
		Status: ubstatus.Success,
		Data:   result,
	}, nil
}

func (m *ManagementImpl) RoleAdd(ctx context.Context,
	command RoleCreateCommand,
	agent string) (r.Response[IdValue], error) {

	// Validation
	ok, issues := command.Validate()
	if !ok {
		return r.Response[IdValue]{
			Status:           ubstatus.ValidationError,
			Message:          "Validation issues",
			ValidationIssues: issues,
		}, nil
	}

	id, err := evercore.InContext(
		ctx,
		m.store,
		func(etx evercore.EventStoreContext) (int64, error) {
			aggregate := RoleAggregate{}
			err := etx.CreateAggregateWithKeyInto(&aggregate, command.SystemName)
			if err != nil {
				return 0, fmt.Errorf("failed to create aggregate: %w", err)
			}
			event := evercore.NewStateEvent(RoleCreatedEvent{
				OrganizationId: command.OrganizationId,
				Name:           command.Name,
				SystemName:     command.SystemName,
			})
			err = etx.ApplyEventTo(&aggregate, event, time.Now(), agent)
			if err != nil {
				return 0, fmt.Errorf("failed to apply role added event: %w", err)
			}

			err = m.dbadapter.AddRole(ctx, aggregate.Id, aggregate.State.OrganizationId, aggregate.State.Name, aggregate.State.SystemName)
			if err != nil {
				return 0, fmt.Errorf("failed to add role in database: %w", err)
			}

			return aggregate.Id, nil
		})

	if err != nil {
		status := MapEvercoreErrorToStatus(err)
		slog.Error("Error creating role", "error", err)
		return r.Response[IdValue]{
			Status:  status,
			Message: "Error creating role",
		}, err
	}

	return r.Response[IdValue]{
		Status: ubstatus.Success,
		Data: IdValue{
			Id: id,
		},
	}, nil
}

func (m *ManagementImpl) RoleUpdate(ctx context.Context,
	command RoleUpdateCommand,
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
			aggregate := RoleAggregate{}
			err := etx.LoadStateInto(&aggregate, command.Id)
			if err != nil {
				return fmt.Errorf("failed to load role: %w", err)
			}

			// Save the previous system name so we can check if it has changed
			previousSystemName := aggregate.State.SystemName

			event := evercore.NewStateEvent(RoleUpdatedEvent{
				Id:         aggregate.Id,
				Name:       command.Name,
				SystemName: command.SystemName,
			})

			err = etx.ApplyEventTo(&aggregate, event, time.Now(), agent)
			if err != nil {
				return fmt.Errorf("failed to apply role updated event: %w", err)
			}

			// Be sure to update the natural key if it has changed
			if command.SystemName != nil && aggregate.State.SystemName != previousSystemName {
				err = etx.ChangeAggregateNaturalKey(aggregate.Id, *command.SystemName)
				if err != nil {
					return fmt.Errorf("failed to change role natural key: %w", err)
				}
			}

			err = m.dbadapter.UpdateRole(ctx, aggregate.Id, aggregate.State.Name, aggregate.State.SystemName)
			if err != nil {
				return fmt.Errorf("failed to update role in database: %w", err)
			}

			return nil
		})

	if err != nil {
		slog.Error("Error updating role", "error", err)
		return r.Response[any]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error updating role",
		}, err
	}

	return r.Response[any]{
		Status: ubstatus.Success,
	}, nil
}

func (m *ManagementImpl) RoleDelete(ctx context.Context,
	command RoleDeleteCommand,
	agent string) (r.Response[any], error) {

	err := m.store.WithContext(
		ctx,
		func(etx evercore.EventStoreContext) error {
			aggregate := RoleAggregate{}
			err := etx.LoadStateInto(&aggregate, command.Id)
			if err != nil {
				return fmt.Errorf("failed to load role: %w", err)
			}

			event := evercore.NewStateEvent(RoleDeletedEvent{})

			err = etx.ApplyEventTo(&aggregate, event, time.Now(), agent)
			if err != nil {
				return fmt.Errorf("failed to apply role deleted event: %w", err)
			}

			err = m.dbadapter.DeleteRole(ctx, aggregate.Id)
			if err != nil {
				return fmt.Errorf("failed to delete role in database: %w", err)
			}

			return nil
		})

	if err != nil {
		slog.Error("Error deleting role", "error", err)
		return r.Response[any]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error deleting role",
		}, err
	}

	return r.Response[any]{
		Status: ubstatus.Success,
	}, nil
}

func (m *ManagementImpl) RoleUndelete(ctx context.Context,
	command RoleUndeleteCommand,
	agent string) (r.Response[any], error) {

	err := m.store.WithContext(
		ctx,
		func(etx evercore.EventStoreContext) error {
			aggregate := RoleAggregate{}
			err := etx.LoadStateInto(&aggregate, command.Id)
			if err != nil {
				return fmt.Errorf("failed to load role: %w", err)
			}

			event := evercore.NewStateEvent(RoleUndeletedEvent{})

			err = etx.ApplyEventTo(&aggregate, event, time.Now(), agent)
			if err != nil {
				return fmt.Errorf("failed to apply role undeleted event: %w", err)
			}

			err = m.dbadapter.AddRole(ctx, aggregate.Id, aggregate.State.OrganizationId, aggregate.State.Name, aggregate.State.SystemName)
			if err != nil {
				return fmt.Errorf("failed to undelete role in database: %w", err)
			}

			return nil
		})

	if err != nil {
		slog.Error("Error undeleting role", "error", err)
		return r.Response[any]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error undeleting role",
		}, err
	}

	return r.Response[any]{
		Status: ubstatus.Success,
	}, nil
}

func (m *ManagementImpl) RolePermissionAdd(ctx context.Context,
	command RolePermissionAddCommand,
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
			aggregate := RoleAggregate{}
			err := etx.LoadStateInto(&aggregate, command.Id)
			if err != nil {
				return fmt.Errorf("failed to load role: %w", err)
			}

			event := RolePermissionAddedEvent{
				Permission: command.Permission,
			}

			err = etx.ApplyEventTo(&aggregate, event, time.Now(), agent)
			if err != nil {
				return fmt.Errorf("failed to apply role permission added event: %w", err)
			}

			err = m.dbadapter.AddPermissionToRole(ctx, command.Id, command.Permission)
			if err != nil {
				return fmt.Errorf("failed to add permission to role in database: %w", err)
			}

			return nil
		})

	if err != nil {
		slog.Error("Error adding permission to role", "error", err)
		return r.Response[any]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error adding permission to role",
		}, err
	}

	return r.Response[any]{
		Status: ubstatus.Success,
	}, nil
}

func (m *ManagementImpl) RolePermissionRemove(ctx context.Context,
	command RolePermissionRemoveCommand,
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
			aggregate := RoleAggregate{}
			err := etx.LoadStateInto(&aggregate, command.Id)
			if err != nil {
				return fmt.Errorf("failed to load role: %w", err)
			}

			event := RolePermissionRemovedEvent{
				Permission: command.Permission,
			}

			err = etx.ApplyEventTo(&aggregate, event, time.Now(), agent)
			if err != nil {
				return fmt.Errorf("failed to apply role permission removed event: %w", err)
			}

			err = m.dbadapter.RemovePermissionFromRole(ctx, command.Id, command.Permission)
			if err != nil {
				return fmt.Errorf("failed to remove permission from role in database: %w", err)
			}

			return nil
		})

	if err != nil {
		slog.Error("Error removing permission from role", "error", err)
		return r.Response[any]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error removing permission from role",
		}, err
	}

	return r.Response[any]{
		Status: ubstatus.Success,
	}, nil
}

func (m *ManagementImpl) RoleGetBySystemName(ctx context.Context,
	systemName string) (r.Response[RoleAggregate], error) {

	aggregate, err := evercore.InContext(
		ctx,
		m.store,
		func(etx evercore.EventStoreContext) (*RoleAggregate, error) {
			aggregate := RoleAggregate{}
			err := etx.LoadStateByKeyInto(&aggregate, systemName)
			if err != nil {
				return nil, fmt.Errorf("failed to load role by system name: %w", err)
			}
			return &aggregate, nil
		})

	if err != nil {
		status := MapEvercoreErrorToStatus(err)
		slog.Error("Error getting role by system name", "error", err)
		return r.Response[RoleAggregate]{
			Status:  status,
			Message: "Error getting role",
		}, err
	}

	return r.Response[RoleAggregate]{
		Status: ubstatus.Success,
		Data:   *aggregate,
	}, nil
}

func (m *ManagementImpl) RoleGetById(ctx context.Context,
	roleId int64) (r.Response[RoleAggregate], error) {

	aggregate, err := evercore.InReadonlyContext(
		ctx,
		m.store,
		func(etx evercore.EventStoreReadonlyContext) (*RoleAggregate, error) {
			aggregate := RoleAggregate{}
			err := etx.LoadStateInto(&aggregate, roleId)
			if err != nil {
				return nil, fmt.Errorf("failed to load role: %w", err)
			}
			return &aggregate, nil
		})

	if err != nil {
		status := MapEvercoreErrorToStatus(err)
		slog.Error("Error getting role by ID", "error", err)
		return r.Response[RoleAggregate]{
			Status:  status,
			Message: "Error getting role",
		}, err
	}

	return r.Response[RoleAggregate]{
		Status: ubstatus.Success,
		Data:   *aggregate,
	}, nil
}
