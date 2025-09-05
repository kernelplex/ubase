package ubmanage

import (
	"time"

	evercore "github.com/kernelplex/evercore/base"
	events "github.com/kernelplex/ubase/internal/evercoregen/events"
	"github.com/kernelplex/ubase/lib/ubvalidation"
)

type ApiKey struct {
	Id             string `json:"id,omitempty"`
	OrganizationId int64  `json:"organizationId,omitempty"`
	SecretHash     string `json:"secretHash,omitempty"`
	Name           string `json:"name,omitempty"`
	ExpiresAt      int64  `json:"expiresAt,omitempty"`
}

type UserState struct {
	Email                 string   `json:"email"`
	PasswordHash          string   `json:"passwordHash"`
	FirstName             string   `json:"firstName"`
	LastName              string   `json:"lastName"`
	VerificationToken     *string  `json:"verificationToken,omitempty"`
	Verified              bool     `json:"verified"`
	Disabled              bool     `json:"disabled"`
	DisplayName           string   `json:"displayName"`
	ResetToken            *string  `json:"resetToken,omitempty"`
	LastLogin             int64    `json:"lastLogin,omitempty"`
	LastLoginAttempt      int64    `json:"lastLoginAttempt,omitempty"`
	FailedLoginAttempts   int64    `json:"failedLoginAttempts,omitempty"`
	TwoFactorSharedSecret *string  `json:"twoFactorSharedSecret,omitempty"`
	LoginCount            int64    `json:"loginCount,omitempty"`
	CreatedAt             int64    `json:"createdAt,omitempty"`
	UpdatedAt             int64    `json:"updatedAt,omitempty"`
	ApiKeys               []ApiKey `json:"apiKeys,omitempty"`
}

// evercore:aggregate
type UserAggregate struct {
	evercore.StateAggregate[UserState]
}

func (t *UserAggregate) ApplyEventState(eventState evercore.EventState, eventTime time.Time, reference string) error {
	var err error = nil

	switch ev := eventState.(type) {
	case UserLoginSucceededEvent:
		t.State.LastLogin = eventTime.Unix()
		t.State.FailedLoginAttempts = 0
		t.State.LoginCount++
		return nil
	case UserLoginPartiallySucceededEvent:
		t.State.LastLogin = eventTime.Unix()
		t.State.LastLoginAttempt = eventTime.Unix()
		t.State.FailedLoginAttempts = 0
		return nil
	case UserLoginFailedEvent:
		t.State.LastLoginAttempt = eventTime.Unix()
		t.State.FailedLoginAttempts++
		return nil
	case UserVerificationTokenGeneratedEvent:
		t.State.VerificationToken = &ev.Token
		t.State.Verified = false
		return nil
	case UserVerificationTokenVerifiedEvent:
		t.State.Verified = true
		t.State.VerificationToken = nil
		return nil
	case UserTwoFactorEnabledEvent:
		t.State.TwoFactorSharedSecret = &ev.SharedSecret
		return nil
	case UserTwoFactorDisabledEvent:
		t.State.TwoFactorSharedSecret = nil
		return nil
	case UserTwoFactorAuthenticatedEvent:
		return nil
	case UserDisabledEvent:
		t.State.Disabled = true
		return nil
	case UserEnabledEvent:
		t.State.Disabled = false
		return nil
	case UserApiKeyAddedEvent:
		t.State.ApiKeys = append(t.State.ApiKeys, ApiKey{
			Id:             ev.Id,
			OrganizationId: ev.OrganizationId,
			SecretHash:     ev.SecretHash,
			Name:           ev.Name,
			ExpiresAt:      ev.ExpiresAt,
		})
		return nil
	case UserApiKeyDeletedEvent:
		newApiKeys := make([]ApiKey, 0, len(t.State.ApiKeys))
		for _, apiKey := range t.State.ApiKeys {
			if apiKey.Id != ev.Id {
				newApiKeys = append(newApiKeys, apiKey)
			}
		}
		t.State.ApiKeys = newApiKeys
		return nil
	default:
		err = t.StateAggregate.ApplyEventState(eventState, eventTime, reference)
	}

	if err != nil {
		return err
	}

	if eventState.GetEventType() != events.UserAddedEventType {
		t.State.ApiKeys = make([]ApiKey, len(t.State.ApiKeys))
		t.State.CreatedAt = eventTime.Unix()
		t.State.UpdatedAt = eventTime.Unix()
	}

	if eventState.GetEventType() == events.UserUpdatedEventType {
		t.State.UpdatedAt = eventTime.Unix()
	}
	return err
}

