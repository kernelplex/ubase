package ubase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/kernelplex/evercore/base"
	"github.com/kernelplex/ubase/lib/ubdata"
	"github.com/kernelplex/ubase/lib/ubevents"
	"github.com/kernelplex/ubase/lib/ubsecurity"
	"github.com/kernelplex/ubase/lib/ubstatus"
)

const maxFailedLoginAttempts = 5
const maxLoginFiledAttemptsTimeout = 10 * time.Minute

type UserService interface {
	CreateUser(ctx context.Context, command UserCreateCommand, agent string) (UserCreateResponse, error)
	UpdateUser(ctx context.Context, command UserUpdateCommand, agent string) (UserUpdatedResponse, error)
	GetUserByEmail(ctx context.Context, email string) (*UserAggregate, error)
	Login(ctx context.Context, command UserLoginCommand) (*UserLoginResponse, error)
	SetRoles(ctx context.Context, command UserSetRolesComand, agent string) (UserSetRolesResponse, error)
	GetUserRolesIds(ctx context.Context, userId int64) ([]int64, error)
	// EmailVerifyGenerate(ctx context.Context, command UserEmailVerifyGenerateCommand, agent) (UserEmailVerifyGenerateResponse, error)
}

type UserServiceImpl struct {
	store          *evercore.EventStore
	hashingService ubsecurity.HashGenerator
	dbadapter      ubdata.DataAdapter
}

func CreateUserService(store *evercore.EventStore, hashingService ubsecurity.HashGenerator, dbadapter ubdata.DataAdapter) UserService {
	service := UserServiceImpl{
		store:          store,
		hashingService: hashingService,
		dbadapter:      dbadapter,
	}
	return service
}

func (s UserServiceImpl) GetUserByEmail(ctx context.Context, email string) (*UserAggregate, error) {

	aggregate, err := evercore.InReadonlyContext(
		ctx,
		s.store,
		func(etx evercore.EventStoreReadonlyContext) (*UserAggregate, error) {
			aggregate := UserAggregate{}
			etx.LoadStateByKeyInto(&aggregate, email)
			return &aggregate, nil
		})
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return aggregate, nil
}

func (s UserServiceImpl) CreateUser(ctx context.Context, command UserCreateCommand, agent string) (UserCreateResponse, error) {

	issues := command.Validate()
	if issues != nil {
		return UserCreateResponse{
			Response: Response{
				Status:           ubstatus.ValidationError,
				Message:          "Validation issues",
				ValidationIssues: issues,
			},
		}, nil
	}

	id, err := evercore.InContext(
		ctx,
		s.store,
		func(etx evercore.EventStoreContext) (int64, error) {
			aggregate := UserAggregate{}
			err := etx.CreateAggregateWithKeyInto(&aggregate, command.Email)
			if err != nil {
				return 0, fmt.Errorf("failed to create aggregate: %w", err)
			}

			slog.Info("Error creating user", "email", command.Email, "err", err)

			passwordHash, err := s.hashingService.GenerateHashBase64(command.Password)
			if err != nil {
				return 0, fmt.Errorf("failed to generate password hash: %w", err)
			}

			stateEvent := evercore.NewStateEvent(
				ubevents.UserCreatedEvent{
					Email:        &command.Email,
					PasswordHash: &passwordHash,
					FirstName:    &command.FirstName,
					LastName:     &command.LastName,
					DisplayName:  &command.DisplayName,
				})

			currentTime := time.Now()
			etx.ApplyEventTo(&aggregate, stateEvent, currentTime, agent)

			err = s.ProjectUser(ctx, &aggregate)
			if err != nil {
				return 0, fmt.Errorf("failed to project user: %w", err)
			}

			return aggregate.Id, nil
		})

	if err != nil {
		slog.Error("Error creating user", "error", err)
		return UserCreateResponse{
			Response: Response{
				Status:  ubstatus.UnexpectedError,
				Message: "Error creating user",
			},
		}, err
	}

	return UserCreateResponse{
		Response: Response{
			Status:  ubstatus.Success,
			Message: "User created successfully",
		},
		Id: id,
	}, nil
}

