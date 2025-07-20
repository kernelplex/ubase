package integration_tests

import (
	"testing"

	evercore "github.com/kernelplex/evercore/base"
	_ "github.com/kernelplex/ubase/internal/evercoregen"
	"github.com/kernelplex/ubase/lib/ubdata"
	"github.com/kernelplex/ubase/lib/ubmanage"
	"github.com/kernelplex/ubase/lib/ubsecurity"
)

type ManagmentServiceTestSuite struct {
	eventStore        *evercore.EventStore
	dbadapter         ubdata.DataAdapter
	managementService ubmanage.ManagementService

	createdOrganizationId int64
	createdUserId         int64
	createdRoleId         int64

	userPasswordHash string
}

func NewManagementServiceTestSuite(eventStore *evercore.EventStore, dbadapter ubdata.DataAdapter) *ManagmentServiceTestSuite {

	hashingService := ubsecurity.DefaultArgon2Id
	encryptionService := ubsecurity.CreateEncryptionService([]byte("SuperSecretPassword123!!!"))

	managemntService := ubmanage.NewManagement(eventStore, dbadapter, hashingService, encryptionService)
	return &ManagmentServiceTestSuite{
		eventStore:        eventStore,
		dbadapter:         dbadapter,
		managementService: managemntService,
	}
}

func (s *ManagmentServiceTestSuite) RunTests(t *testing.T) {
	t.Run("AddOrganization", s.AddOrganization)
	t.Run("GetOrganizationBySystemName", s.GetOrganizationBySystemName)
	t.Run("UpdateOrganization", s.UpdateOrganization)
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
}
