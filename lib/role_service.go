package ubase

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"time"

	evercore "github.com/kernelplex/evercore/base"
	data "github.com/kernelplex/ubase/lib/ubdata"
	"github.com/kernelplex/ubase/lib/ubevents"
)

type RoleService interface {
	AddRole(ctx context.Context, name string, agent string) (int64, error)
	AddPermissionToRole(ctx context.Context, role string, permissionId int64, agent string) error
	RemovePermissionFromRole(ctx context.Context, role string, permissionId int64, agent string) error

	GetRoleList(ctx context.Context) (map[string]int64, error)
}

type RoleServiceImpl struct {
	store *evercore.EventStore
	db    *sql.DB
}

func CreateRoleService(store *evercore.EventStore, db *sql.DB) RoleService {
	service := RoleServiceImpl{
		store: store,
		db:    db,
	}
	return service
}

func (s RoleServiceImpl) AddRole(ctx context.Context, name string, agent string) (int64, error) {
	aggregate := NewRoleAggregate()
	aggregateId, err := evercore.InContext(
		ctx,
		s.store,
		func(etx evercore.EventStoreContext) (int64, error) {
			orm := data.New(s.db)
			err := etx.CreateAggregateWithKeyInto(&aggregate, name)
			if err != nil {
				return 0, err
			}

			etx.CreateAggregateWithKeyInto(&aggregate, name)
			etx.ApplyEventTo(&aggregate, evercore.NewStateEvent(ubevents.RoleCreatedEvent{Name: name}), time.Now(), agent)

			roleParams := data.AddRoleParams{
				RoleID: aggregate.Id,
				Name:   name,
			}
			orm.AddRole(ctx, roleParams)

			return aggregate.Id, nil
		})
	if err != nil {
		return 0, err
	}

	return aggregateId, nil
}

func (s RoleServiceImpl) AddPermissionToRole(ctx context.Context, role string, permissionId int64, agent string) error {
	aggregate := NewRoleAggregate()
	_, err := evercore.InContext(
		ctx,
		s.store,
		func(etx evercore.EventStoreContext) (int64, error) {
			// Load existing role aggregate
			err := etx.LoadStateByKeyInto(&aggregate, role)
			if err != nil {
				return 0, fmt.Errorf("failed to load role: %w", err)
			}

			// Check if permission already exists
			if slices.Contains(aggregate.State.Premissions, permissionId) {
				return 0, fmt.Errorf("permission %d already exists for role %s", permissionId, role)
			}

			// Apply permission added event
			etx.ApplyEventTo(&aggregate,
				evercore.NewStateEvent(ubevents.RolePermissionAddedEvent{
					PermissionId: permissionId,
				}),
				time.Now(),
				agent)

			// Update database
			orm := data.New(s.db)
			params := data.AddPermissionToRoleParams{
				RoleID:       aggregate.Id,
				PermissionID: permissionId,
			}
			err = orm.AddPermissionToRole(ctx, params)
			if err != nil {
				return 0, fmt.Errorf("failed to add permission to role in database: %w", err)
			}

			// TODO: Allow function subscriptions to updates.
			/*
				s.eventBus.Publish(
					domain_events.RolePermissionAddedEventType,
					domain_events.RolePermissionAdded{
						RoleId:     aggregate.Id,
						Permission: permissionId,
					})
			*/

			return aggregate.Id, nil
		})
	return err
}

func (s RoleServiceImpl) RemovePermissionFromRole(ctx context.Context, role string, permissionId int64, agent string) error {
	aggregate := NewRoleAggregate()
	_, err := evercore.InContext(
		ctx,
		s.store,
		func(etx evercore.EventStoreContext) (int64, error) {
			// Load existing role aggregate
			err := etx.LoadStateByKeyInto(&aggregate, role)
			if err != nil {
				return 0, fmt.Errorf("failed to load role: %w", err)
			}

			// Check if permission exists
			if !slices.Contains(aggregate.State.Premissions, permissionId) {
				return 0, fmt.Errorf("permission %d not found for role %s", permissionId, role)
			}

			// Apply permission removed event
			etx.ApplyEventTo(&aggregate,
				evercore.NewStateEvent(ubevents.RolePermissionRemovedEvent{
					PermissionId: permissionId,
				}),
				time.Now(),
				agent)

			// Update database
			orm := data.New(s.db)
			params := data.RemovePermissionFromRoleParams{
				RoleID:       aggregate.Id,
				PermissionID: permissionId,
			}
			err = orm.RemovePermissionFromRole(ctx, params)
			if err != nil {
				return 0, fmt.Errorf("failed to remove permission from role in database: %w", err)
			}

			// TODO: Allow function subscriptions to updates.
			/*
				s.eventBus.Publish(
					domain_events.RolePermissionRemovedEventType,
					domain_events.RolePermissionRemoved{
						RoleId:     aggregate.Id,
						Permission: permissionId,
					})
			*/
			return aggregate.Id, nil
		})
	return err
}

func (s RoleServiceImpl) GetRoleList(ctx context.Context) (map[string]int64, error) {
	orm := data.New(s.db)
	roles, err := orm.GetRoles(ctx)
	if err != nil {
		return nil, err
	}
	roleMap := make(map[string]int64)
	for _, role := range roles {
		roleMap[role.Name] = role.RoleID
	}
	return roleMap, nil
}
