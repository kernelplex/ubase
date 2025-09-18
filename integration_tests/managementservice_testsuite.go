package integration_tests

import (
	"testing"

	evercore "github.com/kernelplex/evercore/base"
	_ "github.com/kernelplex/ubase/internal/evercoregen"
	"github.com/kernelplex/ubase/lib/ub2fa"
	"github.com/kernelplex/ubase/lib/ubdata"
	"github.com/kernelplex/ubase/lib/ubmanage"
	"github.com/kernelplex/ubase/lib/ubsecurity"
)

type ManagmentServiceTestSuite struct {
	eventStore        *evercore.EventStore
	dbadapter         ubdata.DataAdapter
	managementService ubmanage.ManagementService
	twoFactorService  ub2fa.TotpService
	hashingService    ubsecurity.HashGenerator

	createdOrganizationId int64
	createdUserId         int64
	createdRoleId         int64

	userPasswordHash string
	twoFactorSecret  string
}

func NewManagementServiceTestSuite(eventStore *evercore.EventStore, dbadapter ubdata.DataAdapter) *ManagmentServiceTestSuite {

	hashingService := ubsecurity.DefaultArgon2Id
	encryptionService := ubsecurity.NewEncryptionService([]byte{
		0x1a, 0x37, 0x2b, 0x3d, 0x4e, 0x5f, 0x6c, 0x7a,
		0x8d, 0x9e, 0xaf, 0xc4, 0xd8, 0xeb, 0xf1, 0x12,
	})
	totpService := ub2fa.NewTotpService("exaple.test")

	managemntService := ubmanage.NewManagement(eventStore, dbadapter, hashingService, encryptionService, totpService)
	return &ManagmentServiceTestSuite{
		eventStore:        eventStore,
		dbadapter:         dbadapter,
		managementService: managemntService,
		twoFactorService:  totpService,
		hashingService:    hashingService,
	}
}

func (s *ManagmentServiceTestSuite) RunTests(t *testing.T) {
    t.Run("AddOrganization", s.AddOrganization)
    t.Run("GetOrganizationBySystemName", s.GetOrganizationBySystemName)
    t.Run("UpdateOrganization", s.UpdateOrganization)
    t.Run("AddOrganizationSettings", s.AddOrganizationSettings)
    t.Run("AddOrganizationSettingsValidation", s.AddOrganizationSettingsValidation)
    t.Run("RemoveOrganizationSettings", s.RemoveOrganizationSettings)
    t.Run("RemoveOrganizationSettingsValidation", s.RemoveOrganizationSettingsValidation)
	t.Run("AddRole", s.AddRole)
	t.Run("UpdateRole", s.UpdateRole)
	t.Run("DeleteRole", s.DeleteRole)
	t.Run("UndeleteRole", s.UndeleteRole)
	t.Run("AddPermissionToRole", s.AddPermissionToRole)
	t.Run("RemovePermissionFromRole", s.RemovePermissionFromRole)
	t.Run("AddUser", s.AddUser)
	t.Run("GetUserByEmail", s.GetUserByEmail)
	t.Run("UpdateUser", s.UpdateUser)
	t.Run("GetUserByEmailPostUpdate", s.GetUserByEmailPostUpdate)
	t.Run("UpdateUserWithoutPassword", s.UpdateUserWithoutPassword)
	t.Run("UpdateUserSamePassword", s.UpdateUserSamePassword)
	t.Run("LoginWithCorrectPassword", s.TestLoginWithCorrectPassword)
	t.Run("LoginWithIncorrectPassword", s.LoginWithIncorrectPassword)
	t.Run("AddTwoFactorKey", s.AddTwoFactorKey)
	t.Run("LoginWithCorrectPasswordEnsureTwoFactorRequiredIsSet", s.LoginWithCorrectPasswordEnsureTwoFactorRequiredIsSet)
	t.Run("VerifyCorrectTwoFactorCode", s.VerifyCorrectTwoFactorCode)
	t.Run("VerifyIncorrectTwoFactorCode", s.VerifyIncorrectTwoFactorCode)
	t.Run("AddUserToRole", s.AddUserToRole)
	t.Run("RemoveUserFromRole", s.RemoveUserFromRole)

	t.Run("UserAddApiKey", s.UserAddApiKey)
	t.Run("UserGetByApiKey", s.UserGetByApiKey)
	t.Run("UserDeleteApiKey", s.UserDeleteApiKey)

}
