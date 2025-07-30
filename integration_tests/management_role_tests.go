package integration_tests

import (
	"context"
	"slices"
	"testing"

	"github.com/kernelplex/ubase/lib/ubmanage"
	"github.com/kernelplex/ubase/lib/ubstatus"
)

var aRole = ubmanage.RoleState{
	Name:       "Test Role",
	SystemName: "test_role",
}

var updatedRole = ubmanage.RoleState{
	Name:       "Updated Role",
	SystemName: "updated_role",
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

func (s *ManagmentServiceTestSuite) AddPermissionToRole(t *testing.T) {

	// Add permission to role
	permission := "test.permission"
	res, err := s.managementService.RolePermissionAdd(context.Background(), ubmanage.RolePermissionAddCommand{
		Id:         s.createdRoleId,
		Permission: permission,
	}, "test-runner")

	if err != nil {
		t.Fatalf("AddPermissionToRole failed to add permission to role: %v", err)
	}
	if res.Status != ubstatus.Success {
		t.Fatalf("AddPermissionToRole status is not success: %v", res.Status)
	}

	response, err := s.managementService.RoleGetById(context.Background(), s.createdRoleId)

	if err != nil {
		t.Fatalf("AddPermissionToRole failed to get role by ID: %v", err)
	}

	role := response.Data.State

	// Ensure the premission was added to role
	found := slices.Contains(role.Permissions, permission)
	if !found {
		t.Fatalf("AddPermissionToRole failed to add permission to role: %v", err)
	}

	perms, err := s.dbadapter.GetRolePermissions(context.Background(), s.createdRoleId)
	if err != nil {
		t.Fatalf("AddPermissionToRole failed to get role permissions: %v", err)
	}

	found = slices.Contains(perms, permission)
	if !found {
		t.Fatalf("AddPermissionToRole failed to add permission to role: %v", err)
	}
}

func (s *ManagmentServiceTestSuite) RemovePermissionFromRole(t *testing.T) {
	// First ensure test role exists with a permission
	permission := "test.permission_to_remove"
	_, err := s.managementService.RolePermissionAdd(context.Background(), ubmanage.RolePermissionAddCommand{
		Id:         s.createdRoleId,
		Permission: permission,
	}, "test-runner")
	if err != nil {
		t.Fatalf("RemovePermissionFromRole failed to add permission: %v", err)
	}

	// Remove the permission
	res, err := s.managementService.RolePermissionRemove(context.Background(), ubmanage.RolePermissionRemoveCommand{
		Id:         s.createdRoleId,
		Permission: permission,
	}, "test-runner")

	if err != nil {
		t.Fatalf("RemovePermissionFromRole failed to remove permission: %v", err)
	}
	if res.Status != ubstatus.Success {
		t.Fatalf("RemovePermissionFromRole status is not success: %v", res.Status)
	}

	// Verify in aggregate
	response, err := s.managementService.RoleGetById(context.Background(), s.createdRoleId)
	if err != nil {
		t.Fatalf("RemovePermissionFromRole failed to get role by ID: %v", err)
	}

	role := response.Data.State
	if slices.Contains(role.Permissions, permission) {
		t.Fatalf("RemovePermissionFromRole permission still exists in aggregate")
	}

	// Verify in database
	perms, err := s.dbadapter.GetRolePermissions(context.Background(), s.createdRoleId)
	if err != nil {
		t.Fatalf("RemovePermissionFromRole failed to get role permissions: %v", err)
	}

	if slices.Contains(perms, permission) {
		t.Fatalf("RemovePermissionFromRole permission still exists in database")
	}
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
