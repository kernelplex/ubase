// Tests the ubdata.DataAdapter interface
//
// This exercises the basic data access operations of the ubase library
// and verifies that the adapter implementations work as expected.

package integration_tests

import (
	"database/sql"
	"testing"

	"github.com/kernelplex/ubase/lib/ubdata"
)

const (
	PermCanCreateRecord = "can_create_record"
	PermCanReadRecord   = "can_read_record"
	PermCanUpdateRecord = "can_update_record"
	PermCanDeleteRecord = "can_delete_record"
)

var sampleUser = ubdata.User{
	UserID:      1,
	FirstName:   "Test",
	LastName:    "User",
	DisplayName: "Test User",
	Email:       "testuser@example.com",
}

var sampleUpdatedUser = ubdata.User{
	UserID:      1,
	FirstName:   "UpdatedFirst",
	LastName:    "UpdatedLast",
	DisplayName: "Updated Display",
	Email:       "updated@example.com",
}

type Organization struct {
	OrganizationID int64
	Name           string
	SystemName     string
	Status         string
}

var sampleOrganization = Organization{
	OrganizationID: 1,
	Name:           "Test Organization",
	SystemName:     "testorg",
	Status:         "active",
}

var sampleUpdatedOrganization = Organization{
	OrganizationID: 1,
	Name:           "Updated Organization",
	SystemName:     "updatedorg",
	Status:         "inactive",
}

type AdapterExercises struct {
	adapter ubdata.DataAdapter
}

func (s *AdapterExercises) RunTests(t *testing.T) {

	t.Run("TestAddUser", s.TestAddUser)
	t.Run("TestGetUser", s.TestGetUser)
	t.Run("TestGetUserByEmail", s.TestGetUserByEmail)
	t.Run("TestUpdateUser", s.TestUpdateUser)
	t.Run("TestAddOrganization", s.TestAddOrganization)
	t.Run("TestGetOrganization", s.TestGetOrganization)

}

func NewAdapterExercises(db *sql.DB, adapter ubdata.DataAdapter) *AdapterExercises {
	return &AdapterExercises{
		adapter: adapter,
	}
}

func (s *AdapterExercises) TestAddUser(t *testing.T) {
	ctx := t.Context()

	// Create test user
	err := s.adapter.AddUser(ctx, sampleUser.UserID, sampleUser.FirstName, sampleUser.LastName, sampleUser.DisplayName, sampleUser.Email)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
}

func (s *AdapterExercises) TestGetUser(t *testing.T) {
	ctx := t.Context()
	userID := int64(1)

	user, err := s.adapter.GetUser(ctx, userID)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if user.UserID != sampleUser.UserID {
		t.Errorf("Expected user ID %d, got %d", sampleUser.UserID, user.UserID)
	}
	if user.FirstName != sampleUser.FirstName {
		t.Errorf("Expected first name %s, got %s", sampleUser.FirstName, user.FirstName)
	}
	if user.LastName != sampleUser.LastName {
		t.Errorf("Expected last name %s, got %s", sampleUser.LastName, user.LastName)
	}
}

func (s *AdapterExercises) TestGetUserByEmail(t *testing.T) {
	ctx := t.Context()
	email := "testuser@example.com"

	user, err := s.adapter.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("GetUserByEmail failed: %v", err)
	}

	if user.UserID != sampleUser.UserID {
		t.Errorf("Expected user ID %d, got %d", sampleUser.UserID, user.UserID)
	}
	if user.FirstName != sampleUser.FirstName {
		t.Errorf("Expected first name %s, got %s", sampleUser.FirstName, user.FirstName)
	}
	if user.LastName != sampleUser.LastName {
		t.Errorf("Expected last name %s, got %s", sampleUser.LastName, user.LastName)
	}
}

