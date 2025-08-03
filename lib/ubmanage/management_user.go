package ubmanage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	evercore "github.com/kernelplex/evercore/base"
	r "github.com/kernelplex/ubase/lib/ubresponse"
	"github.com/kernelplex/ubase/lib/ubsecurity"
	"github.com/kernelplex/ubase/lib/ubstatus"
)

func (m *ManagementImpl) UserAdd(ctx context.Context,
	command UserCreateCommand,
	agent string) (r.Response[UserCreatedResponse], error) {

	if ok, issues := command.Validate(); !ok {
		return r.Response[UserCreatedResponse]{
			Status:           ubstatus.ValidationError,
			Message:          "Validation issues",
			ValidationIssues: issues,
		}, nil
	}

	type IdCode struct {
		Id   int64
		Code *string
	}

	result, err := evercore.InContext(
		ctx,
		m.store,
		func(etx evercore.EventStoreContext) (IdCode, error) {
			aggregate := UserAggregate{}
			err := etx.CreateAggregateWithKeyInto(&aggregate, command.Email)
			if err != nil {
				return IdCode{}, fmt.Errorf("failed to create aggregate: %w", err)
			}

			passwordHash, err := m.hashingService.GenerateHashBase64(command.Password)
			if err != nil {
				return IdCode{}, fmt.Errorf("failed to generate password hash: %w", err)
			}

			if command.GenerateVerificationToken {
				command.Verified = false
			}

			stateEvent := evercore.NewStateEvent(
				UserAddedEvent{
					Email:        command.Email,
					PasswordHash: passwordHash,
					FirstName:    command.FirstName,
					LastName:     command.LastName,
					DisplayName:  command.DisplayName,
					Verified:     command.Verified,
				})

			currentTime := time.Now()
			etx.ApplyEventTo(&aggregate, stateEvent, currentTime, agent)

			var token string
			if command.GenerateVerificationToken {
				token = ubsecurity.GenerateSecureRandomString(10)
				event := UserVerificationTokenGeneratedEvent{
					Token: token,
				}

				err = etx.ApplyEventTo(&aggregate, event, time.Now(), agent)
				if err != nil {
					return IdCode{}, fmt.Errorf("failed to apply user verification token generated event: %w", err)
				}
			}

			err = m.dbadapter.AddUser(
				ctx, aggregate.Id,
				aggregate.State.FirstName,
				aggregate.State.LastName,
				aggregate.State.DisplayName,
				aggregate.State.Email)
			if err != nil {
				return IdCode{}, fmt.Errorf("failed to add user in database: %w", err)
			}
			return IdCode{Id: aggregate.Id, Code: &token}, nil
		})

	if err != nil {
		slog.Error("Error creating user", "error", err)
		return r.Response[UserCreatedResponse]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error creating user",
		}, err
	}

	return r.Response[UserCreatedResponse]{
		Status: ubstatus.Success,
		Data: UserCreatedResponse{
			Id:                result.Id,
			VerificationToken: result.Code,
		},
	}, nil
}

func (m *ManagementImpl) UserGetById(ctx context.Context,
	userId int64) (r.Response[UserAggregate], error) {

	aggregate, err := evercore.InReadonlyContext(
		ctx,
		m.store,
		func(etx evercore.EventStoreReadonlyContext) (*UserAggregate, error) {
			aggregate := UserAggregate{}
			err := etx.LoadStateInto(&aggregate, userId)
			if err != nil {
				return nil, fmt.Errorf("failed to load user: %w", err)
			}
			return &aggregate, nil
		})
	if err != nil {
		return r.Response[UserAggregate]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error getting user",
		}, err
	}
	return r.Response[UserAggregate]{
		Status: ubstatus.Success,
		Data:   *aggregate,
	}, nil
}

func (m *ManagementImpl) UserGetByEmail(ctx context.Context,
	email string) (r.Response[UserAggregate], error) {

	aggregate, err := evercore.InReadonlyContext(
		ctx,
		m.store,
		func(etx evercore.EventStoreReadonlyContext) (*UserAggregate, error) {
			aggregate := UserAggregate{}
			etx.LoadStateByKeyInto(&aggregate, email)
			return &aggregate, nil
		})
	if err != nil {
		return r.Response[UserAggregate]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error getting user",
		}, err
	}
	return r.Response[UserAggregate]{
		Status: ubstatus.Success,
		Data:   *aggregate,
	}, nil
}

func (m *ManagementImpl) UserUpdate(ctx context.Context,
	command UserUpdateCommand,
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
		return r.Response[any]{
			Status:  ubstatus.UnexpectedError,
			Message: "Error updating user",
		}, err
	}

	return r.Response[any]{
		Status: ubstatus.Success,
	}, nil
}

