package ubmanage

import (
    "testing"
)

func TestUserRolesCommandValidation(t *testing.T) {
    add := UserAddToRoleCommand{UserId: 1, RoleId: 2}
    if ok, _ := add.Validate(); !ok { t.Fatal("expected valid add to role") }
    add = UserAddToRoleCommand{UserId: 0, RoleId: 2}
    if ok, _ := add.Validate(); ok { t.Fatal("expected invalid add to role (userId)") }

    rem := UserRemoveFromRoleCommand{UserId: 1, RoleId: 2}
    if ok, _ := rem.Validate(); !ok { t.Fatal("expected valid remove from role") }
    rem = UserRemoveFromRoleCommand{UserId: 1, RoleId: 0}
    if ok, _ := rem.Validate(); ok { t.Fatal("expected invalid remove from role (roleId)") }
}

