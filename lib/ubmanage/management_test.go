package ubmanage

import (
    "context"
    "io"
    "testing"
    "time"

    evercore "github.com/kernelplex/evercore/base"
    "github.com/kernelplex/ubase/lib/ubdata"
    "github.com/kernelplex/ubase/lib/ubstatus"
)

// Minimal fakes to satisfy interfaces without full behavior
type fakeDB struct{}

func (f *fakeDB) AddUser(ctx context.Context, userID int64, firstName, lastName, displayName, email string, verified bool, createdAt int64, updatedAt int64) error {
    return nil
}
func (f *fakeDB) GetUser(ctx context.Context, userID int64) (ubdata.User, error) { return ubdata.User{}, nil }
func (f *fakeDB) GetUserByEmail(ctx context.Context, email string) (ubdata.User, error) { return ubdata.User{}, nil }
func (f *fakeDB) SearchUsers(ctx context.Context, searchTerm string, limit, offset int) ([]ubdata.User, error) {
    return nil, nil
}
func (f *fakeDB) UpdateUser(ctx context.Context, userID int64, firstName, lastName, displayName, email string, verified bool, updatedAt int64) error {
    return nil
}
func (f *fakeDB) AddOrganization(ctx context.Context, id int64, name string, systemName string, status string) error { return nil }
func (f *fakeDB) GetOrganization(ctx context.Context, organizationID int64) (ubdata.Organization, error) {
    return ubdata.Organization{}, nil
}
func (f *fakeDB) ListOrganizations(ctx context.Context) ([]ubdata.Organization, error) { return nil, nil }
func (f *fakeDB) GetOrganizationBySystemName(ctx context.Context, systemName string) (ubdata.Organization, error) {
    return ubdata.Organization{}, nil
}
func (f *fakeDB) UpdateOrganization(ctx context.Context, id int64, name string, systemName string, status string) error { return nil }
func (f *fakeDB) AddRole(ctx context.Context, roleID int64, organizationID int64, name string, systemName string) error { return nil }
func (f *fakeDB) UpdateRole(ctx context.Context, roleID int64, name string, systemName string) error { return nil }
func (f *fakeDB) DeleteRole(ctx context.Context, roleID int64) error { return nil }
func (f *fakeDB) GetOrganizationRoles(ctx context.Context, organizationID int64) ([]ubdata.RoleRow, error) { return nil, nil }
func (f *fakeDB) AddPermissionToRole(ctx context.Context, roleID int64, permission string) error { return nil }
func (f *fakeDB) RemovePermissionFromRole(ctx context.Context, roleID int64, permission string) error { return nil }
func (f *fakeDB) GetRolePermissions(ctx context.Context, roleID int64) ([]string, error) { return nil, nil }
func (f *fakeDB) AddUserToRole(ctx context.Context, userID int64, roleID int64) error { return nil }
func (f *fakeDB) RemoveUserFromRole(ctx context.Context, userID int64, roleID int64) error { return nil }
func (f *fakeDB) RemoveAllRolesFromUser(ctx context.Context, userID int64) error { return nil }
func (f *fakeDB) GetUserOrganizationRoles(ctx context.Context, userID int64, organizationId int64) ([]ubdata.RoleRow, error) {
    return nil, nil
}
func (f *fakeDB) GetAllUserOrganizationRoles(ctx context.Context, userID int64) ([]ubdata.ListUserOrganizationRolesRow, error) { return nil, nil }
func (f *fakeDB) ListUserOrganizationRoles(ctx context.Context, userID int64) ([]ubdata.ListUserOrganizationRolesRow, error) { return nil, nil }
func (f *fakeDB) OrganizationsCount(ctx context.Context) (int64, error) { return 0, nil }
func (f *fakeDB) UsersCount(ctx context.Context) (int64, error) { return 0, nil }
func (f *fakeDB) UpdateUserLoginStats(ctx context.Context, userID int64, lastLogin int64, loginCount int64) error { return nil }
func (f *fakeDB) ListOrganizationsRolesWithUserCounts(ctx context.Context, organizationId int64) ([]ubdata.ListRolesWithUserCountsRow, error) {
    return nil, nil
}
func (f *fakeDB) GetUsersInRole(ctx context.Context, roleID int64) ([]ubdata.User, error) { return nil, nil }
func (f *fakeDB) GetRolesForUser(ctx context.Context, userID int64) ([]ubdata.RoleRow, error) { return nil, nil }
func (f *fakeDB) UserAddApiKey(ctx context.Context, userID int64, organizationId int64, apiKeyId string, apiKeyHash, name string, createdAt time.Time, expiresAt time.Time) error {
    return nil
}
func (f *fakeDB) UserDeleteApiKey(ctx context.Context, userID int64, apiKeyId string) error { return nil }
func (f *fakeDB) UserListApiKeys(ctx context.Context, userID int64) ([]ubdata.UserApiKeyNoHash, error) { return nil, nil }
func (f *fakeDB) UserGetApiKey(ctx context.Context, apiKeyId string) (ubdata.UserApiKeyWithHash, error) {
    return ubdata.UserApiKeyWithHash{}, nil
}