func (s *AdapterExercises) TestUpdateUser(t *testing.T) {
	ctx := t.Context()

	// Update the user
	err := s.adapter.UpdateUser(ctx, sampleUpdatedUser.UserID, sampleUpdatedUser.FirstName, sampleUpdatedUser.LastName, sampleUpdatedUser.DisplayName, sampleUpdatedUser.Email)
	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	// Verify the updates were applied
	user, err := s.adapter.GetUser(ctx, sampleUpdatedUser.UserID)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	if user.FirstName != sampleUpdatedUser.FirstName {
		t.Errorf("FirstName not updated, expected '%s' got '%s'", sampleUpdatedUser.FirstName, user.FirstName)
	}
	if user.LastName != sampleUpdatedUser.LastName {
		t.Errorf("LastName not updated, expected '%s' got '%s'", sampleUpdatedUser.LastName, user.LastName)
	}
	if user.DisplayName != sampleUpdatedUser.DisplayName {
		t.Errorf("DisplayName not updated, expected '%s' got '%s'", sampleUpdatedUser.DisplayName, user.DisplayName)
	}
	if user.Email != sampleUpdatedUser.Email {
		t.Errorf("Email not updated, expected '%s' got '%s'", sampleUpdatedUser.Email, user.Email)
	}
}

func (s *AdapterExercises) TestAddOrganization(t *testing.T) {
	ctx := t.Context()

	// Create test organization
	err := s.adapter.AddOrganization(ctx, sampleOrganization.OrganizationID, sampleOrganization.Name, sampleOrganization.SystemName, sampleOrganization.Status)
	if err != nil {
		t.Fatalf("CreateOrganization failed: %v", err)
	}
}

func (s *AdapterExercises) TestGetOrganization(t *testing.T) {
	ctx := t.Context()
	organizationID := int64(1)

	organization, err := s.adapter.GetOrganization(ctx, organizationID)
	if err != nil {
		t.Fatalf("GetOrganization failed: %v", err)
	}

	if organization.ID != sampleOrganization.OrganizationID {
		t.Errorf("Expected organization ID %d, got %d", sampleOrganization.OrganizationID, organization.ID)
	}
	if organization.Name != sampleOrganization.Name {
		t.Errorf("Expected organization name %s, got %s", sampleOrganization.Name, organization.Name)
	}
	if organization.SystemName != sampleOrganization.SystemName {
		t.Errorf("Expected organization system name %s, got %s", sampleOrganization.SystemName, organization.SystemName)
	}
	if organization.Status != sampleOrganization.Status {
		t.Errorf("Expected organization status %s, got %s", sampleOrganization.Status, organization.Status)
	}
}

func (s *AdapterExercises) TestGetOrganizationBySystemName(t *testing.T) {
	ctx := t.Context()
	systemName := "testorg"

	organization, err := s.adapter.GetOrganizationBySystemName(ctx, systemName)
	if err != nil {
		t.Fatalf("GetOrganizationBySystemName failed: %v", err)
	}

	if organization.ID != sampleOrganization.OrganizationID {
		t.Errorf("Expected organization ID %d, got %d", sampleOrganization.OrganizationID, organization.ID)
	}
	if organization.Name != sampleOrganization.Name {
		t.Errorf("Expected organization name %s, got %s", sampleOrganization.Name, organization.Name)
	}
	if organization.SystemName != sampleOrganization.SystemName {
		t.Errorf("Expected organization system name %s, got %s", sampleOrganization.SystemName, organization.SystemName)
	}
	if organization.Status != sampleOrganization.Status {
		t.Errorf("Expected organization status %s, got %s", sampleOrganization.Status, organization.Status)
	}
}

func (s *AdapterExercises) TestAddRole(t *testing.T) {
	ctx := t.Context()
	roleID := int64(1)
	orgID := int64(1)

	err := s.adapter.AddRole(ctx, roleID, orgID, "Admin", "admin")
	if err != nil {
		t.Fatalf("AddRole failed: %v", err)
	}
}