// ============================================================================
// Commands
// ============================================================================

type UserCreateCommand struct {
	Email                     string `json:"email"`
	Password                  string `json:"password"`
	FirstName                 string `json:"firstName"`
	LastName                  string `json:"lastName"`
	DisplayName               string `json:"displayName"`
	Verified                  bool   `json:"verified"`
	GenerateVerificationToken bool   `json:"verificationRequired"`
}

func (c UserCreateCommand) Validate() (bool, []ubvalidation.ValidationIssue) {
	validationTracker := ubvalidation.NewValidationTracker()

	// Validate required fields
	validationTracker.ValidateEmail("Email", c.Email)
	validationTracker.ValidatePasswordComplexity("Password", c.Password)
	validationTracker.ValidateField("Password", c.Password, true, 0)
	validationTracker.ValidateField("FirstName", c.FirstName, true, 0)
	validationTracker.ValidateField("LastName", c.LastName, true, 0)
	validationTracker.ValidateField("DisplayName", c.DisplayName, true, 0)

	return validationTracker.Valid()
}

type UserUpdateCommand struct {
	Id          int64   `json:"id"`
	Email       *string `json:"email"`
	Password    *string `json:"password"`
	FirstName   *string `json:"firstName"`
	LastName    *string `json:"lastName"`
	DisplayName *string `json:"displayName"`
	Verified    *bool   `json:"verified"`
}

func (c UserUpdateCommand) Validate() (bool, []ubvalidation.ValidationIssue) {
	validationTracker := ubvalidation.NewValidationTracker()

	validationTracker.ValidateIntMinValue("Id", c.Id, 1)
	validationTracker.ValidateOptionalField("Email", c.Email, 0)
	validationTracker.ValidateOptionalField("Password", c.Password, 0)
	validationTracker.ValidateOptionalField("FirstName", c.FirstName, 0)
	validationTracker.ValidateOptionalField("LastName", c.LastName, 0)
	validationTracker.ValidateOptionalField("DisplayName", c.DisplayName, 0)

	return validationTracker.Valid()
}

type UserGenerateVerificationTokenCommand struct {
	Id int64 `json:"id"`
}

type UserGenerateVerificationTokenResponse struct {
	Token string `json:"token"`
}

func (c UserGenerateVerificationTokenCommand) Validate() (bool, []ubvalidation.ValidationIssue) {
	validationTracker := ubvalidation.NewValidationTracker()

	validationTracker.ValidateIntMinValue("Id", c.Id, 1)

	return validationTracker.Valid()
}

type UserVerifyCommand struct {
	Id           int64  `json:"id"`
	Verification string `json:"verification"`
}

type GenerateTwoFactorSharedSecretCommand struct {
	Id           int64  `json:"id"`
	SharedSecret string `json:"sharedSecret"`
}

type GenerateTwoFactorSharedSecretResponse struct {
	SharedSecret string `json:"sharedSecret"`
}

type UserLoginCommand struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserSetTwoFactorSharedSecretCommand struct {
	Id     int64  `json:"id"`
	Secret string `json:"secret"`
}

type UserVerifyTwoFactorLoginCommand struct {
	UserId int64  `json:"id"`
	Code   string `json:"code"`
}

type UserDisableCommand struct {
	Id int64 `json:"id"`
}

type UserEnableCommand struct {
	Id int64 `json:"id"`
}

type UserGenerateApiKeyCommand struct {
	UserId         int64     `json:"userId"`
	Name           string    `json:"name"`
	OrganizationId int64     `json:"organizationId,omitempty"`
	ExpiresAt      time.Time `json:"expiresAt"`
}

func (c UserGenerateApiKeyCommand) Validate() (bool, []ubvalidation.ValidationIssue) {
	validationTracker := ubvalidation.NewValidationTracker()

	validationTracker.ValidateIntMinValue("UserId", c.UserId, 1)
	validationTracker.ValidateField("Name", c.Name, true, 0)
	validationTracker.ValidateIntMinValue("OrganizationId", c.OrganizationId, 1)
	return validationTracker.Valid()
}

type UserDeleteApiKeyCommand struct {
	UserId int64  `json:"userId"`
	ApiKey string `json:"apiKey"`
}

// ============================================================================
// Events
// ============================================================================

