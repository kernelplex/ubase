package integration_tests

import (
    "context"
    "testing"

    "github.com/kernelplex/ubase/lib/ubmanage"
    "github.com/kernelplex/ubase/lib/ubstatus"
)

func (s *ManagmentServiceTestSuite) AddOrganizationSettingsValidation(t *testing.T) {
    // Invalid: empty settings and invalid id
    res, err := s.managementService.OrganizationSettingsAdd(context.Background(), ubmanage.OrganizationSettingsAddCommand{
        Id:       0,
        Settings: map[string]string{},
    }, "test-runner")
    if err != nil {
        t.Fatalf("AddOrganizationSettingsValidation unexpected error: %v", err)
    }
    if res.Status != ubstatus.ValidationError {
        t.Fatalf("AddOrganizationSettingsValidation expected validation error, got: %v", res.Status)
    }

    // Invalid: contains empty key
    res2, err := s.managementService.OrganizationSettingsAdd(context.Background(), ubmanage.OrganizationSettingsAddCommand{
        Id: s.createdOrganizationId,
        Settings: map[string]string{
            "": "value",
        },
    }, "test-runner")
    if err != nil {
        t.Fatalf("AddOrganizationSettingsValidation unexpected error (empty key): %v", err)
    }
    if res2.Status != ubstatus.ValidationError {
        t.Fatalf("AddOrganizationSettingsValidation expected validation error for empty key, got: %v", res2.Status)
    }
}

func (s *ManagmentServiceTestSuite) RemoveOrganizationSettingsValidation(t *testing.T) {
    // Invalid: empty keys and invalid id
    res, err := s.managementService.OrganizationSettingsRemove(context.Background(), ubmanage.OrganizationSettingsRemoveCommand{
        Id:          0,
        SettingKeys: []string{},
    }, "test-runner")
    if err != nil {
        t.Fatalf("RemoveOrganizationSettingsValidation unexpected error: %v", err)
    }
    if res.Status != ubstatus.ValidationError {
        t.Fatalf("RemoveOrganizationSettingsValidation expected validation error, got: %v", res.Status)
    }

    // Invalid: contains empty key
    res2, err := s.managementService.OrganizationSettingsRemove(context.Background(), ubmanage.OrganizationSettingsRemoveCommand{
        Id:          s.createdOrganizationId,
        SettingKeys: []string{""},
    }, "test-runner")
    if err != nil {
        t.Fatalf("RemoveOrganizationSettingsValidation unexpected error (empty key): %v", err)
    }
    if res2.Status != ubstatus.ValidationError {
        t.Fatalf("RemoveOrganizationSettingsValidation expected validation error for empty key, got: %v", res2.Status)
    }
}

