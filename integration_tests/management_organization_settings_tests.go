package integration_tests

import (
    "context"
    "testing"

    "github.com/kernelplex/ubase/lib/ubmanage"
    "github.com/kernelplex/ubase/lib/ubstatus"
)

func (s *ManagmentServiceTestSuite) AddOrganizationSettings(t *testing.T) {
    // Add a couple of settings
    settings := map[string]string{
        "theme":  "dark",
        "region": "us-east-1",
    }

    res, err := s.managementService.OrganizationSettingsAdd(context.Background(), ubmanage.OrganizationSettingsAddCommand{
        Id:       s.createdOrganizationId,
        Settings: settings,
    }, "test-runner")
    if err != nil {
        t.Fatalf("AddOrganizationSettings failed: %v", err)
    }
    if res.Status != ubstatus.Success {
        t.Fatalf("AddOrganizationSettings status is not success: %v", res.Status)
    }

    // Verify settings are present on aggregate
    getRes, err := s.managementService.OrganizationGet(context.Background(), s.createdOrganizationId)
    if err != nil {
        t.Fatalf("AddOrganizationSettings failed to get organization: %v", err)
    }
    if getRes.Status != ubstatus.Success {
        t.Fatalf("AddOrganizationSettings get status is not success: %v", getRes.Status)
    }

    if getRes.Data.State.Settings == nil {
        t.Fatalf("AddOrganizationSettings settings map is nil")
    }
    if v, ok := getRes.Data.State.Settings["theme"]; !ok || v != "dark" {
        t.Fatalf("AddOrganizationSettings theme not set correctly: %v", v)
    }
    if v, ok := getRes.Data.State.Settings["region"]; !ok || v != "us-east-1" {
        t.Fatalf("AddOrganizationSettings region not set correctly: %v", v)
    }

    // Add another setting to ensure merge behavior
    res2, err := s.managementService.OrganizationSettingsAdd(context.Background(), ubmanage.OrganizationSettingsAddCommand{
        Id:       s.createdOrganizationId,
        Settings: map[string]string{"timezone": "UTC"},
    }, "test-runner")
    if err != nil {
        t.Fatalf("AddOrganizationSettings (merge) failed: %v", err)
    }
    if res2.Status != ubstatus.Success {
        t.Fatalf("AddOrganizationSettings (merge) status is not success: %v", res2.Status)
    }

    getRes2, err := s.managementService.OrganizationGet(context.Background(), s.createdOrganizationId)
    if err != nil {
        t.Fatalf("AddOrganizationSettings (merge) failed to get organization: %v", err)
    }
    if v, ok := getRes2.Data.State.Settings["timezone"]; !ok || v != "UTC" {
        t.Fatalf("AddOrganizationSettings timezone not set correctly: %v", v)
    }
    // Ensure previous keys remain
    if v, ok := getRes2.Data.State.Settings["theme"]; !ok || v != "dark" {
        t.Fatalf("AddOrganizationSettings theme lost after merge: %v", v)
    }
}

func (s *ManagmentServiceTestSuite) RemoveOrganizationSettings(t *testing.T) {
    // Remove one of the settings
    res, err := s.managementService.OrganizationSettingsRemove(context.Background(), ubmanage.OrganizationSettingsRemoveCommand{
        Id:          s.createdOrganizationId,
        SettingKeys: []string{"region"},
    }, "test-runner")
    if err != nil {
        t.Fatalf("RemoveOrganizationSettings failed: %v", err)
    }
    if res.Status != ubstatus.Success {
        t.Fatalf("RemoveOrganizationSettings status is not success: %v", res.Status)
    }

    // Verify removal and retention
    getRes, err := s.managementService.OrganizationGet(context.Background(), s.createdOrganizationId)
    if err != nil {
        t.Fatalf("RemoveOrganizationSettings failed to get organization: %v", err)
    }
    if getRes.Status != ubstatus.Success {
        t.Fatalf("RemoveOrganizationSettings get status is not success: %v", getRes.Status)
    }

    if _, ok := getRes.Data.State.Settings["region"]; ok {
        t.Fatalf("RemoveOrganizationSettings region was not removed")
    }
    if v, ok := getRes.Data.State.Settings["theme"]; !ok || v != "dark" {
        t.Fatalf("RemoveOrganizationSettings unexpected theme value: %v", v)
    }
    if v, ok := getRes.Data.State.Settings["timezone"]; !ok || v != "UTC" {
        t.Fatalf("RemoveOrganizationSettings unexpected timezone value: %v", v)
    }
}

