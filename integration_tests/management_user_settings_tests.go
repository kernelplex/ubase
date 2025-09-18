package integration_tests

import (
    "context"
    "testing"

    "github.com/kernelplex/ubase/lib/ubmanage"
    "github.com/kernelplex/ubase/lib/ubstatus"
)

func (s *ManagmentServiceTestSuite) AddUserSettings(t *testing.T) {
    res, err := s.managementService.UserSettingsAdd(context.Background(), ubmanage.UserSettingsAddCommand{
        Id: s.createdUserId,
        Settings: map[string]string{
            "newsletter": "subscribed",
            "theme":      "light",
        },
    }, "test-runner")
    if err != nil {
        t.Fatalf("AddUserSettings failed: %v", err)
    }
    if res.Status != ubstatus.Success {
        t.Fatalf("AddUserSettings status is not success: %v", res.Status)
    }

    // Verify via aggregate
    getRes, err := s.managementService.UserGetById(context.Background(), s.createdUserId)
    if err != nil {
        t.Fatalf("AddUserSettings get user failed: %v", err)
    }
    if getRes.Status != ubstatus.Success {
        t.Fatalf("AddUserSettings get status is not success: %v", getRes.Status)
    }
    if v, ok := getRes.Data.State.Settings["newsletter"]; !ok || v != "subscribed" {
        t.Fatalf("AddUserSettings newsletter not set correctly: %v", v)
    }
    if v, ok := getRes.Data.State.Settings["theme"]; !ok || v != "light" {
        t.Fatalf("AddUserSettings theme not set correctly: %v", v)
    }

    // Merge in another setting
    res2, err := s.managementService.UserSettingsAdd(context.Background(), ubmanage.UserSettingsAddCommand{
        Id:       s.createdUserId,
        Settings: map[string]string{"timezone": "PST"},
    }, "test-runner")
    if err != nil {
        t.Fatalf("AddUserSettings merge failed: %v", err)
    }
    if res2.Status != ubstatus.Success {
        t.Fatalf("AddUserSettings merge status is not success: %v", res2.Status)
    }
    getRes2, err := s.managementService.UserGetById(context.Background(), s.createdUserId)
    if err != nil {
        t.Fatalf("AddUserSettings merge get failed: %v", err)
    }
    if v, ok := getRes2.Data.State.Settings["timezone"]; !ok || v != "PST" {
        t.Fatalf("AddUserSettings timezone not set correctly: %v", v)
    }
}

func (s *ManagmentServiceTestSuite) RemoveUserSettings(t *testing.T) {
    res, err := s.managementService.UserSettingsRemove(context.Background(), ubmanage.UserSettingsRemoveCommand{
        Id:          s.createdUserId,
        SettingKeys: []string{"newsletter"},
    }, "test-runner")
    if err != nil {
        t.Fatalf("RemoveUserSettings failed: %v", err)
    }
    if res.Status != ubstatus.Success {
        t.Fatalf("RemoveUserSettings status is not success: %v", res.Status)
    }
    // Verify removal
    getRes, err := s.managementService.UserGetById(context.Background(), s.createdUserId)
    if err != nil {
        t.Fatalf("RemoveUserSettings get failed: %v", err)
    }
    if _, ok := getRes.Data.State.Settings["newsletter"]; ok {
        t.Fatalf("RemoveUserSettings newsletter was not removed")
    }
    if v, ok := getRes.Data.State.Settings["theme"]; !ok || v != "light" {
        t.Fatalf("RemoveUserSettings theme should remain: %v", v)
    }
}

func (s *ManagmentServiceTestSuite) AddUserSettingsValidation(t *testing.T) {
    // invalid id and empty settings
    res, err := s.managementService.UserSettingsAdd(context.Background(), ubmanage.UserSettingsAddCommand{Id: 0, Settings: map[string]string{}}, "test-runner")
    if err != nil {
        t.Fatalf("AddUserSettingsValidation unexpected error: %v", err)
    }
    if res.Status != ubstatus.ValidationError {
        t.Fatalf("AddUserSettingsValidation expected validation error, got: %v", res.Status)
    }
    // empty key
    res2, err := s.managementService.UserSettingsAdd(context.Background(), ubmanage.UserSettingsAddCommand{Id: s.createdUserId, Settings: map[string]string{"": "x"}}, "test-runner")
    if err != nil {
        t.Fatalf("AddUserSettingsValidation unexpected error 2: %v", err)
    }
    if res2.Status != ubstatus.ValidationError {
        t.Fatalf("AddUserSettingsValidation expected validation error for empty key, got: %v", res2.Status)
    }
}

func (s *ManagmentServiceTestSuite) RemoveUserSettingsValidation(t *testing.T) {
    // invalid id and empty keys
    res, err := s.managementService.UserSettingsRemove(context.Background(), ubmanage.UserSettingsRemoveCommand{Id: 0, SettingKeys: []string{}}, "test-runner")
    if err != nil {
        t.Fatalf("RemoveUserSettingsValidation unexpected error: %v", err)
    }
    if res.Status != ubstatus.ValidationError {
        t.Fatalf("RemoveUserSettingsValidation expected validation error, got: %v", res.Status)
    }
    // empty key value
    res2, err := s.managementService.UserSettingsRemove(context.Background(), ubmanage.UserSettingsRemoveCommand{Id: s.createdUserId, SettingKeys: []string{""}}, "test-runner")
    if err != nil {
        t.Fatalf("RemoveUserSettingsValidation unexpected error 2: %v", err)
    }
    if res2.Status != ubstatus.ValidationError {
        t.Fatalf("RemoveUserSettingsValidation expected validation error for empty key, got: %v", res2.Status)
    }
}

