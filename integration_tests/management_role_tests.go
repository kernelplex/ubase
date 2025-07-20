package integration_tests

import (
	"context"
	"testing"

	"github.com/kernelplex/ubase/lib/ubmanage"
	"github.com/kernelplex/ubase/lib/ubstatus"
)

var aRole = ubmanage.RoleState{
	Name:       "Test Role",
	SystemName: "test-role",
}

var updatedRole = ubmanage.RoleState{
	Name:       "Updated Role",
	SystemName: "updated-role",
}

func (s *ManagmentServiceTestSuite) AddRole(t *testing.T) {
	response, err := s.managementService.RoleAdd(context.Background(), ubmanage.RoleCreateCommand{
		OrganizationId: s.createdOrganizationId,
		Name:           aRole.Name,
		SystemName:     aRole.SystemName,
	}, "test-runner")

	if err != nil {
		t.Fatalf("AddRole failed to add role: %v", err)
	}

	if response.Status != ubstatus.Success {
		t.Fatalf("AddRole status is not success: %v", response.Status)
	}

	s.createdRoleId = response.Data.Id

	// Verify the role was added to the database
	foundRoles, err := s.dbadapter.GetOrganizationRoles(context.Background(), s.createdOrganizationId)
	if err != nil {
		t.Fatalf("AddRole failed to get organization roles: %v", err)
	}

	if len(foundRoles) != 1 {
		t.Fatalf("AddRole failed to get organization roles: %v", err)
	}

	foundRoles[0].ID = response.Data.Id
	foundRoles[0].Name = aRole.Name
	foundRoles[0].SystemName = aRole.SystemName

}

func (s *ManagmentServiceTestSuite) UpdateRole(t *testing.T) {
	res, err := s.managementService.RoleUpdate(context.Background(), ubmanage.RoleUpdateCommand{
		Id:         s.createdRoleId,
		Name:       &updatedRole.Name,
		SystemName: &updatedRole.SystemName,
	}, "test-runner")

	if err != nil {
		t.Fatalf("UpdateRole failed to update role: %v", err)
	}
	if res.Status != ubstatus.Success {
		t.Fatalf("UpdateRole status is not success: %v", res.Status)
	}

	// Verify the role was updated in the database
	foundRoles, err := s.dbadapter.GetOrganizationRoles(context.Background(), s.createdOrganizationId)
	if err != nil {
		t.Fatalf("UpdateRole failed to get organization roles: %v", err)
	}

	if len(foundRoles) != 1 {
		t.Fatalf("UpdateRole failed to get organization roles: %v", err)
	}

	foundRoles[0].ID = s.createdRoleId
	foundRoles[0].Name = updatedRole.Name
	foundRoles[0].SystemName = updatedRole.SystemName
}

func (s *ManagmentServiceTestSuite) DeleteRole(t *testing.T) {
	res, err := s.managementService.RoleDelete(context.Background(), ubmanage.RoleDeleteCommand{
		Id: s.createdRoleId,
	}, "test-runner")

	if err != nil {
		t.Fatalf("DeleteRole failed to delete role: %v", err)
	}
	if res.Status != ubstatus.Success {
		t.Fatalf("DeleteRole status is not success: %v", res.Status)
	}

	foundRoles, err := s.dbadapter.GetOrganizationRoles(context.Background(), s.createdOrganizationId)
	if err != nil {
		t.Fatalf("DeleteRole failed to get organization roles: %v", err)
	}

	if len(foundRoles) != 0 {
		t.Fatalf("DeleteRole failed to delete role expected foundRows to be 0")
	}
}

func (s *ManagmentServiceTestSuite) UndeleteRole(t *testing.T) {
	res, err := s.managementService.RoleUndelete(context.Background(), ubmanage.RoleUndeleteCommand{
		Id: s.createdRoleId,
	}, "test-runner")

	if err != nil {
		t.Fatalf("UndeleteRole failed to undelete role: %v", err)
	}
	if res.Status != ubstatus.Success {
		t.Fatalf("UndeleteRole status is not success: %v", res.Status)
	}

	foundRoles, err := s.dbadapter.GetOrganizationRoles(context.Background(), s.createdOrganizationId)
	if err != nil {
		t.Fatalf("DeleteRole failed to get organization roles: %v", err)
	}

	if len(foundRoles) != 1 {
		t.Fatalf("UndeleteRole failed to undelete role expected foundRows to be 1")
	}

	foundRoles[0].ID = s.createdRoleId
	foundRoles[0].Name = updatedRole.Name
	foundRoles[0].SystemName = updatedRole.SystemName
}
