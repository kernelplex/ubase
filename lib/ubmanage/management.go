package ubmanage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	evercore "github.com/kernelplex/evercore/base"
	"github.com/kernelplex/ubase/lib/ubdata"
	"github.com/kernelplex/ubase/lib/ubsecurity"
	"github.com/kernelplex/ubase/lib/ubstatus"
)

type IdValue struct {
	Id int64 `json:"id"`
}

type ManagementService interface {

	// Oganization operations

	OrganizationAdd(ctx context.Context,
		command OrganizationCreateCommand,
		agent string) (Response[IdValue], error)
	OrganizationGetBySystemName(
		ctx context.Context,
		systemName string) (Response[OrganizationAggregate], error)
	OrganizationUpdate(ctx context.Context,
		command OrganizationUpdateCommand,
		agent string) (Response[any], error)

	RoleAdd(ctx context.Context,
		command RoleCreateCommand,
		agent string) (Response[IdValue], error)

	RoleUpdate(ctx context.Context,
		command RoleUpdateCommand,
		agent string) (Response[any], error)

	RoleGetById(ctx context.Context,
		roleId int64) (Response[RoleAggregate], error)

	RoleGetBySystemName(ctx context.Context,
		systemName string) (Response[RoleAggregate], error)

	RoleDelete(ctx context.Context,
		command RoleDeleteCommand,
		agent string) (Response[any], error)

	RoleUndelete(ctx context.Context,
		command RoleUndeleteCommand,
		agent string) (Response[any], error)

	RolePermissionAdd(ctx context.Context,
		command RolePermissionAddCommand,
		agent string) (Response[any], error)

	RolePermissionRemove(ctx context.Context,
		command RolePermissionRemoveCommand,
		agent string) (Response[any], error)

	// User operations
	UserAdd(ctx context.Context,
		command UserCreateCommand,
		agent string) (Response[IdValue], error)

	UserGetByEmail(ctx context.Context,
		email string) (Response[UserAggregate], error)

	UserUpdate(ctx context.Context,
		command UserUpdateCommand,
		agent string) (Response[any], error)

	UserAuthenticate(ctx context.Context,
		command UserLoginCommand,
		agent string) (Response[any], error)

	UserVerifyTwoFactorCode(ctx context.Context,
		command UserVerifyTwoFactorLoginCommand,
		agent string) (Response[any], error)

	UserGenerateVerificationToken(ctx context.Context,
		command UserGenerateVerificationTokenCommand,
		agent string) (Response[any], error)

	UserVerify(ctx context.Context,
		command UserVerifyCommand,
		agent string) (Response[any], error)

	UserGenerateTwoFactorSharedSecret(ctx context.Context,
		command UserGenerateTwoFactorSharedSecretCommand,
		agent string) (Response[any], error)

	UserDisable(ctx context.Context,
		command UserDisableCommand,
		agent string) (Response[any], error)

	UserEnable(ctx context.Context,
		command UserEnableCommand,
		agent string) (Response[any], error)
}

type ManagementImpl struct {
	store             *evercore.EventStore
	dbadapter         ubdata.DataAdapter
	hashingService    ubsecurity.HashGenerator
	encryptionService ubsecurity.EncryptionService
}

func NewManagement(
	store *evercore.EventStore,
	dbadapter ubdata.DataAdapter,
	hashingService ubsecurity.HashGenerator,
	encryptionService ubsecurity.EncryptionService,
) ManagementService {
	management := ManagementImpl{
		store:             store,
		dbadapter:         dbadapter,
		hashingService:    hashingService,
		encryptionService: encryptionService,
	}
	return &management
}

func (m *ManagementImpl) OrganizationAdd(ctx context.Context,
	command OrganizationCreateCommand,
	agent string) (Response[IdValue], error) {

	// Validation
	ok, issues := command.Validate()
	if !ok {
		return Response[IdValue]{
			Status:           ubstatus.ValidationError,
			Message:          "Validation issues",
			ValidationIssues: issues,
		}, nil
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
		return Response[IdValue]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error creating organization",
		}, err
	}

	return Response[IdValue]{
		Status: ubstatus.Success,
		Data: IdValue{
			Id: id,
		},
	}, nil
}

func (m *ManagementImpl) OrganizationUpdate(ctx context.Context,
	command OrganizationUpdateCommand,
	agent string) (Response[any], error) {

	// Validation
	ok, issues := command.Validate()
	if !ok {
		return Response[any]{
			Status:           ubstatus.ValidationError,
			Message:          "Validation issues",
			ValidationIssues: issues,
		}, nil
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
		return Response[any]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error updating organization",
		}, err
	}

	return Response[any]{
		Status: ubstatus.Success,
	}, nil

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
		return Response[OrganizationAggregate]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error getting organization",
		}, err
	}
	return Response[OrganizationAggregate]{
		Status: ubstatus.Success,
		Data:   *aggregate,
	}, nil

}

