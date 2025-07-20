package integration_tests

import (
	"context"
	"testing"

	"github.com/kernelplex/ubase/lib/ubmanage"
	"github.com/kernelplex/ubase/lib/ubstatus"
)

type UserTestStruct struct {
	Email       string `json:"email"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	DisplayName string `json:"displayName"`
	Password    string `json:"password"`
}

var aUser = UserTestStruct{
	Email:       "test@example.com",
	FirstName:   "Test",
	LastName:    "User",
	DisplayName: "Test User",
	Password:    "TestPassword123!",
}

func (s *ManagmentServiceTestSuite) AddUser(t *testing.T) {
	response, err := s.managementService.UserAdd(context.Background(), ubmanage.UserCreateCommand{
		Email:       aUser.Email,
		Password:    "TestPassword123!",
		FirstName:   aUser.FirstName,
		LastName:    aUser.LastName,
		DisplayName: aUser.DisplayName,
	}, "test-runner")

	if err != nil {
		t.Fatalf("AddUser failed to add user: %v", err)
	}

	if response.Status != ubstatus.Success {
		t.Fatalf("AddUser status is not success: %v", response.Status)
	}

	// Verify the user was added to the database
	user, err := s.dbadapter.GetUserByEmail(context.Background(), aUser.Email)
	if err != nil {
		t.Fatalf("AddUser failed to get user by email: %v", err)
	}

	if user.Email != aUser.Email {
		t.Fatalf("AddUser email does not match: %v", user.Email)
	}
	if user.FirstName != aUser.FirstName {
		t.Fatalf("AddUser first name does not match: %v", user.FirstName)
	}
	if user.LastName != aUser.LastName {
		t.Fatalf("AddUser last name does not match: %v", user.LastName)
	}
	if user.DisplayName != aUser.DisplayName {
		t.Fatalf("AddUser display name does not match: %v", user.DisplayName)
	}

	s.createdUserId = response.Data.Id
}

func (s *ManagmentServiceTestSuite) GetUserByEmail(t *testing.T) {

	// Now test getting the user by email
	response, err := s.managementService.UserGetByEmail(context.Background(), aUser.Email)
	if err != nil {
		t.Fatalf("GetUserByEmail failed to get user by email: %v", err)
	}

	s.userPasswordHash = response.Data.State.PasswordHash

	if response.Status != ubstatus.Success {
		t.Fatalf("GetUserByEmail status is not success: %v", response.Status)
	}

	if response.Data.State.Email != aUser.Email {
		t.Fatalf("GetUserByEmail email does not match: %v", response.Data.State.Email)
	}
	if response.Data.State.FirstName != aUser.FirstName {
		t.Fatalf("GetUserByEmail first name does not match: %v", response.Data.State.FirstName)
	}
	if response.Data.State.LastName != aUser.LastName {
		t.Fatalf("GetUserByEmail last name does not match: %v", response.Data.State.LastName)
	}
	if response.Data.State.DisplayName != aUser.DisplayName {
		t.Fatalf("GetUserByEmail display name does not match: %v", response.Data.State.DisplayName)
	}
}

var updatedUser = UserTestStruct{
	Email:       "updated@example.com",
	FirstName:   "Updated",
	LastName:    "User",
	DisplayName: "Updated User",
	Password:    "UpdatedPassword123!",
}

func (s *ManagmentServiceTestSuite) UpdateUser(t *testing.T) {

	// Update the user
	res, err := s.managementService.UserUpdate(context.Background(), ubmanage.UserUpdateCommand{
		Id:          s.createdUserId,
		Email:       &updatedUser.Email,
		FirstName:   &updatedUser.FirstName,
		LastName:    &updatedUser.LastName,
		DisplayName: &updatedUser.DisplayName,
		Password:    &updatedUser.Password,
	}, "test-runner")

	if err != nil {
		t.Fatalf("UpdateUser failed to update user: %v", err)
	}
	if res.Status != ubstatus.Success {
		t.Fatalf("UpdateUser status is not success: %v", res.Status)
	}

	// Verify the user was updated in the database
	user, err := s.dbadapter.GetUserByEmail(context.Background(), updatedUser.Email)
	if err != nil {
		t.Fatalf("UpdateUser failed to get updated user by email: %v", err)
	}

	if user.Email != updatedUser.Email {
		t.Fatalf("UpdateUser email does not match: %v", user.Email)
	}
	if user.FirstName != updatedUser.FirstName {
		t.Fatalf("UpdateUser first name does not match: %v", user.FirstName)
	}
	if user.LastName != updatedUser.LastName {
		t.Fatalf("UpdateUser last name does not match: %v", user.LastName)
	}
	if user.DisplayName != updatedUser.DisplayName {
		t.Fatalf("UpdateUser display name does not match: %v", user.DisplayName)
	}

}

func (s *ManagmentServiceTestSuite) GetUserByEmailPostUpdate(t *testing.T) {
	// Test getting the updated user by email
	response, err := s.managementService.UserGetByEmail(context.Background(), updatedUser.Email)
	if err != nil {
		t.Fatalf("GetUserByEmailPostUpdate failed to get user by email: %v", err)
	}

	if response.Status != ubstatus.Success {
		t.Fatalf("GetUserByEmailPostUpdate status is not success: %v", response.Status)
	}

	if response.Data.State.Email != updatedUser.Email {
		t.Fatalf("GetUserByEmailPostUpdate email does not match: %v", response.Data.State.Email)
	}
	if response.Data.State.FirstName != updatedUser.FirstName {
		t.Fatalf("GetUserByEmailPostUpdate first name does not match: %v", response.Data.State.FirstName)
	}
	if response.Data.State.LastName != updatedUser.LastName {
		t.Fatalf("GetUserByEmailPostUpdate last name does not match: %v", response.Data.State.LastName)
	}
	if response.Data.State.DisplayName != updatedUser.DisplayName {
		t.Fatalf("GetUserByEmailPostUpdate display name does not match: %v", response.Data.State.DisplayName)
	}

	if response.Data.State.PasswordHash == s.userPasswordHash {
		t.Fatalf("UpdateUser password hash is not updated.")
	}
}

func (s *ManagmentServiceTestSuite) UpdateUserWithoutPassword(t *testing.T) {
	// First ensure test user exists
	_, err := s.managementService.UserAdd(context.Background(), ubmanage.UserCreateCommand{
		Email:       "nopassupdate@example.com",
		Password:    "TestPassword123!",
		FirstName:   "NoPass",
		LastName:    "Update",
		DisplayName: "NoPass Update",
	}, "test-runner")
	if err != nil {
		t.Fatalf("UpdateUserWithoutPassword failed to add test user: %v", err)
	}

	// Get initial password hash
	initialResponse, err := s.managementService.UserGetByEmail(context.Background(), "nopassupdate@example.com")
	if err != nil {
		t.Fatalf("UpdateUserWithoutPassword failed to get initial user: %v", err)
	}
	initialHash := initialResponse.Data.State.PasswordHash

	// Update user without password
	res, err := s.managementService.UserUpdate(context.Background(), ubmanage.UserUpdateCommand{
		Id:          initialResponse.Data.Id,
		FirstName:   ptr("UpdatedFirst"),
		LastName:    ptr("UpdatedLast"),
		DisplayName: ptr("Updated Display"),
	}, "test-runner")

	if err != nil {
		t.Fatalf("UpdateUserWithoutPassword failed to update user: %v", err)
	}
	if res.Status != ubstatus.Success {
		t.Fatalf("UpdateUserWithoutPassword status is not success: %v", res.Status)
	}

	// Verify password hash wasn't changed
	updatedResponse, err := s.managementService.UserGetByEmail(context.Background(), "nopassupdate@example.com")
	if err != nil {
		t.Fatalf("UpdateUserWithoutPassword failed to get updated user: %v", err)
	}

	if updatedResponse.Data.State.PasswordHash != initialHash {
		t.Fatalf("UpdateUserWithoutPassword password hash was unexpectedly changed")
	}
}

func ptr(s string) *string {
	return &s
}

func (s *ManagmentServiceTestSuite) UpdateUserSamePassword(t *testing.T) {
	// First ensure test user exists
	_, err := s.managementService.UserAdd(context.Background(), ubmanage.UserCreateCommand{
		Email:       "samepass@example.com",
		Password:    "SamePassword123!",
		FirstName:   "SamePass",
		LastName:    "User",
		DisplayName: "SamePass User",
	}, "test-runner")
	if err != nil {
		t.Fatalf("UpdateUserSamePassword failed to add test user: %v", err)
	}

	// Get initial password hash
	initialResponse, err := s.managementService.UserGetByEmail(context.Background(), "samepass@example.com")
	if err != nil {
		t.Fatalf("UpdateUserSamePassword failed to get initial user: %v", err)
	}
	initialHash := initialResponse.Data.State.PasswordHash

	// Update user with same password
	res, err := s.managementService.UserUpdate(context.Background(), ubmanage.UserUpdateCommand{
		Id:       initialResponse.Data.Id,
		Password: ptr("SamePassword123!"),
	}, "test-runner")

	if err != nil {
		t.Fatalf("UpdateUserSamePassword failed to update user: %v", err)
	}
	if res.Status != ubstatus.Success {
		t.Fatalf("UpdateUserSamePassword status is not success: %v", res.Status)
	}

	// Verify password hash changed (since salts should be regenerated)
	updatedResponse, err := s.managementService.UserGetByEmail(context.Background(), "samepass@example.com")
	if err != nil {
		t.Fatalf("UpdateUserSamePassword failed to get updated user: %v", err)
	}

	if updatedResponse.Data.State.PasswordHash == initialHash {
		t.Fatalf("UpdateUserSamePassword password hash was not changed as expected")
	}
}