// evercore:state-event
type UserAddedEvent struct {
	Email        string `json:"email"`
	PasswordHash string `json:"passwordHash"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	DisplayName  string `json:"displayName"`
	Verified     bool   `json:"verified"`
}

// evercore:state-event
type UserUpdatedEvent struct {
	Id           int64   `json:"id"`
	Email        *string `json:"email"`
	PasswordHash *string `json:"passwordHash"`
	FirstName    *string `json:"firstName"`
	LastName     *string `json:"lastName"`
	DisplayName  *string `json:"displayName"`
	Verified     *bool   `json:"verified"`
}

// evercore:event
type UserLoginSucceededEvent struct {
}

func (a UserLoginSucceededEvent) GetEventType() string {
	return events.UserLoginSucceededEventType
}

func (a UserLoginSucceededEvent) Serialize() string {
	return evercore.SerializeToJson(a)
}

// evercore:event
type UserLoginPartiallySucceededEvent struct {
	RequiresTwoFactor    bool `json:"requiresTwoFactor,omitempty"`
	RequiresVerification bool `json:"requiresVerify,omitempty"`
}

func (a UserLoginPartiallySucceededEvent) GetEventType() string {
	return events.UserLoginPartiallySucceededEventType
}

func (a UserLoginPartiallySucceededEvent) Serialize() string {
	return evercore.SerializeToJson(a)
}

// evercore:event
type UserLoginFailedEvent struct {
	Reason string `json:"reason,omitempty"`
}

func (a UserLoginFailedEvent) GetEventType() string {
	return events.UserLoginFailedEventType
}
func (a UserLoginFailedEvent) Serialize() string {
	return evercore.SerializeToJson(a)
}

// evercore:event
type UserVerificationTokenGeneratedEvent struct {
	Token string `json:"token"`
}

func (a UserVerificationTokenGeneratedEvent) GetEventType() string {
	return events.UserVerificationTokenGeneratedEventType
}

func (a UserVerificationTokenGeneratedEvent) Serialize() string {
	return evercore.SerializeToJson(a)
}

// evercore:event
type UserVerificationTokenVerifiedEvent struct {
}

func (a UserVerificationTokenVerifiedEvent) GetEventType() string {
	return events.UserVerificationTokenVerifiedEventType
}

func (a UserVerificationTokenVerifiedEvent) Serialize() string {
	return evercore.SerializeToJson(a)
}

// evercore:event
type UserTwoFactorEnabledEvent struct {
	SharedSecret string `json:"sharedSecret"`
}

func (a UserTwoFactorEnabledEvent) GetEventType() string {
	return events.UserTwoFactorEnabledEventType
}

func (a UserTwoFactorEnabledEvent) Serialize() string {
	return evercore.SerializeToJson(a)
}

// evercore:event
type UserTwoFactorDisabledEvent struct {
}

func (a UserTwoFactorDisabledEvent) GetEventType() string {
	return events.UserTwoFactorDisabledEventType
}

func (a UserTwoFactorDisabledEvent) Serialize() string {
	return evercore.SerializeToJson(a)
}

// evercore:event
type UserTwoFactorAuthenticatedEvent struct {
}

func (a UserTwoFactorAuthenticatedEvent) GetEventType() string {
	return events.UserTwoFactorAuthenticatedEventType
}

func (a UserTwoFactorAuthenticatedEvent) Serialize() string {
	return evercore.SerializeToJson(a)
}

// evercore:event
type UserDisabledEvent struct {
}

func (a UserDisabledEvent) GetEventType() string {
	return events.UserDisabledEventType
}

func (a UserDisabledEvent) Serialize() string {
	return evercore.SerializeToJson(a)
}

// evercore:event
type UserEnabledEvent struct {
}

func (a UserEnabledEvent) GetEventType() string {
	return events.UserEnabledEventType
}

func (a UserEnabledEvent) Serialize() string {
	return evercore.SerializeToJson(a)
}

// evercore:event
type UserApiKeyAddedEvent struct {
	Id             string `json:"id"`
	OrganizationId int64  `json:"organizationId"`
	SecretHash     string `json:"secretHash"`
	Name           string `json:"name"`
	CreatedAt      int64  `json:"createdAt"`
	ExpiresAt      int64  `json:"expiresAt"`
}

func (a UserApiKeyAddedEvent) GetEventType() string {
	return events.UserApiKeyAddedEventType
}

func (a UserApiKeyAddedEvent) Serialize() string {
	return evercore.SerializeToJson(a)
}

// evercore:event
type UserApiKeyDeletedEvent struct {
	Id string `json:"apiKeyHash"`
}

func (a UserApiKeyDeletedEvent) GetEventType() string {
	return events.UserApiKeyDeletedEventType
}
func (a UserApiKeyDeletedEvent) Serialize() string {
	return evercore.SerializeToJson(a)
}
