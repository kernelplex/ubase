package integration_tests

import (
	"context"
	"testing"

	"github.com/kernelplex/ubase/lib/ubmanage"
	"github.com/kernelplex/ubase/lib/ubstatus"
)

func (s *ManagmentServiceTestSuite) AddUserToRole(t *testing.T) {
	// Add user to role
	res, err := s.managementService.UserAddToRole(context.Background(), ubmanage.UserAddToRoleCommand{
		UserId: s.createdUserId,
		RoleId: s.createdRoleId,
	}, "test-runner")

	if err != nil {
		t.Fatalf("AddUserToRole failed to add user to role: %v", err)
	}
	if res.Status != ubstatus.Success {
		t.Fatalf("AddUserToRole status is not success: %v", res.Status)
	}

	// Verify the user was added to the role
	response, err := s.managementService.UserGetOrganizationRoles(context.Background(), s.createdUserId, s.createdOrganizationId)
	if err != nil {
		t.Fatalf("UserGetOrganizationRoles failed to get user organization roles: %v", err)
	}
	if len(response.Data) != 1 {
		t.Fatalf("UserGetOrganizationRoles failed to get user organization roles: %v", err)
	}

	if response.Data[0].ID != s.createdRoleId {
		t.Fatalf("UserGetOrganizationRoles failed to get user organization roles: %v", err)
	}

	if response.Data[0].Name != updatedRole.Name {
		t.Fatalf("UserGetOrganizationRoles failed to get user organization roles: %v", err)
	}

	if response.Data[0].SystemName != updatedRole.SystemName {
		t.Fatalf("UserGetOrganizationRoles failed to get user organization roles: %v", err)
	}

}

func (s *ManagmentServiceTestSuite) RemoveUserFromRole(t *testing.T) {
	res, err := s.managementService.UserRemoveFromRole(context.Background(), ubmanage.UserRemoveFromRoleCommand{
		UserId: s.createdUserId,
		RoleId: s.createdRoleId,
	}, "test-runner")

	if err != nil {
		t.Fatalf("RemoveUserToRole failed to add user to role: %v", err)
	}
	if res.Status != ubstatus.Success {
		t.Fatalf("RemoveUserFromRole status is not success: %v", res.Status)
	}

	// Verify the user was added to the role
	response, err := s.managementService.UserGetOrganizationRoles(context.Background(), s.createdUserId, s.createdOrganizationId)
	if err != nil {
		t.Fatalf("RemoveUserFromRole failed to get user organization roles: %v", err)
	}
	if len(response.Data) != 0 {
		t.Fatalf("RemoveUserFromRole failed to get user organization roles: %v", err)
	}
}