type UserAuthenticationResponse struct {
	UserId               int64  `json:"user_id"`
	Email                string `json:"email"`
	RequiresTwoFactor    bool   `json:"requires_two_factor"`
	RequiresVerification bool   `json:"requires_verification"`
}

func (m *ManagementImpl) UserAuthenticate(ctx context.Context,
	command UserLoginCommand,
	agent string) (r.Response[*UserAuthenticationResponse], error) {

	return evercore.InReadonlyContext(
		ctx,
		m.store,
		func(etx evercore.EventStoreReadonlyContext) (r.Response[*UserAuthenticationResponse], error) {
			aggregate := UserAggregate{}
			err := etx.LoadStateByKeyInto(&aggregate, command.Email)
			if err != nil {
				slog.Error("Error getting user", "error", err)
				return r.Error[*UserAuthenticationResponse]("Email or password is incorrect"), err
			}

			match, err := m.hashingService.VerifyBase64(command.Password, aggregate.State.PasswordHash)
			if err != nil {
				slog.Error("Error verifying password", "error", err)
				return r.Error[*UserAuthenticationResponse]("Could not verify this account at this time."), err
			}

			if !match {
				slog.Error("Password does not match", "email", command.Email)
				return r.StatusError[*UserAuthenticationResponse](ubstatus.NotAuthorized, "Email or password is incorrect"), nil
			}

			if aggregate.State.Disabled {
				slog.Error("User is disabled", "email", command.Email)
				return r.StatusError[*UserAuthenticationResponse](ubstatus.NotAuthorized, "This account is not currently active. Please contact support."), nil
			}

			return r.Success(&UserAuthenticationResponse{
				UserId: aggregate.Id,
				Email:  aggregate.State.Email,
				RequiresTwoFactor: aggregate.State.TwoFactorSharedSecret != nil &&
					len(*aggregate.State.TwoFactorSharedSecret) > 0,
				RequiresVerification: aggregate.State.Verified == false,
			}), nil
		})
}

func (m *ManagementImpl) UserVerifyTwoFactorCode(ctx context.Context,
	command UserVerifyTwoFactorLoginCommand,
	agent string) (r.Response[any], error) {

	match, err := evercore.InContext(
		ctx,
		m.store,
		func(etx evercore.EventStoreContext) (bool, error) {
			aggregate := UserAggregate{}
			err := etx.LoadStateInto(&aggregate, command.UserId)
			if err != nil {
				return false, fmt.Errorf("failed to load user: %w", err)
			}

			if aggregate.State.TwoFactorSharedSecret == nil {
				return false, fmt.Errorf("user does not have two factor enabled")
			}

			decryptedUrlBytes, err := m.encryptionService.Decrypt64(*aggregate.State.TwoFactorSharedSecret)
			if err != nil {
				return false, fmt.Errorf("failed to decrypt totp: %w", err)
			}
			decryptedUrl := string(decryptedUrlBytes)

			return m.twoFactorService.ValidateTotp(decryptedUrl, command.Code)
		})

	if err != nil {
		slog.Error("Error verifying two factor code", "error", err)
		return r.Error[any]("Error verifying two factor code"), err
	}

	if !match {
		slog.Error("Two factor code does not match", "code", command.Code)
		return r.StatusError[any](ubstatus.NotAuthorized, "Two factor code does not match"), nil
	}
	return r.SuccessAny(), nil
}

func (m *ManagementImpl) UserGenerateVerificationToken(ctx context.Context,
	command UserGenerateVerificationTokenCommand,
	agent string) (r.Response[UserGenerateVerificationTokenResponse], error) {

	token, err := evercore.InContext(
		ctx,
		m.store,
		func(etx evercore.EventStoreContext) (string, error) {
			aggregate := UserAggregate{}
			err := etx.LoadStateInto(&aggregate, command.Id)
			if err != nil {
				return "", fmt.Errorf("failed to load user: %w", err)
			}

			token := ubsecurity.GenerateSecureRandomString(10)
			event := UserVerificationTokenGeneratedEvent{
				Token: token,
			}

			err = etx.ApplyEventTo(&aggregate, event, time.Now(), agent)
			if err != nil {
				return "", fmt.Errorf("failed to apply user verification token generated event: %w", err)
			}

			return token, nil
		})
	if err != nil {
		slog.Error("Error generating verification token", "error", err)
		return r.Error[UserGenerateVerificationTokenResponse]("Error generating verification token"), err
	}
	return r.Success(UserGenerateVerificationTokenResponse{
		Token: token,
	}), nil
}

