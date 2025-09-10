package ubmanage

import (
    "testing"
    "time"

    evercore "github.com/kernelplex/evercore/base"
)

func TestRoleAggregateApplyEventState(t *testing.T) {
    agg := &RoleAggregate{}

    // Seed created state via generic state event to satisfy ApplyEventState default branch
    created := evercore.NewStateEvent(RoleCreatedEvent{Id: 1, OrganizationId: 2, Name: "Admin", SystemName: "admin"})
    if err := agg.ApplyEventState(created, time.Now(), "tester"); err != nil {
        t.Fatalf("apply created: %v", err)
    }

    // Add permission
    if err := agg.ApplyEventState(RolePermissionAddedEvent{Permission: "users:read"}, time.Now(), "tester"); err != nil {
        t.Fatalf("add perm: %v", err)
    }
    if len(agg.State.Permissions) != 1 || agg.State.Permissions[0] != "users:read" {
        t.Fatalf("expected one permission users:read, got %+v", agg.State.Permissions)
    }

    // Remove permission - not present should be no-op
    if err := agg.ApplyEventState(RolePermissionRemovedEvent{Permission: "users:write"}, time.Now(), "tester"); err != nil {
        t.Fatalf("remove non-existing perm: %v", err)
    }
    if len(agg.State.Permissions) != 1 {
        t.Fatalf("expected permissions unchanged, got %+v", agg.State.Permissions)
    }

    // Remove existing permission
    if err := agg.ApplyEventState(RolePermissionRemovedEvent{Permission: "users:read"}, time.Now(), "tester"); err != nil {
        t.Fatalf("remove existing perm: %v", err)
    }
    if len(agg.State.Permissions) != 0 {
        t.Fatalf("expected permissions empty, got %+v", agg.State.Permissions)
    }

    // Delete + undelete
    if err := agg.ApplyEventState(RoleDeletedEvent{}, time.Now(), "tester"); err != nil {
        t.Fatalf("delete: %v", err)
    }
    if !agg.State.Deleted {
        t.Fatal("expected deleted=true")
    }
    if err := agg.ApplyEventState(RoleUndeletedEvent{}, time.Now(), "tester"); err != nil {
        t.Fatalf("undelete: %v", err)
    }
    if agg.State.Deleted {
        t.Fatal("expected deleted=false")
    }
}

func TestRoleCommandValidation(t *testing.T) {
    // RoleCreateCommand
    rc := RoleCreateCommand{Name: "Name", SystemName: "System_1", OrganizationId: 1}
    if ok, _ := rc.Validate(); !ok { t.Fatal("expected valid create") }
    if ok, _ := (RoleCreateCommand{}).Validate(); ok { t.Fatal("expected invalid create") }

    // RoleUpdateCommand
    name := "New"
    sys := "Sys_2"
    ru := RoleUpdateCommand{Id: 1, Name: &name, SystemName: &sys}
    if ok, _ := ru.Validate(); !ok { t.Fatal("expected valid update") }
    ru = RoleUpdateCommand{Id: 0, Name: &name}
    if ok, _ := ru.Validate(); ok { t.Fatal("expected invalid update") }

    // Permission add/remove commands
    add := RolePermissionAddCommand{Id: 1, Permission: "users:read"}
    if ok, _ := add.Validate(); !ok { t.Fatal("expected valid add permission") }
    add = RolePermissionAddCommand{Id: 1, Permission: ""}
    if ok, _ := add.Validate(); ok { t.Fatal("expected invalid add permission") }

    rem := RolePermissionRemoveCommand{Id: 1, Permission: "users:read"}
    if ok, _ := rem.Validate(); !ok { t.Fatal("expected valid remove permission") }
    rem = RolePermissionRemoveCommand{Id: 0, Permission: "users:read"}
    if ok, _ := rem.Validate(); ok { t.Fatal("expected invalid remove permission (id)") }
}