type fakeHasher struct{}
func (f *fakeHasher) GenerateHashBytes(s string) []byte             { return nil }
func (f *fakeHasher) GenerateHashBase64(s string) (string, error)   { return "", nil }
func (f *fakeHasher) VerifyBase64(a, b string) (bool, error)        { return false, nil }
func (f *fakeHasher) Verify(a, b []byte) bool                       { return false }

type fakeEnc struct{}
func (f *fakeEnc) Encrypt(data []byte) ([]byte, error)  { return nil, nil }
func (f *fakeEnc) Decrypt(data []byte) ([]byte, error)  { return nil, nil }
func (f *fakeEnc) Encrypt64(s string) (string, error)   { return "", nil }
func (f *fakeEnc) Decrypt64(s string) ([]byte, error)   { return nil, nil }

type fakeTotp struct{}
func (f *fakeTotp) GenerateTotp(accountName string) (string, error) { return "", nil }
func (f *fakeTotp) ValidateTotp(url string, code string) (bool, error) { return false, nil }
func (f *fakeTotp) GenerateTotpPng(w io.Writer, url string) error { return nil }
func (f *fakeTotp) GetTotpCode(url string) (string, error) { return "", nil }

func TestNewManagementPanicsOnNilDeps(t *testing.T) {
    // Each dependency nil should trigger a panic
    mustPanic := func(f func()) {
        t.Helper()
        defer func() { if r := recover(); r == nil { t.Fatalf("expected panic") } }()
        f()
    }

    // nil store
    mustPanic(func() { _ = NewManagement(nil, &fakeDB{}, &fakeHasher{}, &fakeEnc{}, &fakeTotp{}) })
    // nil db
    mustPanic(func() { _ = NewManagement(&evercore.EventStore{}, nil, &fakeHasher{}, &fakeEnc{}, &fakeTotp{}) })
    // nil hasher
    mustPanic(func() { _ = NewManagement(&evercore.EventStore{}, &fakeDB{}, nil, &fakeEnc{}, &fakeTotp{}) })
    // nil enc
    mustPanic(func() { _ = NewManagement(&evercore.EventStore{}, &fakeDB{}, &fakeHasher{}, nil, &fakeTotp{}) })
    // nil totp
    mustPanic(func() { _ = NewManagement(&evercore.EventStore{}, &fakeDB{}, &fakeHasher{}, &fakeEnc{}, nil) })
}

func TestMapEvercoreErrorToStatus(t *testing.T) {
    // Constraint violation -> AlreadyExists
    ce := &evercore.StorageEngineError{ErrorType: evercore.ErrorTypeConstraintViolation}
    if got := MapEvercoreErrorToStatus(ce); got != ubstatus.AlreadyExists {
        t.Fatalf("expected AlreadyExists, got %v", got)
    }
    // Not found -> NotFound
    ne := &evercore.StorageEngineError{ErrorType: evercore.ErrorNotFound}
    if got := MapEvercoreErrorToStatus(ne); got != ubstatus.NotFound {
        t.Fatalf("expected NotFound, got %v", got)
    }
    // Other -> UnexpectedError
    oe := &evercore.StorageEngineError{}
    if got := MapEvercoreErrorToStatus(oe); got != ubstatus.UnexpectedError {
        t.Fatalf("expected UnexpectedError, got %v", got)
    }
}
