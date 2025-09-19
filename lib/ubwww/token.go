package ubwww

import (
	"fmt"
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