func (s UserServiceImpl) UpdateUser(ctx context.Context, command UserUpdateCommand, agent string) (UserUpdatedResponse, error) {
	issues := command.Validate()
	if issues != nil {
		return UserUpdatedResponse{
			Response: Response{
				Status:           ubstatus.ValidationError,
				Message:          "Validation issues",
				ValidationIssues: issues,
			},
		}, nil
	}

	err := s.store.WithContext(
		ctx,
		func(etx evercore.EventStoreContext) error {
			aggregate := UserAggregate{}
			err := etx.LoadStateInto(&aggregate, command.Id)
			if err != nil {
				slog.Error("Failed to load user", "error", err)
				return fmt.Errorf("failed to load user: %w", err)
			}

			var passwordHash *string
			if command.Password != nil {
				hash, err := s.hashingService.GenerateHashBase64(*command.Password)
				if err != nil {
					return fmt.Errorf("failed to generate password hash: %w", err)
				}
				passwordHash = &hash
			}

			stateEvent := evercore.NewStateEvent(
				ubevents.UserUpdatedEvent{
					PasswordHash: passwordHash,
					FirstName:    command.FirstName,
					LastName:     command.LastName,
					DisplayName:  command.DisplayName,
				})

			currentTime := time.Now()
			etx.ApplyEventTo(&aggregate, stateEvent, currentTime, agent)

			// Update the projected data
			err = s.ProjectUser(ctx, &aggregate)
			if err != nil {
				slog.Error("Failed to project user", "error", err)
				return fmt.Errorf("failed to project user: %w", err)
			}

			return nil
		})

	if err != nil {
		slog.Error("Error updating user", "error", err)
		return UserUpdatedResponse{
			Response: Response{
				Status:  ubstatus.UnexpectedError,
				Message: "Error updating user",
			},
		}, err
	}

	return UserUpdatedResponse{
		Response: Response{
			Status:  ubstatus.Success,
			Message: "User updated successfully",
		},
	}, nil
}

func (s UserServiceImpl) handleMaxAttempts(etx evercore.EventStoreContext, aggregate *UserAggregate) *UserLoginResponse {
	timeSinceLastAttempt := time.Since(time.Unix(aggregate.State.LastLoginAttempt, 0))
	if timeSinceLastAttempt < maxLoginFiledAttemptsTimeout {
		etx.ApplyEventTo(aggregate, evercore.NewStateEvent(ubevents.UserLoginFailedEvent{
			LastLoginAttempt:    time.Now().Unix(),
			FailedLoginAttempts: aggregate.State.FailedLoginAttempts + 1,
		}), time.Now(), "")
		return &UserLoginResponse{
			Response: Response{
				Status:  ubstatus.NotAuthorized,
				Message: "Max failed login attempts exceeded. Please try again later",
			},
		}
	}
	return nil
}

func (s UserServiceImpl) verifyPassword(etx evercore.EventStoreContext, aggregate *UserAggregate, password string) (*UserLoginResponse, error) {
	valid, err := s.hashingService.VerifyBase64(password, aggregate.State.PasswordHash)
	if err != nil || !valid {
		etx.ApplyEventTo(aggregate, evercore.NewStateEvent(ubevents.UserLoginFailedEvent{
			LastLoginAttempt:    time.Now().Unix(),
			FailedLoginAttempts: aggregate.State.FailedLoginAttempts + 1,
		}), time.Now(), "")
		return &UserLoginResponse{
			Response: Response{
				Status:  ubstatus.NotAuthorized,
				Message: "Invalid email or password",
			},
		}, nil
	}
	return nil, nil
}

