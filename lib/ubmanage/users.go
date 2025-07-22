package ubmanage

import (
	"time"

	evercore "github.com/kernelplex/evercore/base"
	events "github.com/kernelplex/ubase/internal/evercoregen/events"
	"github.com/kernelplex/ubase/lib/ubvalidation"
)

type UserState struct {
	Email                 string  `json:"email"`
	PasswordHash          string  `json:"passwordHash"`
	FirstName             string  `json:"firstName"`
	LastName              string  `json:"lastName"`
	VerificationToken     *string `json:"verificationToken,omitempty"`
	Verified              bool    `json:"verified"`
	Disabled              bool    `json:"disabled"`
	DisplayName           string  `json:"displayName"`
	ResetToken            *string `json:"resetToken,omitempty"`
	LastLogin             int64   `json:"lastLogin,omitempty"`
	LastLoginAttempt      int64   `json:"lastLoginAttempt,omitempty"`
	FailedLoginAttempts   int64   `json:"failedLoginAttempts,omitempty"`
	TwoFactorSharedSecret *string `json:"twoFactorSharedSecret,omitempty"`
	Roles                 []int64 `json:"roles,omitempty"`
}

// evercore:aggregate
type UserAggregate struct {
	evercore.StateAggregate[UserState]
}

func (t *UserAggregate) ApplyEventState(eventState evercore.EventState, eventTime time.Time, reference string) error {
	switch ev := eventState.(type) {
	case UserLoginSucceededEvent:
		t.State.LastLogin = eventTime.Unix()
		t.State.FailedLoginAttempts = 0
		return nil
	case UserLoginFailedEvent:
		t.State.LastLoginAttempt = eventTime.Unix()
		t.State.FailedLoginAttempts++
		return nil
	case UserVerificationTokenGeneratedEvent:
		t.State.VerificationToken = &ev.Token
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
	default:
		return t.StateAggregate.ApplyEventState(eventState, eventTime, reference)
	}
}

// ============================================================================
// Commands
// ============================================================================

type UserCreateCommand struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	DisplayName string `json:"displayName"`
	Verified    bool   `json:"verified"`
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

type UserGenerateTwoFactorSharedSecretCommand struct {
	Id           int64  `json:"id"`
	SharedSecret string `json:"sharedSecret"`
}

type UserGenerateTwoFactorSharedSecretResponse struct {
	SharedSecret string `json:"sharedSecret"`
}

type UserLoginCommand struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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
type UserLoginFailedEvent struct {
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
