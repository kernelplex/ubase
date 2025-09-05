package integration_tests

import (
	"context"
	"testing"
	"time"

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
		Verified:    true,
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

func (s *ManagmentServiceTestSuite) TestLoginWithCorrectPassword(t *testing.T) {
	// Test login with the correct password for the updated user
	response, err := s.managementService.UserAuthenticate(context.Background(), ubmanage.UserLoginCommand{
		Email:    updatedUser.Email,
		Password: updatedUser.Password,
	}, "test-runner")

	if err != nil {
		t.Fatalf("TestLoginWithCorrectPassword failed to authenticate: %v", err)
	}

	if response.Status != ubstatus.Success {
		t.Fatalf("TestLoginWithCorrectPassword status is not success: %v", response.Status)
	}

	if response.Data == nil {
		t.Fatal("LoginWithCorrectPassword response data is nil")
	}

	if response.Data.UserId != s.createdUserId {
		t.Fatal("WithCorrectPassword user was not authenticated")
	}

	if response.Data.Email != updatedUser.Email {
		t.Fatal("LoginWithCorrectPassword email does not match")
	}

	if response.Data.RequiresTwoFactor {
		t.Fatal("WithCorrectPassword does not require two factor")
	}

}

func (s *ManagmentServiceTestSuite) LoginWithIncorrectPassword(t *testing.T) {
	// Test login with incorrect password for the updated user
	response, err := s.managementService.UserAuthenticate(context.Background(), ubmanage.UserLoginCommand{
		Email:    updatedUser.Email,
		Password: "WrongPassword123!",
	}, "test-runner")

	if err != nil {
		t.Fatalf("LoginWithIncorrectPassword failed during authentication: %v", err)
	}

	if response.Status != ubstatus.NotAuthorized {
		t.Fatalf("LoginWithIncorrectPassword expected Unauthorized status but got: %v", response.Status)
	}

	if response.Data != nil {
		t.Fatal("LoginWithIncorrectPassword should not have authenticated with wrong password")
	}
}

func (s *ManagmentServiceTestSuite) AddTwoFactorKey(t *testing.T) {

	sharedSecret := "otpauth://totp/MyIssuer:chavez@example.com?algorithm=SHA1&digits=6&issuer=MyIssuer&period=30&secret=74S6UFOJSZYSCRGTELKDDGPS6EW524ZZ"

	// Add two factor authentication to the test user

	command := ubmanage.UserSetTwoFactorSharedSecretCommand{
		Id:     s.createdUserId,
		Secret: sharedSecret,
	}
	response, err := s.managementService.UserSetTwoFactorSharedSecret(context.Background(), command, "test-runner")

	if err != nil {
		t.Fatalf("AddTwoFactorKey failed to generate shared secret: %v", err)
	}

	if response.Status != ubstatus.Success {
		t.Fatalf("AddTwoFactorKey status is not success: %v", response.Status)
	}

	s.twoFactorSecret = sharedSecret
}

func (s *ManagmentServiceTestSuite) LoginWithCorrectPasswordEnsureTwoFactorRequiredIsSet(t *testing.T) {
	// Test login with correct password after 2FA was added
	response, err := s.managementService.UserAuthenticate(context.Background(), ubmanage.UserLoginCommand{
		Email:    updatedUser.Email,
		Password: updatedUser.Password,
	}, "test-runner")

	if err != nil {
		t.Fatalf("LoginWithCorrectPasswordEnsureTwoFactorRequiredIsSet failed to authenticate: %v", err)
	}

	if response.Status != ubstatus.PartialSuccess {
		t.Fatalf("LoginWithCorrectPasswordEnsureTwoFactorRequiredIsSet status is not partial_success: %v", response.Status)
	}

	if !response.Data.RequiresTwoFactor {
		t.Fatal("LoginWithCorrectPasswordEnsureTwoFactorRequiredIsSet should require two factor after secret was added")
	}

	if response.Data.UserId != s.createdUserId {
		t.Fatal("LoginWithCorrectPasswordEnsureTwoFactorRequiredIsSet user ID does not match")
	}

	if response.Data.Email != updatedUser.Email {
		t.Fatal("LoginWithCorrectPasswordEnsureTwoFactorRequiredIsSet email does not match")
	}
}

func (s *ManagmentServiceTestSuite) VerifyCorrectTwoFactorCode(t *testing.T) {
	code, err := s.twoFactorService.GetTotpCode(s.twoFactorSecret)
	if err != nil {
		t.Fatalf("VerifyCorrectTwoFactorCode failed to generate code: %v", err)
	}

	response, err := s.managementService.UserVerifyTwoFactorCode(context.Background(), ubmanage.UserVerifyTwoFactorLoginCommand{
		UserId: s.createdUserId,
		Code:   code,
	}, "test-runner")

	if err != nil {
		t.Fatalf("VerifyCorrectTwoFactorCode failed during verification: %v", err)
	}

	if response.Status != ubstatus.Success {
		t.Fatalf("VerifyCorrectTwoFactorCode expected Success status but got: %v", response.Status)
	}
}

func (s *ManagmentServiceTestSuite) VerifyIncorrectTwoFactorCode(t *testing.T) {
	invalidResponse, err := s.managementService.UserVerifyTwoFactorCode(context.Background(), ubmanage.UserVerifyTwoFactorLoginCommand{
		UserId: s.createdUserId,
		Code:   "123456", // Invalid code
	}, "test-runner")

	if err != nil {
		t.Fatalf("VerifyIncorrectTwoFactorCode failed during verification: %v", err)
	}

	if invalidResponse.Status != ubstatus.NotAuthorized {
		t.Fatalf("VerifyIncorrectTwoFactorCode expected NotAuthorized status but got: %v", invalidResponse.Status)
	}
}

func (s *ManagmentServiceTestSuite) UserAddApiKey(t *testing.T) {
	ctx := context.Background()

	// Generate an API key using the management service
	response, err := s.managementService.UserGenerateApiKey(ctx, ubmanage.UserGenerateApiKeyCommand{
		UserId:    s.createdUserId,
		Name:      "Test API Key",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, "test-runner")

	if err != nil {
		t.Fatalf("UserGenerateApiKey failed: %v", err)
	}

	if response.Status != ubstatus.Success {
		t.Fatalf("UserGenerateApiKey status is not success: %v", response.Status)
	}

	if response.Data == "" {
		t.Fatal("UserGenerateApiKey returned empty API key")
	}

	// Verify the API key was stored by listing API keys for the user
	apiKeys, err := s.dbadapter.UserListApiKeys(ctx, s.createdUserId)
	if err != nil {
		t.Fatalf("UserListApiKeys failed: %v", err)
	}

	// Find our added API key
	found := false
	for _, key := range apiKeys {
		if key.Name == "Test API Key" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("Added API key was not found in the list")
	}
}

func (s *ManagmentServiceTestSuite) UserGetByApiKey(t *testing.T) {
	ctx := context.Background()

	// First generate an API key
	response, err := s.managementService.UserGenerateApiKey(ctx, ubmanage.UserGenerateApiKeyCommand{
		UserId:    s.createdUserId,
		Name:      "Test API Key for Get",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, "test-runner")

	if err != nil {
		t.Fatalf("UserGenerateApiKey failed: %v", err)
	}

	if response.Status != ubstatus.Success {
		t.Fatalf("UserGenerateApiKey status is not success: %v", response.Status)
	}

	if response.Data == "" {
		t.Fatal("UserGenerateApiKey returned empty API key")
	}

	t.Logf("************** Generated API Key: %s", response.Data)
	hash, _ := s.hashingService.GenerateHashBase64(response.Data)
	t.Logf("************** Generated API Key Hash: %s", hash)

	// Now try to get the user by the API key
	userResponse, err := s.managementService.UserGetByApiKey(ctx, response.Data)
	if err != nil {
		t.Fatalf("UserGetByApiKey failed: %v", err)
	}

	if userResponse.Status != ubstatus.Success {
		t.Fatalf("UserGetByApiKey status is not success: %v", userResponse.Status)
	}

	if userResponse.Data.Id != s.createdUserId {
		t.Fatalf("UserGetByApiKey returned wrong user ID: expected %d, got %d", s.createdUserId, userResponse.Data.Id)
	}

	if userResponse.Data.State.Email != updatedUser.Email {
		t.Fatalf("UserGetByApiKey returned wrong email: expected %s, got %s", updatedUser.Email, userResponse.Data.State.Email)
	}
}

func (s *ManagmentServiceTestSuite) UserDeleteApiKey(t *testing.T) {
	ctx := context.Background()

	// First generate an API key to delete
	response, err := s.managementService.UserGenerateApiKey(ctx, ubmanage.UserGenerateApiKeyCommand{
		UserId:    s.createdUserId,
		Name:      "Test API Key to Delete",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, "test-runner")

	if err != nil {
		t.Fatalf("UserGenerateApiKey failed: %v", err)
	}

	if response.Status != ubstatus.Success {
		t.Fatalf("UserGenerateApiKey status is not success: %v", response.Status)
	}

	if response.Data == "" {
		t.Fatal("UserGenerateApiKey returned empty API key")
	}

	// Delete the API key
	deleteResponse, err := s.managementService.UserDeleteApiKey(ctx, ubmanage.UserDeleteApiKeyCommand{
		UserId: s.createdUserId,
		ApiKey: response.Data, // The API key ID is the same as the key itself in this implementation
	}, "test-runner")

	if err != nil {
		t.Fatalf("UserDeleteApiKey failed: %v", err)
	}

	if deleteResponse.Status != ubstatus.Success {
		t.Fatalf("UserDeleteApiKey status is not success: %v", deleteResponse.Status)
	}

	// Verify the API key was deleted by trying to get the user by the API key
	userResponse, err := s.managementService.UserGetByApiKey(ctx, response.Data)
	if err != nil {
		t.Fatalf("UserGetByApiKey after deletion failed: %v", err)
	}

	// Should not be able to find the user with the deleted API key
	if userResponse.Status != ubstatus.NotAuthorized {
		t.Fatalf("Expected NotAuthorized status after API key deletion, got: %v", userResponse.Status)
	}

	// Get the user by ID.
	userByIdResponse, err := s.managementService.UserGetById(ctx, s.createdUserId)
	if err != nil {
		t.Fatalf("UserGetById failed: %v", err)
	}

	// Ensure the api key is not in the user's API keys
	for _, apiKey := range userByIdResponse.Data.State.ApiKeys {
		if apiKey.Name == "Test API Key to Delete" {
			t.Fatal("Deleted API key was still found in user's API keys")
		}
	}
}