func (m *ManagementImpl) RoleAdd(ctx context.Context,
	command RoleCreateCommand,
	agent string) (Response[IdValue], error) {

	// Validation
	ok, issues := command.Validate()
	if !ok {
		return Response[IdValue]{
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
		slog.Error("Error creating role", "error", err)
		return Response[IdValue]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error creating role",
		}, err
	}

	return Response[IdValue]{
		Status: ubstatus.Success,
		Data: IdValue{
			Id: id,
		},
	}, nil
}

func (m *ManagementImpl) RoleUpdate(ctx context.Context,
	command RoleUpdateCommand,
	agent string) (Response[any], error) {

	// Validation
	ok, issues := command.Validate()
	if !ok {
		return Response[any]{
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
		return Response[any]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error updating role",
		}, err
	}

	return Response[any]{
		Status: ubstatus.Success,
	}, nil
}

func (m *ManagementImpl) RoleDelete(ctx context.Context,
	command RoleDeleteCommand,
	agent string) (Response[any], error) {

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
		return Response[any]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error deleting role",
		}, err
	}

	return Response[any]{
		Status: ubstatus.Success,
	}, nil
}

func (m *ManagementImpl) RoleUndelete(ctx context.Context,
	command RoleUndeleteCommand,
	agent string) (Response[any], error) {

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
		return Response[any]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error undeleting role",
		}, err
	}

	return Response[any]{
		Status: ubstatus.Success,
	}, nil
}

func (m *ManagementImpl) RolePermissionAdd(ctx context.Context,
	command RolePermissionAddCommand,
	agent string) (Response[any], error) {

	// Validation
	ok, issues := command.Validate()
	if !ok {
		return Response[any]{
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
		return Response[any]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error adding permission to role",
		}, err
	}

	return Response[any]{
		Status: ubstatus.Success,
	}, nil
}

func (m *ManagementImpl) RolePermissionRemove(ctx context.Context,
	command RolePermissionRemoveCommand,
	agent string) (Response[any], error) {

	// Validation
	ok, issues := command.Validate()
	if !ok {
		return Response[any]{
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
		return Response[any]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error removing permission from role",
		}, err
	}

	return Response[any]{
		Status: ubstatus.Success,
	}, nil
}

func (m *ManagementImpl) RoleGetBySystemName(ctx context.Context,
	systemName string) (Response[RoleAggregate], error) {

	aggregate, err := evercore.InReadonlyContext(
		ctx,
		m.store,
		func(etx evercore.EventStoreReadonlyContext) (*RoleAggregate, error) {
			aggregate := RoleAggregate{}
			err := etx.LoadStateByKeyInto(&aggregate, systemName)
			if err != nil {
				return nil, fmt.Errorf("failed to load role by system name: %w", err)
			}
			return &aggregate, nil
		})

	if err != nil {
		slog.Error("Error getting role by system name", "error", err)
		return Response[RoleAggregate]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error getting role",
		}, err
	}

	return Response[RoleAggregate]{
		Status: ubstatus.Success,
		Data:   *aggregate,
	}, nil
}

func (m *ManagementImpl) RoleGetById(ctx context.Context,
	roleId int64) (Response[RoleAggregate], error) {

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
		slog.Error("Error getting role by ID", "error", err)
		return Response[RoleAggregate]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error getting role",
		}, err
	}

	return Response[RoleAggregate]{
		Status: ubstatus.Success,
		Data:   *aggregate,
	}, nil
}

func (m *ManagementImpl) UserAdd(ctx context.Context,
	command UserCreateCommand,
	agent string) (Response[IdValue], error) {

	if ok, issues := command.Validate(); !ok {
		return Response[IdValue]{
			Status:           ubstatus.ValidationError,
			Message:          "Validation issues",
			ValidationIssues: issues,
		}, nil
	}

	id, err := evercore.InContext(
		ctx,
		m.store,
		func(etx evercore.EventStoreContext) (int64, error) {
			aggregate := UserAggregate{}
			err := etx.CreateAggregateWithKeyInto(&aggregate, command.Email)
			if err != nil {
				return 0, fmt.Errorf("failed to create aggregate: %w", err)
			}

			passwordHash, err := m.hashingService.GenerateHashBase64(command.Password)
			if err != nil {
				return 0, fmt.Errorf("failed to generate password hash: %w", err)
			}

			stateEvent := evercore.NewStateEvent(
				UserAddedEvent{
					Email:        command.Email,
					PasswordHash: passwordHash,
					FirstName:    command.FirstName,
					LastName:     command.LastName,
					DisplayName:  command.DisplayName,
				})

			currentTime := time.Now()
			etx.ApplyEventTo(&aggregate, stateEvent, currentTime, agent)

			err = m.dbadapter.AddUser(
				ctx, aggregate.Id,
				aggregate.State.FirstName,
				aggregate.State.LastName,
				aggregate.State.DisplayName,
				aggregate.State.Email)
			if err != nil {
				return 0, fmt.Errorf("failed to add user in database: %w", err)
			}
			return aggregate.Id, nil
		})

	if err != nil {
		slog.Error("Error creating user", "error", err)
		return Response[IdValue]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error creating user",
		}, err
	}
	return Response[IdValue]{
		Status: ubstatus.Success,
		Data: IdValue{
			Id: id,
		},
	}, nil
}

