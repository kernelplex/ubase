package ubase

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/kernelplex/evercore/base"
	"github.com/kernelplex/ubase/lib/statuscode"
	data "github.com/kernelplex/ubase/lib/ubase_db"
	"github.com/kernelplex/ubase/lib/ubase_events"
	"github.com/kernelplex/ubase/lib/ubasesec"
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
}

type UserServiceImpl struct {
	store          *evercore.EventStore
	hashingService ubasesec.HashGenerator
	db             *sql.DB
}

func CreateUserService(store *evercore.EventStore, hashingService ubasesec.HashGenerator, db *sql.DB) UserService {
	service := UserServiceImpl{
		store:          store,
		hashingService: hashingService,
		db:             db,
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
		return nil, err
	}
	return aggregate, nil
}

func (s UserServiceImpl) CreateUser(ctx context.Context, command UserCreateCommand, agent string) (UserCreateResponse, error) {

	issues := command.Validate()
	if issues != nil {
		return UserCreateResponse{
			Response: Response{
				Status:           statuscode.ValidationError,
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
			etx.CreateAggregateWithKeyInto(&aggregate, command.Email)
			passwordHash, err := s.hashingService.GenerateHashBase64(command.Password)
			if err != nil {
				return 0, err
			}

			stateEvent := evercore.NewStateEvent(
				ubase_events.UserCreatedEvent{
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
				return 0, err
			}

			return aggregate.Id, nil
		})

	if err != nil {
		slog.Error("Error creating user", "error", err)
		return UserCreateResponse{
			Response: Response{
				Status:  statuscode.UnexpectedError,
				Message: "Error creating user",
			},
		}, err
	}

	return UserCreateResponse{
		Response: Response{
			Status:  statuscode.Success,
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
				Status:           statuscode.ValidationError,
				Message:          "Validation issues",
				ValidationIssues: issues,
			},
		}, nil
	}

	err := s.store.WithContext(
		ctx,
		func(etx evercore.EventStoreContext) error {
			aggregate := UserAggregate{}
			etx.LoadStateInto(&aggregate, command.Id)
			slog.Info("user id", "id", aggregate.Id)

			var passwordHash *string
			if command.Password != nil {
				hash, err := s.hashingService.GenerateHashBase64(*command.Password)
				if err != nil {
					return err
				}
				passwordHash = &hash
			}

			stateEvent := evercore.NewStateEvent(
				ubase_events.UserUpdatedEvent{
					PasswordHash: passwordHash,
					FirstName:    command.FirstName,
					LastName:     command.LastName,
					DisplayName:  command.DisplayName,
				})

			currentTime := time.Now()
			etx.ApplyEventTo(&aggregate, stateEvent, currentTime, agent)

			// Update the projected data
			err := s.ProjectUser(ctx, &aggregate)
			if err != nil {
				return err
			}

			return nil
		})

	if err != nil {
		slog.Error("Error updating user", "error", err)
		return UserUpdatedResponse{
			Response: Response{
				Status:  statuscode.UnexpectedError,
				Message: "Error updating user",
			},
		}, err
	}

	return UserUpdatedResponse{
		Response: Response{
			Status:  statuscode.Success,
			Message: "User updated successfully",
		},
	}, nil
}

func (s UserServiceImpl) handleMaxAttempts(etx evercore.EventStoreContext, aggregate *UserAggregate) *UserLoginResponse {
	timeSinceLastAttempt := time.Since(time.Unix(aggregate.State.LastLoginAttempt, 0))
	if timeSinceLastAttempt < maxLoginFiledAttemptsTimeout {
		etx.ApplyEventTo(aggregate, evercore.NewStateEvent(ubase_events.UserLoginFailedEvent{
			LastLoginAttempt:    time.Now().Unix(),
			FailedLoginAttempts: aggregate.State.FailedLoginAttempts + 1,
		}), time.Now(), "")
		return &UserLoginResponse{
			Response: Response{
				Status:  statuscode.NotAuthorized,
				Message: "Max failed login attempts exceeded. Please try again later",
			},
		}
	}
	return nil
}

func (s UserServiceImpl) verifyPassword(etx evercore.EventStoreContext, aggregate *UserAggregate, password string) (*UserLoginResponse, error) {
	valid, err := s.hashingService.VerifyBase64(password, aggregate.State.PasswordHash)
	if err != nil || !valid {
		etx.ApplyEventTo(aggregate, evercore.NewStateEvent(ubase_events.UserLoginFailedEvent{
			LastLoginAttempt:    time.Now().Unix(),
			FailedLoginAttempts: aggregate.State.FailedLoginAttempts + 1,
		}), time.Now(), "")
		return &UserLoginResponse{
			Response: Response{
				Status:  statuscode.NotAuthorized,
				Message: "Invalid email or password",
			},
		}, nil
	}
	return nil, nil
}

func (s UserServiceImpl) createSuccessResponse(etx evercore.EventStoreContext, aggregate *UserAggregate) *UserLoginResponse {
	etx.ApplyEventTo(aggregate, evercore.NewStateEvent(ubase_events.UserLoginSucceededEvent{
		LastLoginAttempt:    time.Now().Unix(),
		FailedLoginAttempts: 0,
	}), time.Now(), "")

	return &UserLoginResponse{
		Response: Response{
			Status:  statuscode.Success,
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
				return resp, err

			}

			return s.createSuccessResponse(etx, &aggregate), nil
		})

	if err != nil {
		return nil, err
	}
	return response, nil

}