func (m *ManagementImpl) UserVerify(ctx context.Context,
	command UserVerifyCommand,
	agent string) (r.Response[any], error) {

	err := m.store.WithContext(
		ctx,
		func(etx evercore.EventStoreContext) error {
			aggregate := UserAggregate{}
			err := etx.LoadStateInto(&aggregate, command.Id)
			if err != nil {
				return fmt.Errorf("failed to load user: %w", err)
			}

			if aggregate.State.VerificationToken == nil {
				return fmt.Errorf("user does not have verification token enabled")
			}

			if *aggregate.State.VerificationToken != command.Verification {
				return fmt.Errorf("verification token does not match")
			}

			event := UserVerificationTokenVerifiedEvent{}
			err = etx.ApplyEventTo(&aggregate, event, time.Now(), agent)
			if err != nil {
				return fmt.Errorf("failed to apply user verification token verified event: %w", err)
			}

			return nil
		})
	if err != nil {
		slog.Error("Error verifying user", "error", err)
		return r.Error[any]("Error verifying user"), err
	}
	return r.SuccessAny(), nil
}

func (m *ManagementImpl) GenerateTwoFactorSharedSecret(ctx context.Context,
	command GenerateTwoFactorSharedSecretCommand) (r.Response[GenerateTwoFactorSharedSecretResponse], error) {

	sharedSecret, err := evercore.InReadonlyContext(
		ctx,
		m.store,
		func(etx evercore.EventStoreReadonlyContext) (string, error) {
			aggregate := UserAggregate{}
			err := etx.LoadStateInto(&aggregate, command.Id)
			if err != nil {
				return "", fmt.Errorf("failed to load user: %w", err)
			}

			twoFactorUrl, err := m.twoFactorService.GenerateTotp(aggregate.State.Email)
			if err != nil {
				return "", fmt.Errorf("failed to generate totp: %w", err)
			}
			return twoFactorUrl, nil
		})

	if err != nil {
		slog.Error("Error generating verification token", "error", err)
		return r.Error[GenerateTwoFactorSharedSecretResponse]("Error generating verification token"), err
	}
	return r.Success(GenerateTwoFactorSharedSecretResponse{
		SharedSecret: sharedSecret,
	}), nil
}

func (m *ManagementImpl) UserSetTwoFactorSharedSecret(ctx context.Context, command UserSetTwoFactorSharedSecretCommand, agent string) (r.Response[any], error) {
	err := m.store.WithContext(
		ctx,
		func(etx evercore.EventStoreContext) error {
			aggregate := UserAggregate{}
			err := etx.LoadStateInto(&aggregate, command.Id)
			if err != nil {
				return fmt.Errorf("failed to load user: %w", err)
			}

			encryptedSecret, err := m.encryptionService.Encrypt64(command.Secret)

			event := UserTwoFactorEnabledEvent{
				SharedSecret: encryptedSecret,
			}

			err = etx.ApplyEventTo(&aggregate, event, time.Now(), agent)
			if err != nil {
				return fmt.Errorf("failed to apply user verification token generated event: %w", err)
			}

			return nil
		})
	if err != nil {
		slog.Error("Error setting two factor shared secret", "error", err)
		return r.Error[any]("Error setting two factor shared secret"), err
	}

	return r.SuccessAny(), nil
}

func (m *ManagementImpl) UserDisable(ctx context.Context,
	command UserDisableCommand, agent string) (r.Response[any], error) {
	err := m.store.WithContext(
		ctx,
		func(etx evercore.EventStoreContext) error {
			aggregate := UserAggregate{}
			err := etx.LoadStateInto(&aggregate, command.Id)
			if err != nil {
				return fmt.Errorf("failed to load user: %w", err)
			}

			event := UserDisabledEvent{}

			err = etx.ApplyEventTo(&aggregate, event, time.Now(), agent)
			if err != nil {
				return fmt.Errorf("failed to apply user disabled event: %w", err)
			}

			return nil
		})
	if err != nil {
		slog.Error("Error disabling user", "error", err)
		return r.Error[any]("Error disabling user"), err
	}
	return r.SuccessAny(), nil
}

func (m *ManagementImpl) UserEnable(ctx context.Context,
	command UserEnableCommand,
	agent string) (r.Response[any], error) {
	err := m.store.WithContext(
		ctx,
		func(etx evercore.EventStoreContext) error {
			aggregate := UserAggregate{}
			err := etx.LoadStateInto(&aggregate, command.Id)
			if err != nil {
				return fmt.Errorf("failed to load user: %w", err)
			}

			event := UserEnabledEvent{}

			err = etx.ApplyEventTo(&aggregate, event, time.Now(), agent)
			if err != nil {
				return fmt.Errorf("failed to apply user enabled event: %w", err)
			}

			return nil
		})
	if err != nil {
		slog.Error("Error enabling user", "error", err)
		return r.Error[any]("Error enabling user"), err
	}
	return r.SuccessAny(), nil
}
