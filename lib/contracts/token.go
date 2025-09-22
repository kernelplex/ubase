package contracts

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/kernelplex/ubase/lib/ubdata"
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

// ViewModel structs for template data consolidation

type BaseViewModel struct {
	Fragment bool
	Links    []AdminLink
}

type LoginViewModel struct {
	BaseViewModel
	Error string
}

type TwoFactorViewModel struct {
	BaseViewModel
	UserID int64
	Error  string
}

type AdminPanelViewModel struct {
	BaseViewModel
	OrgCount   int64
	UserCount  int64
	RoleCount  int64
	Recent     []RecentUser
}

type RecentUser struct {
	ID          int64
	DisplayName string
	Email       string
	LastLogin   int64
}

type OrganizationsPageViewModel struct {
	BaseViewModel
	Organizations []ubdata.Organization
	Query         string
}

type OrganizationOverviewViewModel struct {
	BaseViewModel
	ID         int64
	Name       string
	SystemName string
	Roles      []ubdata.ListRolesWithUserCountsRow
}

type OrganizationFormViewModel struct {
	BaseViewModel
	IsEdit      bool
	Organization *ubdata.Organization
	Error       string
	FieldErrors map[string][]string
}

type UsersPageViewModel struct {
	BaseViewModel
	Users []ubdata.User
	Query string
}

type UserOverviewViewModel struct {
	BaseViewModel
	ID                   int64
	DisplayName          string
	Email                string
	FirstName            string
	LastName             string
	Verified             bool
	Disabled             bool
	LastLogin            int64
	LoginCount           int64
	LastFailedLogin      int64
	FailedLoginAttempts  int64
	Organizations        []ubdata.Organization
	SelectedOrganization int64
}

type UserFormViewModel struct {
	BaseViewModel
	IsEdit     bool
	User       *ubdata.User
	Error      string
	FieldErrors map[string][]string
}

type UserRolesViewModel struct {
	BaseViewModel
	UserID     int64
	Roles      []ubdata.RoleRow
	MemberSet  map[int64]bool
	OrgID      int64
}

type RoleOverviewViewModel struct {
	BaseViewModel
	ID             int64
	Name           string
	SystemName     string
	OrganizationID int64
}

type RoleUsersViewModel struct {
	BaseViewModel
	RoleID    int64
	Users     []ubdata.User
	MemberSet map[int64]bool
}

type RolePermissionsViewModel struct {
	BaseViewModel
	RoleID    int64
	Permissions []string
	MemberSet  map[string]bool
}

type RoleFormViewModel struct {
	BaseViewModel
	IsEdit           bool
	Role             *ubdata.RoleRow
	Organizations    []ubdata.Organization
	SelectedOrg      int64
	Error            string
	FieldErrors      map[string][]string
}