func (s UserServiceImpl) SetRoles(ctx context.Context, command UserSetRolesComand, agent string) (UserSetRolesResponse, error) {
	response := UserSetRolesResponse{
		Response: Response{
			Status:  statuscode.Success,
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

			event := evercore.NewStateEvent(ubase_events.UserRolesUpdatedEvent{
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
		response.Response.Status = statuscode.UnexpectedError
		response.Response.Message = "Failed to update roles"
		return response, err
	}

	return response, nil
}

func (s UserServiceImpl) GetUserRolesIds(ctx context.Context, userId int64) ([]int64, error) {
	orm := data.New(s.db)
	roles, err := orm.GetUserRoles(ctx, userId)
	if err != nil {
		slog.Error("failed to get user roles", "userId", userId, "error", err)
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	var roleIds []int64
	for _, role := range roles {
		if role.RoleID.Valid {
			roleIds = append(roleIds, role.RoleID.Int64)
		}
	}

	return roleIds, nil
}

func (s UserServiceImpl) ProjectRoles(ctx context.Context, userAggregate *UserAggregate) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	orm := data.New(tx)
	return projectRoles(ctx, orm, userAggregate)

}

// ProjectUser projects the user to the relational database.
func (s UserServiceImpl) ProjectUser(ctx context.Context, userAggregate *UserAggregate) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	orm := data.New(tx)

	_, err = orm.GetUser(ctx, userAggregate.Id)
	// If the user doesn't exist, create it
	if err != nil {
		if err != sql.ErrNoRows {
			slog.Error("Failed to get user", "error", err)
			return err
		}
		_, err = orm.AddUser(ctx, data.AddUserParams{
			UserID:      userAggregate.Id,
			FirstName:   userAggregate.State.FirstName,
			LastName:    userAggregate.State.LastName,
			DisplayName: userAggregate.State.DisplayName,
			Email:       userAggregate.State.Email,
		})
		if err != nil {
			slog.Error("Failed to project user", "error", err)
			return err
		}
	} else {
		// If the user exists, update it
		err = orm.UpdateUser(ctx, data.UpdateUserParams{
			UserID:      userAggregate.Id,
			FirstName:   userAggregate.State.FirstName,
			LastName:    userAggregate.State.LastName,
			DisplayName: userAggregate.State.DisplayName,
			Email:       userAggregate.State.Email,
		})
		if err != nil {
			slog.Error("Failed to project user", "error", err)
			return err
		}
	}

	// Remove all existing roles
	err = projectRoles(ctx, orm, userAggregate)
	if err != nil {
		slog.Error("Failed to project roles", "error", err)
		return err
	}

	tx.Commit()

	return nil
}

func projectRoles(ctx context.Context, orm *data.Queries, userAggregate *UserAggregate) error {
	err := orm.RemoveAllRolesFromUser(ctx, userAggregate.Id)
	if err != nil {
		slog.Error("Failed to remove all roles from user", "error", err)
		return err
	}

	// Add roles
	for _, roleId := range userAggregate.State.Roles {
		err = orm.AddRoleToUser(ctx, data.AddRoleToUserParams{
			UserID: userAggregate.Id,
			RoleID: roleId,
		})
		if err != nil {
			slog.Error("Failed to add role to user", "error", err)
			return err
		}
	}
	return nil
}
