package ubmanage

import (
	"context"

	evercore "github.com/kernelplex/evercore/base"
	"github.com/kernelplex/ubase/lib/ensure"
	"github.com/kernelplex/ubase/lib/ub2fa"
	"github.com/kernelplex/ubase/lib/ubdata"
	"github.com/kernelplex/ubase/lib/ubsecurity"
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

	// Role operations

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
		agent string) (Response[*UserAuthenticationResponse], error)

	UserVerifyTwoFactorCode(ctx context.Context,
		command UserVerifyTwoFactorLoginCommand,
		agent string) (Response[any], error)

	UserGenerateVerificationToken(ctx context.Context,
		command UserGenerateVerificationTokenCommand,
		agent string) (Response[UserGenerateVerificationTokenResponse], error)

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

	UserAddToRole(ctx context.Context,
		command UserAddToRoleCommand,
		agent string) (Response[any], error)

	UserGetOrganizationRoles(ctx context.Context, userId int64, organizationId int64) (Response[[]ubdata.RoleRow], error)

	UserRemoveFromRole(ctx context.Context,
		command UserRemoveFromRoleCommand,
		agent string) (Response[any], error)
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
