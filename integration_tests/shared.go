package integration_tests

import (
	"context"
	"database/sql"
	"log/slog"
	"strings"
	"testing"

	evercore "github.com/kernelplex/evercore/base"
	ubase "github.com/kernelplex/ubase/lib"
	"github.com/kernelplex/ubase/lib/dbinterface"
	"github.com/kernelplex/ubase/lib/ubconst"
	"github.com/kernelplex/ubase/lib/ubsecurity"
)

type StorageEngineTestSuite struct {
	eventStore        *evercore.EventStore
	db                *sql.DB
	dbType            ubconst.DatabaseType
	roleService       ubase.RoleService
	userService       ubase.UserService
	permissionService ubase.PermissionService
	existinguserId    int64
}

func NewStorageEngineTestSuite(eventStore *evercore.EventStore, db *sql.DB, dbType ubconst.DatabaseType) *StorageEngineTestSuite {
	return &StorageEngineTestSuite{
		eventStore: eventStore,
		db:         db,
		dbType:     dbType,
	}
}

func (s *StorageEngineTestSuite) RunTests(t *testing.T) {
	dbadapter := dbinterface.NewDatabase(s.dbType, s.db)

	hashingService := ubsecurity.DefaultArgon2Id
	s.userService = ubase.CreateUserService(s.eventStore, hashingService, dbadapter)
	s.roleService = ubase.CreateRoleService(s.eventStore, dbadapter)
	s.permissionService = ubase.NewPermissionService(dbadapter, s.roleService)

	t.Run("CreateUser", s.CreateUser)
	t.Run("CreateUser_WithDuplicateEmailFails", s.CreateUser_WithDuplicateEmailFails)
	t.Run("UpdateUser", s.UpdateUser)
	/*
		t.Run("CreateRole", s.CreateRole)
		t.Run("AddPermissionToRole", s.AddPermissionToRole)
		t.Run("RemovePermissionFromRole", s.RemovePermissionFromRole)
		t.Run("GetRolePermissions", s.GetRolePermissions)
		t.Run("UpdateRole", s.UpdateRole)
		t.Run("DeleteRole", s.DeleteRole)
	*/
}

func (s *StorageEngineTestSuite) CreateUser(t *testing.T) {
	ctx := context.Background()
	testEmail := "testuser@example.com"

	// Create test user
	resp, err := s.userService.CreateUser(ctx, ubase.UserCreateCommand{
		Email:       testEmail,
		Password:    "SecurePassword123!",
		FirstName:   "Test",
		LastName:    "User",
		DisplayName: "Test User",
	}, "test")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Verify response
	if resp.Id <= 0 {
		slog.Error("Response", "response", resp)
		t.Errorf("Expected positive user ID, got %d", resp.Id)
	}
	if resp.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", resp.Status)
	}
	if resp.ValidationIssues != nil {
		slog.Error("Validation issues", "issues", resp.ValidationIssues)
		t.Errorf("Expected no validation issues, got %v", resp.ValidationIssues)
	}
	s.existinguserId = resp.Id
}

func (s *StorageEngineTestSuite) CreateUser_WithDuplicateEmailFails(t *testing.T) {
	ctx := context.Background()
	testEmail := "duplicate@example.com"

	// First create a user with this email
	_, err := s.userService.CreateUser(ctx, ubase.UserCreateCommand{
		Email:       testEmail,
		Password:    "Password123!",
		FirstName:   "First",
		LastName:    "User",
		DisplayName: "First User",
	}, "test")
	if err != nil {
		t.Fatalf("Failed to create initial user: %v", err)
	}

	// Try to create another user with same email
	_, err = s.userService.CreateUser(ctx, ubase.UserCreateCommand{
		Email:       testEmail,
		Password:    "Password456!",
		FirstName:   "Second",
		LastName:    "User",
		DisplayName: "Second User",
	}, "test")

	// Verify duplicate email fails
	if err == nil {
		t.Error("Expected error when creating user with duplicate email, got nil")
		return
	}

	if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
		t.Errorf("Expected duplicate email error, got: %v", err)
		return
	}
}

func (s *StorageEngineTestSuite) UpdateUser(t *testing.T) {
	ctx := context.Background()

	// First create a test user if none exists
	if s.existinguserId == 0 {
		s.CreateUser(t)
	}

	newFirstName := "UpdatedFirst"
	newLastName := "UpdatedLast"
	newDisplayName := "Updated Display"

	// Update the user
	resp, err := s.userService.UpdateUser(ctx, ubase.UserUpdateCommand{
		Id:          s.existinguserId,
		FirstName:   &newFirstName,
		LastName:    &newLastName,
		DisplayName: &newDisplayName,
	}, "test")
	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	// Verify response
	if resp.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", resp.Status)
	}

	// Verify the updates were applied
	userAgg, err := s.userService.GetUserByEmail(ctx, "testuser@example.com")
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	if userAgg.State.FirstName != newFirstName {
		t.Errorf("FirstName not updated, expected '%s' got '%s'", newFirstName, userAgg.State.FirstName)
	}
	if userAgg.State.LastName != newLastName {
		t.Errorf("LastName not updated, expected '%s' got '%s'", newLastName, userAgg.State.LastName)
	}
	if userAgg.State.DisplayName != newDisplayName {
		t.Errorf("DisplayName not updated, expected '%s' got '%s'", newDisplayName, userAgg.State.DisplayName)
	}
}

/*

func (s *StorageEngineTestSuite) CreateRole(t *testing.T) {
	panic("implement me")
}

func (s *StorageEngineTestSuite) AddPermissionToRole(t *testing.T) {
	panic("implement me")
}

func (s *StorageEngineTestSuite) RemovePermissionFromRole(t *testing.T) {
	panic("implement me")
}

func (s *StorageEngineTestSuite) GetRolePermissions(t *testing.T) {
	panic("implement me")
}

func (s *StorageEngineTestSuite) UpdateRole(t *testing.T) {
	panic("implement me")
}

func (s *StorageEngineTestSuite) DeleteRole(t *testing.T) {
	panic("implement me")
}
*/
