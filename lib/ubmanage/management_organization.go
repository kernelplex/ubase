package ubmanage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	evercore "github.com/kernelplex/evercore/base"
	"github.com/kernelplex/ubase/lib/ubdata"
)

func (m *ManagementImpl) OrganizationList(ctx context.Context) (Response[[]ubdata.Organization], error) {
	organizations, err := m.dbadapter.ListOrganizations(ctx)
	if err != nil {
		return Error[[]ubdata.Organization]("Error listing organizations"), err
	}
	result := make([]ubdata.Organization, len(organizations))
	for i, o := range organizations {
		result[i] = ubdata.Organization(o)
	}
	return Success(result), nil
}

func (m *ManagementImpl) OrganizationAdd(ctx context.Context,
	command OrganizationCreateCommand,
	agent string) (Response[IdValue], error) {

	// Validation
	ok, issues := command.Validate()
	if !ok {
		return ValidationError[IdValue](issues), nil
	}

	id, err := evercore.InContext(
		ctx,
		m.store,
		func(etx evercore.EventStoreContext) (int64, error) {
			aggregate := OrganizationAggregate{}
			err := etx.CreateAggregateWithKeyInto(&aggregate, command.SystemName)
			if err != nil {
				return 0, fmt.Errorf("failed to create aggregate: %w", err)
			}
			event := evercore.NewStateEvent(OrganizationAddedEvent{
				Name:       command.Name,
				SystemName: command.SystemName,
				Status:     command.Status,
			})
			err = etx.ApplyEventTo(&aggregate, event, time.Now(), agent)
			if err != nil {
				return 0, fmt.Errorf("failed to apply organization added event: %w", err)
			}

			err = m.dbadapter.AddOrganization(ctx, aggregate.Id, aggregate.State.Name, aggregate.State.SystemName, aggregate.State.Status)

			if err != nil {
				return 0, fmt.Errorf("failed to add organization in database: %w", err)
			}

			return aggregate.Id, nil
		})

	if err != nil {
		slog.Error("Error creating organization", "error", err)
		return Error[IdValue]("Error creating organization"), err
	}

	return Success(IdValue{
		Id: id,
	}), nil
}

func (m *ManagementImpl) OrganizationUpdate(ctx context.Context,
	command OrganizationUpdateCommand,
	agent string) (Response[any], error) {

	// Validation
	ok, issues := command.Validate()
	if !ok {
		return ValidationError[any](issues), nil
	}

	err := m.store.WithContext(
		ctx,
		func(etx evercore.EventStoreContext) error {
			aggregate := OrganizationAggregate{}
			err := etx.LoadStateInto(&aggregate, command.Id)
			if err != nil {
				return fmt.Errorf("failed to load organization by system name: %w", err)
			}

			// Save the previous system name so we can check if it has changed
			previousSystemName := aggregate.State.SystemName

			event := evercore.NewStateEvent(OrganizationUpdatedEvent{
				Id:         aggregate.Id,
				Name:       command.Name,
				SystemName: command.SystemName,
				Status:     command.Status,
			})

			err = etx.ApplyEventTo(&aggregate, event, time.Now(), agent)
			if err != nil {
				return fmt.Errorf("failed to apply organization updated event: %w", err)
			}

			// Be sure to update the natural key if it has changed
			if aggregate.State.SystemName != previousSystemName {
				err = etx.ChangeAggregateNaturalKey(aggregate.Id, *command.SystemName)
				if err != nil {
					return fmt.Errorf("failed to change organization natural key: %w", err)
				}
			}

			err = m.dbadapter.UpdateOrganization(ctx, aggregate.Id, aggregate.State.Name, aggregate.State.SystemName, aggregate.State.Status)

			if err != nil {
				return fmt.Errorf("failed to update organization in database: %w", err)
			}

			return nil
		})

	if err != nil {
		slog.Error("Error updating organization", "error", err)
		return Error[any]("Error updating organization"), err
	}

	return SuccessAny(), nil
}

func (m *ManagementImpl) OrganizationGetBySystemName(
	ctx context.Context,
	systemName string) (Response[OrganizationAggregate], error) {

	aggregate, err := evercore.InContext(
		ctx,
		m.store,
		func(etx evercore.EventStoreContext) (*OrganizationAggregate, error) {
			aggregate := OrganizationAggregate{}
			err := etx.LoadStateByKeyInto(&aggregate, systemName)
			if err != nil {
				return nil, fmt.Errorf("failed to load organization by system name: %w", err)
			}

			return &aggregate, nil
		})

	if err != nil {
		return Error[OrganizationAggregate]("Error getting organization"), err
	}

	return Success(*aggregate), nil
}