func (m *ManagementImpl) UserGetByEmail(ctx context.Context,
	email string) (Response[UserAggregate], error) {

	aggregate, err := evercore.InReadonlyContext(
		ctx,
		m.store,
		func(etx evercore.EventStoreReadonlyContext) (*UserAggregate, error) {
			aggregate := UserAggregate{}
			etx.LoadStateByKeyInto(&aggregate, email)
			return &aggregate, nil
		})
	if err != nil {
		return Response[UserAggregate]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error getting user",
		}, err
	}
	return Response[UserAggregate]{
		Status: ubstatus.Success,
		Data:   *aggregate,
	}, nil
}

func (m *ManagementImpl) UserUpdate(ctx context.Context,
	command UserUpdateCommand,
	agent string) (Response[any], error) {

	// Validation
	ok, issues := command.Validate()
	if !ok {
		return Response[any]{
			Status:           ubstatus.ValidationError,
			Message:          "Validation issues",
			ValidationIssues: issues,
		}, nil
	}

	err := m.store.WithContext(
		ctx,
		func(etx evercore.EventStoreContext) error {
			aggregate := UserAggregate{}
			err := etx.LoadStateInto(&aggregate, command.Id)
			if err != nil {
				return fmt.Errorf("failed to load user: %w", err)
			}

			// Save the previous email so we can check if it has changed
			previousEmail := aggregate.State.Email

			// Update password if provided
			var passwordHash *string = nil
			if command.Password != nil {
				hash, err := m.hashingService.GenerateHashBase64(*command.Password)
				if err != nil {
					return fmt.Errorf("failed to generate password hash: %w", err)
				}
				passwordHash = &hash
			}

			event := evercore.NewStateEvent(UserUpdatedEvent{
				Id:           command.Id,
				Email:        command.Email,
				FirstName:    command.FirstName,
				LastName:     command.LastName,
				DisplayName:  command.DisplayName,
				PasswordHash: passwordHash,
				Verified:     command.Verified,
			})

			err = etx.ApplyEventTo(&aggregate, event, time.Now(), agent)
			if err != nil {
				return fmt.Errorf("failed to apply user updated event: %w", err)
			}

			// Update natural key if email changed
			if command.Email != nil && aggregate.State.Email != previousEmail {
				err = etx.ChangeAggregateNaturalKey(aggregate.Id, *command.Email)
				if err != nil {
					return fmt.Errorf("failed to change user natural key: %w", err)
				}
			}

			err = m.dbadapter.UpdateUser(
				ctx,
				aggregate.Id,
				aggregate.State.FirstName,
				aggregate.State.LastName,
				aggregate.State.DisplayName,
				aggregate.State.Email)
			if err != nil {
				return fmt.Errorf("failed to update user in database: %w", err)
			}

			return nil
		})

	if err != nil {
		slog.Error("Error updating user", "error", err)
		return Response[any]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error updating user",
		}, err
	}

	return Response[any]{
		Status: ubstatus.Success,
	}, nil
}

func (m *ManagementImpl) UserAuthenticate(ctx context.Context,
	command UserLoginCommand,
	agent string) (Response[any], error) {
	panic("not implemented")
}

func (m *ManagementImpl) UserVerifyTwoFactorCode(ctx context.Context,
	command UserVerifyTwoFactorLoginCommand,
	agent string) (Response[any], error) {
	panic("not implemented")
}

func (m *ManagementImpl) UserGenerateVerificationToken(ctx context.Context,
	command UserGenerateVerificationTokenCommand,
	agent string) (Response[any], error) {
	panic("not implemented")
}

func (m *ManagementImpl) UserVerify(ctx context.Context,
	command UserVerifyCommand,
	agent string) (Response[any], error) {
	panic("not implemented")
}

func (m *ManagementImpl) UserGenerateTwoFactorSharedSecret(ctx context.Context,
	command UserGenerateTwoFactorSharedSecretCommand,
	agent string) (Response[any], error) {
	panic("not implemented")
}

func (m *ManagementImpl) UserDisable(ctx context.Context,
	command UserDisableCommand,
	agent string) (Response[any], error) {
	panic("not implemented")
}

func (m *ManagementImpl) UserEnable(ctx context.Context,
	command UserEnableCommand,
	agent string) (Response[any], error) {
	panic("not implemented")
}
