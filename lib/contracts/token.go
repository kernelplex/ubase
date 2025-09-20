package contracts

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type AuthTokenCookieContextKey string

const AuthTokenCookieContextKeyStr = AuthTokenCookieContextKey("authToken")

type AuthToken struct {
	UserId               int64  `json:"userId"`
	OrganizationId       int64  `json:"organizationId"`
	Email                string `json:"email"`
	TwoFactorRequired    bool   `json:"twoFactorRequired"`
	RequiresVerification bool   `json:"requiresVerification"`
	SoftExpiry           int64  `json:"softExpiry"`
	HardExpiry           int64  `json:"hardExpiry"`
}

func (t *AuthToken) ToAgent() string {
	return fmt.Sprintf("user:%d", t.UserId)
}

func (t *AuthToken) IsExpired() bool {
	now := time.Now().Unix()
	return now > t.HardExpiry || now > t.SoftExpiry
}

func (t *AuthToken) Touch(updatedTime int64) {
	t.SoftExpiry = updatedTime
}

func (t *AuthToken) ToUserIdentity() UserIdentity {
	return UserIdentity{
		UserID:         t.UserId,
		OrganizationID: t.OrganizationId,
		Email:          t.Email,
	}
}

type CookieContextKey string
type IdentityContextKey string

type UserIdentity struct {
	UserID         int64
	Email          string
	OrganizationID int64
}

func (u *UserIdentity) ToAgent() string {
	return fmt.Sprintf("user:%d", u.UserID)
}

type AuthTokenCookie interface {
	ToUserIdentity() UserIdentity
	IsExpired() bool
	Touch(unixSec int64)
}

type AuthTokenCookieManager interface {
	ClearAuthTokenCookie(w http.ResponseWriter)
	ReadAuthTokenCookie(r *http.Request) (bool, AuthToken, error)
	WriteAuthTokenCookie(w http.ResponseWriter, token AuthToken) error
	Middleware(handler http.Handler) http.Handler
	MiddlewareFunc(handler http.HandlerFunc) http.HandlerFunc
	TokenFromContext(ctx context.Context) (AuthToken, bool)
	IdentityFromContext(ctx context.Context) (UserIdentity, bool)
}