func (s *AdapterExercises) TestAddPermissionToRole(t *testing.T) {
	ctx := t.Context()
	roleID := int64(1)

	// First create role and permission
	err := s.adapter.AddRole(ctx, roleID, sampleOrganization.OrganizationID, "Admin", "admin")
	if err != nil {
		t.Fatalf("Setup: AddRole failed: %v", err)
	}

	// Test adding permission to role
	err = s.adapter.AddPermissionToRole(ctx, roleID, "user:create")
	if err != nil {
		t.Fatalf("AddPermissionToRole failed: %v", err)
	}
}

func (s *AdapterExercises) TestRemovePermissionFromRole(t *testing.T) {
	ctx := t.Context()
	roleID := int64(1)

	// Setup - add permission to role first
	err := s.adapter.AddPermissionToRole(ctx, roleID, "user:delete")
	if err != nil {
		t.Fatalf("Setup: AddPermissionToRole failed: %v", err)
	}

	// Test removal
	err = s.adapter.RemovePermissionFromRole(ctx, roleID, "user:delete")
	if err != nil {
		t.Fatalf("RemovePermissionFromRole failed: %v", err)
	}
}

func (s *AdapterExercises) TestAddUserToRole(t *testing.T) {
	ctx := t.Context()
	userID := int64(1)
	roleID := int64(1)

	// Setup - create user and role
	err := s.adapter.AddUser(ctx, userID, "Test", "User", "Test User", "test@example.com")
	if err != nil {
		t.Fatalf("Setup: AddUser failed: %v", err)
	}
	err = s.adapter.AddRole(ctx, roleID, sampleOrganization.OrganizationID, "Admin", "admin")
	if err != nil {
		t.Fatalf("Setup: AddRole failed: %v", err)
	}

	// Test adding user to role
	err = s.adapter.AddUserToRole(ctx, userID, roleID)
	if err != nil {
		t.Fatalf("AddUserToRole failed: %v", err)
	}
}

func (s *AdapterExercises) TestRemoveUserFromRole(t *testing.T) {
	ctx := t.Context()
	userID := int64(1)
	roleID := int64(1)

	// Setup - add user to role first
	err := s.adapter.AddUserToRole(ctx, userID, roleID)
	if err != nil {
		t.Fatalf("Setup: AddUserToRole failed: %v", err)
	}

	// Test removal
	err = s.adapter.RemoveUserFromRole(ctx, userID, roleID)
	if err != nil {
		t.Fatalf("RemoveUserFromRole failed: %v", err)
	}
}

func (s *AdapterExercises) TestRemoveAllRolesFromUser(t *testing.T) {
	ctx := t.Context()
	userID := int64(1)
	roleID1 := int64(1)
	roleID2 := int64(2)

	// Setup - add user to multiple roles
	err := s.adapter.AddUserToRole(ctx, userID, roleID1)
	if err != nil {
		t.Fatalf("Setup: AddUserToRole 1 failed: %v", err)
	}
	err = s.adapter.AddUserToRole(ctx, userID, roleID2)
	if err != nil {
		t.Fatalf("Setup: AddUserToRole 2 failed: %v", err)
	}

	// Test removal of all roles
	err = s.adapter.RemoveAllRolesFromUser(ctx, userID)
	if err != nil {
		t.Fatalf("RemoveAllRolesFromUser failed: %v", err)
	}
}

func (s *AdapterExercises) RunAllTests(t *testing.T) {
	s.RunTests(t)
	t.Run("TestAddRole", s.TestAddRole)
	t.Run("TestAddPermissionToRole", s.TestAddPermissionToRole)
	t.Run("TestRemovePermissionFromRole", s.TestRemovePermissionFromRole)
	t.Run("TestAddUserToRole", s.TestAddUserToRole)
	t.Run("TestRemoveUserFromRole", s.TestRemoveUserFromRole)
	t.Run("TestRemoveAllRolesFromUser", s.TestRemoveAllRolesFromUser)
}
