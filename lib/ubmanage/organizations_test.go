package ubmanage

import (
    "testing"
)

func TestOrganizationCreateCommandValidate(t *testing.T) {
    tests := []struct {
        name string
        cmd  OrganizationCreateCommand
        ok   bool
    }{
        {"valid", OrganizationCreateCommand{Name: "Org", SystemName: "Org_1", Status: "active"}, true},
        {"missing name", OrganizationCreateCommand{Name: "", SystemName: "Org_1", Status: "active"}, false},
        {"missing systemName", OrganizationCreateCommand{Name: "Org", SystemName: "", Status: "active"}, false},
        {"missing status", OrganizationCreateCommand{Name: "Org", SystemName: "Org_1", Status: ""}, false},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ok, _ := tt.cmd.Validate()
            if ok != tt.ok {
                t.Fatalf("expected ok=%v, got %v", tt.ok, ok)
            }
        })
    }
}

func TestOrganizationUpdateCommandValidate(t *testing.T) {
    name := "NewName"
    sys := "New_System"
    status := "active"

    tests := []struct {
        name string
        cmd  OrganizationUpdateCommand
        ok   bool
    }{
        {"valid all fields", OrganizationUpdateCommand{Id: 1, Name: &name, SystemName: &sys, Status: &status}, true},
        {"valid only id", OrganizationUpdateCommand{Id: 1}, true},
        {"invalid id 0", OrganizationUpdateCommand{Id: 0, Name: &name}, false},
        {"empty optional name", OrganizationUpdateCommand{Id: 2, Name: ptr("")}, false},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ok, _ := tt.cmd.Validate()
            if ok != tt.ok {
                t.Fatalf("expected ok=%v, got %v", tt.ok, ok)
            }
        })
    }
}

func ptr(s string) *string { return &s }

