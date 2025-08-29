package ubmanage

import (
	"context"

	evercore "github.com/kernelplex/evercore/base"
	"github.com/kernelplex/ubase/lib/ensure"
	"github.com/kernelplex/ubase/lib/ub2fa"
	"github.com/kernelplex/ubase/lib/ubdata"
	r "github.com/kernelplex/ubase/lib/ubresponse"
	"github.com/kernelplex/ubase/lib/ubsecurity"
)

type IdValue struct {
	Id int64 `json:"id"`
}

type UserCreatedResponse struct {
	Id                int64   `json:"id"`
	VerificationToken *string `json:"-"`
}

// ManagementService defines the interface for user, organization and role management operations
type ManagementService interface {

	// Organization operations

	// OrganizationList lists all organizations
	OrganizationList(ctx context.Context) (r.Response[[]ubdata.Organization], error)

	// OrganizationAdd creates a new organization with the given details
	// Returns the ID of the newly created organization or an error
	OrganizationAdd(ctx context.Context,
		command OrganizationCreateCommand,
		agent string) (r.Response[IdValue], error)

	OrganizationGet(ctx context.Context, organizationID int64) (r.Response[OrganizationAggregate], error)

	// OrganizationGetBySystemName retrieves an organization by its system name
	// Returns the organization details or an error if not found
	OrganizationGetBySystemName(
		ctx context.Context,
		systemName string) (r.Response[OrganizationAggregate], error)

	// OrganizationUpdate modifies an existing organization's details
	// Returns success/failure status or an error
	OrganizationUpdate(ctx context.Context,
		command OrganizationUpdateCommand,
		agent string) (r.Response[any], error)

	// Role operations

	// RoleAdd creates a new role with the given details
	// Returns the ID of the newly created role or an error
	RoleAdd(ctx context.Context,
		command RoleCreateCommand,
		agent string) (r.Response[IdValue], error)

	// RoleUpdate modifies an existing role's details
	// Returns success/failure status or an error
	RoleUpdate(ctx context.Context,
		command RoleUpdateCommand,
		agent string) (r.Response[any], error)

	// RoleList lists all roles for a given organization
	RoleList(ctx context.Context, OrganizationId int64) (r.Response[[]ubdata.RoleRow], error)

	// RoleGetById retrieves a role by its ID
	// Returns the role details or an error if not found
	RoleGetById(ctx context.Context,
		roleId int64) (r.Response[RoleAggregate], error)

	// RoleGetBySystemName retrieves a role by its system name
	// Returns the role details or an error if not found
	RoleGetBySystemName(ctx context.Context,
		systemName string) (r.Response[RoleAggregate], error)

	// RoleDelete marks a role as deleted (soft delete)
	// Returns success/failure status or an error
	RoleDelete(ctx context.Context,
		command RoleDeleteCommand,
		agent string) (r.Response[any], error)

	// RoleUndelete restores a previously deleted role
	// Returns success/failure status or an error
	RoleUndelete(ctx context.Context,
		command RoleUndeleteCommand,
		agent string) (r.Response[any], error)

	// RolePermissionAdd grants a permission to a role
	// Returns success/failure status or an error
	RolePermissionAdd(ctx context.Context,
		command RolePermissionAddCommand,
		agent string) (r.Response[any], error)

	// RolePermissionRemove revokes a permission from a role
	// Returns success/failure status or an error
	RolePermissionRemove(ctx context.Context,
		command RolePermissionRemoveCommand,
		agent string) (r.Response[any], error)

	// User operations

	// UserAdd creates a new user with the given details
	// Returns the ID of the newly created user or an error
	UserAdd(ctx context.Context,
		command UserCreateCommand,
		agent string) (r.Response[UserCreatedResponse], error)

	// UserGetById retrieves a user by their ID
	// Returns the user details or an error if not found
	UserGetById(ctx context.Context,
		userId int64) (r.Response[UserAggregate], error)

	// UserGetByEmail retrieves a user by email address
	// Returns the user details or an error if not found
	UserGetByEmail(ctx context.Context,
		email string) (r.Response[UserAggregate], error)

	// UserUpdate modifies an existing user's details
	// Returns success/failure status or an error
	UserUpdate(ctx context.Context,
		command UserUpdateCommand,
		agent string) (r.Response[any], error)

	// UserAuthenticate verifies user credentials and authenticates the user
	// Returns authentication status and user details if successful
	UserAuthenticate(ctx context.Context,
		command UserLoginCommand,
		agent string) (r.Response[*UserAuthenticationResponse], error)

	// UserVerifyTwoFactorCode verifies a 2FA code for an authenticated user
	// Returns success/failure status or an error
	UserVerifyTwoFactorCode(ctx context.Context,
		command UserVerifyTwoFactorLoginCommand,
		agent string) (r.Response[any], error)

	// UserGenerateVerificationToken creates a verification token for the user
	// Returns the generated token or an error
	UserGenerateVerificationToken(ctx context.Context,
		command UserGenerateVerificationTokenCommand,
		agent string) (r.Response[UserGenerateVerificationTokenResponse], error)

	// UserVerify verifies a user's account using a verification token
	// Returns success/failure status or an error
	UserVerify(ctx context.Context,
		command UserVerifyCommand,
		agent string) (r.Response[any], error)

	// UserGenerateTwoFactorSharedSecret generates a new 2FA shared secret for the user
	// Returns the secret and setup details or an error
	GenerateTwoFactorSharedSecret(
		ctx context.Context,
		command GenerateTwoFactorSharedSecretCommand) (r.Response[GenerateTwoFactorSharedSecretResponse], error)

	// UserSetTwoFactorSharedSecret sets the 2FA shared secret for the user
	UserSetTwoFactorSharedSecret(ctx context.Context, command UserSetTwoFactorSharedSecretCommand, agent string) (r.Response[any], error)

	// UserDisable deactivates a user account
	// Returns success/failure status or an error
	UserDisable(ctx context.Context,
		command UserDisableCommand,
		agent string) (r.Response[any], error)

	// UserEnable reactivates a previously disabled user account
	// Returns success/failure status or an error
	UserEnable(ctx context.Context,
		command UserEnableCommand,
		agent string) (r.Response[any], error)

	// UserAddToRole assigns a role to a user
	// Returns success/failure status or an error
	UserAddToRole(ctx context.Context,
		command UserAddToRoleCommand,
		agent string) (r.Response[any], error)

	// UserGetOrganizationRoles retrieves all roles a user has in a specific organization
	// Returns the list of roles or an error
	UserGetOrganizationRoles(ctx context.Context, userId int64, organizationId int64) (r.Response[[]ubdata.RoleRow], error)

	UserGetAllOrganizationRoles(ctx context.Context, userId int64) (r.Response[[]ubdata.ListUserOrganizationRolesRow], error)

	// UserRemoveFromRole revokes a role from a user
	// Returns success/failure status or an error
	UserRemoveFromRole(ctx context.Context,
		command UserRemoveFromRoleCommand,
		agent string) (r.Response[any], error)

	// UsersCount returns the total number of users in the system
	UsersCount(ctx context.Context) (r.Response[int64], error)
	// OrganizationsCount returns the total number of organizations in the system
	OrganizationsCount(ctx context.Context) (r.Response[int64], error)

	// ListOrganizationsWithUserCounts lists all organizations along with the count of users in each
	OrganizationRolesWithUserCount(ctx context.Context, organizationId int64) (r.Response[[]ubdata.ListRolesWithUserCountsRow], error)
}

type ManagementImpl struct {
	store             *evercore.EventStore
	dbadapter         ubdata.DataAdapter
	hashingService    ubsecurity.HashGenerator
	encryptionService ubsecurity.EncryptionService
	twoFactorService  ub2fa.TotpService
}

func Must(condition bool, message string) {
	if !condition {
		panic(message)
	}
}

func NewManagement(
	store *evercore.EventStore,
	dbadapter ubdata.DataAdapter,
	hashingService ubsecurity.HashGenerator,
	encryptionService ubsecurity.EncryptionService,
	twoFactorService ub2fa.TotpService,
) ManagementService {

	// Negative space assertions
	ensure.That(store != nil, "store cannot be nil")
	ensure.That(dbadapter != nil, "dbadapter cannot be nil")
	ensure.That(hashingService != nil, "hashingService cannot be nil")
	ensure.That(encryptionService != nil, "encryptionService cannot be nil")
	ensure.That(twoFactorService != nil, "twoFactorService cannot be nil")

	management := ManagementImpl{
		store:             store,
		dbadapter:         dbadapter,
		hashingService:    hashingService,
		encryptionService: encryptionService,
		twoFactorService:  twoFactorService,
	}
	return &management
}