func (s UserServiceImpl) createSuccessResponse(etx evercore.EventStoreContext, aggregate *UserAggregate) *UserLoginResponse {
	etx.ApplyEventTo(aggregate, evercore.NewStateEvent(ubevents.UserLoginSucceededEvent{
		LastLoginAttempt:    time.Now().Unix(),
		FailedLoginAttempts: 0,
	}), time.Now(), "")

	return &UserLoginResponse{
		Response: Response{
			Status:  ubstatus.Success,
			Message: "User login successful",
		},
		UserId:      aggregate.Id,
		LastName:    &aggregate.State.LastName,
		FirstName:   &aggregate.State.FirstName,
		DisplayName: &aggregate.State.DisplayName,
		Email:       &aggregate.State.Email,
	}
}

func (s UserServiceImpl) Login(ctx context.Context, command UserLoginCommand) (*UserLoginResponse, error) {
	response, err := evercore.InContext(
		ctx,
		s.store,
		func(etx evercore.EventStoreContext) (*UserLoginResponse, error) {
			aggregate := UserAggregate{}
			etx.LoadStateByKeyInto(&aggregate, command.Email)

			if aggregate.State.FailedLoginAttempts >= maxFailedLoginAttempts {
				if resp := s.handleMaxAttempts(etx, &aggregate); resp != nil {
					return resp, nil
				}
			}

			if resp, err := s.verifyPassword(etx, &aggregate, command.Password); resp != nil || err != nil {
				return resp, fmt.Errorf("failed to verify password: %w", err)

			}

			return s.createSuccessResponse(etx, &aggregate), nil
		})

	if err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}
	return response, nil

}

func (s UserServiceImpl) SetRoles(ctx context.Context, command UserSetRolesComand, agent string) (UserSetRolesResponse, error) {
	response := UserSetRolesResponse{
		Response: Response{
			Status:  ubstatus.Success,
			Message: "Roles updated successfully",
		},
	}

	err := s.store.WithContext(
		ctx,
		func(etx evercore.EventStoreContext) error {
			// Load user aggregate
			aggregate := UserAggregate{}
			err := etx.LoadStateInto(&aggregate, command.Id)
			if err != nil {
				return fmt.Errorf("failed to load user: %w", err)
			}

			event := evercore.NewStateEvent(ubevents.UserRolesUpdatedEvent{
				Roles: command.RoleIds,
			})
			currentTime := time.Now()
			etx.ApplyEventTo(&aggregate, event, currentTime, agent)

			err = s.ProjectRoles(ctx, &aggregate)
			if err != nil {
				return fmt.Errorf("failed to update roles on user: %w", err)
			}

			// TODO:Implement callback subscriptions
			/*
				s.eventBus.Publish(
					domain_events.UserRolesUpdatedEventType,
					domain_events.UserRolesUpdated{
						UserId: command.Id,
					})
			*/
			return nil
		})

	if err != nil {
		slog.Error("failed to set roles", "userId", command.Id, "error", err)
		response.Response.Status = ubstatus.UnexpectedError
		response.Response.Message = "Failed to update roles"
		return response, fmt.Errorf("failed to set roles: %w", err)
	}

	return response, nil
}

func (s UserServiceImpl) GetUserRolesIds(ctx context.Context, userId int64) ([]int64, error) {
	roles, err := s.dbadapter.GetUserRoles(ctx, userId)
	if err != nil {
		slog.Error("failed to get user roles", "userId", userId, "error", err)
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	roleIds := make([]int64, 0, len(roles))
	for _, role := range roles {
		roleIds = append(roleIds, role.RoleID)
	}

	return roleIds, nil
}

func (s UserServiceImpl) ProjectRoles(ctx context.Context, userAggregate *UserAggregate) error {
	return s.dbadapter.ProjectUserRoles(ctx, userAggregate.Id, userAggregate.State.Roles)
}

// ProjectUser projects the user to the relational database.
func (s UserServiceImpl) ProjectUser(ctx context.Context, userAggregate *UserAggregate) error {
	return s.dbadapter.ProjectUser(ctx, userAggregate.Id, userAggregate.State)

}
