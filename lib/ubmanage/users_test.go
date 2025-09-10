package ubmanage

import (
    "testing"
    "time"

    evercore "github.com/kernelplex/evercore/base"
)

func TestUserAggregateApplyEventState_LoginAndTokens(t *testing.T) {
    agg := &UserAggregate{}

    // Seed with added event using generic state event mechanism
    add := evercore.NewStateEvent(UserAddedEvent{
        Email:        "user@example.com",
        PasswordHash: "hash",
        FirstName:    "A",
        LastName:     "B",
        DisplayName:  "A B",
        Verified:     false,
    })
    now := time.Now()
    if err := agg.ApplyEventState(add, now, "tester"); err != nil {
        t.Fatalf("apply add: %v", err)
    }

    // Login failed increments failed count and sets attempt time
    before := agg.State.FailedLoginAttempts
    if err := agg.ApplyEventState(UserLoginFailedEvent{Reason: "bad"}, now.Add(time.Second), "tester"); err != nil {
        t.Fatalf("apply login failed: %v", err)
    }
    if agg.State.FailedLoginAttempts != before+1 {
        t.Fatalf("expected failed attempts %d, got %d", before+1, agg.State.FailedLoginAttempts)
    }

    // Login succeeded resets failed count and sets last login
    if err := agg.ApplyEventState(UserLoginSucceededEvent{}, now.Add(2*time.Second), "tester"); err != nil {
        t.Fatalf("apply login succeeded: %v", err)
    }
    if agg.State.FailedLoginAttempts != 0 {
        t.Fatal("expected failed attempts reset to 0")
    }
    if agg.State.LastLogin == 0 || agg.State.LastLogin < now.Unix() {
        t.Fatal("expected last login set")
    }

    // Verification token generate/verify
    if err := agg.ApplyEventState(UserVerificationTokenGeneratedEvent{Token: "enc-token"}, now.Add(3*time.Second), "tester"); err != nil {
        t.Fatalf("apply token generated: %v", err)
    }
    if agg.State.VerificationToken == nil || *agg.State.VerificationToken != "enc-token" || agg.State.Verified {
        t.Fatal("expected token set and verified=false")
    }
    if err := agg.ApplyEventState(UserVerificationTokenVerifiedEvent{}, now.Add(4*time.Second), "tester"); err != nil {
        t.Fatalf("apply token verified: %v", err)
    }
    if !agg.State.Verified || agg.State.VerificationToken != nil {
        t.Fatal("expected verified=true and token cleared")
    }
}

func TestUserAggregateApplyEventState_TwoFactorAndApiKeys(t *testing.T) {
    agg := &UserAggregate{}
    add := evercore.NewStateEvent(UserAddedEvent{
        Email:        "user@example.com",
        PasswordHash: "hash",
        FirstName:    "A",
        LastName:     "B",
        DisplayName:  "A B",
        Verified:     true,
    })
    now := time.Now()
    if err := agg.ApplyEventState(add, now, "tester"); err != nil {
        t.Fatalf("apply add: %v", err)
    }

    // 2FA enable/disable
    if err := agg.ApplyEventState(UserTwoFactorEnabledEvent{SharedSecret: "enc-secret"}, now.Add(time.Second), "tester"); err != nil {
        t.Fatalf("apply 2fa enabled: %v", err)
    }
    if agg.State.TwoFactorSharedSecret == nil || *agg.State.TwoFactorSharedSecret != "enc-secret" {
        t.Fatal("expected 2fa secret set")
    }
    if err := agg.ApplyEventState(UserTwoFactorDisabledEvent{}, now.Add(2*time.Second), "tester"); err != nil {
        t.Fatalf("apply 2fa disabled: %v", err)
    }
    if agg.State.TwoFactorSharedSecret != nil {
        t.Fatal("expected 2fa secret cleared")
    }

    // API key add/delete
    addKey := UserApiKeyAddedEvent{Id: "id123", OrganizationId: 1, SecretHash: "h", Name: "n", CreatedAt: now.Unix(), ExpiresAt: now.Add(24*time.Hour).Unix()}
    if err := agg.ApplyEventState(addKey, now.Add(3*time.Second), "tester"); err != nil {
        t.Fatalf("apply api key added: %v", err)
    }
    if len(agg.State.ApiKeys) != 1 || agg.State.ApiKeys[0].Id != "id123" {
        t.Fatalf("expected 1 api key with id id123, got %+v", agg.State.ApiKeys)
    }
    if err := agg.ApplyEventState(UserApiKeyDeletedEvent{Id: "id123"}, now.Add(4*time.Second), "tester"); err != nil {
        t.Fatalf("apply api key deleted: %v", err)
    }
    if len(agg.State.ApiKeys) != 0 {
        t.Fatalf("expected no api keys, got %+v", agg.State.ApiKeys)
    }
}

func TestUserCommandValidation(t *testing.T) {
    // UserCreateCommand valid/invalid
    valid := UserCreateCommand{Email: "a@b", Password: "Abcdef1!", FirstName: "A", LastName: "B", DisplayName: "AB", Verified: true}
    if ok, _ := valid.Validate(); !ok { t.Fatal("expected valid create") }
    invalid := UserCreateCommand{}
    if ok, _ := invalid.Validate(); ok { t.Fatal("expected invalid create") }

    // UserUpdateCommand
    email := "c@d"
    pw := "Xyzzzz1!"
    first := "C"
    upd := UserUpdateCommand{Id: 1, Email: &email, Password: &pw, FirstName: &first}
    if ok, _ := upd.Validate(); !ok { t.Fatal("expected valid update") }
    upd = UserUpdateCommand{Id: 0, Email: &email}
    if ok, _ := upd.Validate(); ok { t.Fatal("expected invalid update id") }

    // UserGenerateVerificationTokenCommand
    if ok, _ := (UserGenerateVerificationTokenCommand{Id: 1}).Validate(); !ok { t.Fatal("expected valid token cmd") }
    if ok, _ := (UserGenerateVerificationTokenCommand{Id: 0}).Validate(); ok { t.Fatal("expected invalid token cmd") }

    // UserGenerateApiKeyCommand
    gen := UserGenerateApiKeyCommand{UserId: 1, Name: "key", OrganizationId: 1, ExpiresAt: time.Now().Add(24*time.Hour)}
    if ok, _ := gen.Validate(); !ok { t.Fatal("expected valid api key cmd") }
    gen = UserGenerateApiKeyCommand{UserId: 0, Name: "key", OrganizationId: 1, ExpiresAt: time.Now().Add(24*time.Hour)}
    if ok, _ := gen.Validate(); ok { t.Fatal("expected invalid api key cmd userId") }
}

