# UBase - Modern Identity and Access Management Framework

UBase is a comprehensive Go framework for identity management, authentication,
and authorization with event sourcing capabilities.

## Key Components

### Core Services
- **ManagementService**: Unified interface for all identity operations
- **EncryptionService**: AES-256 encryption with configurable keys
- **HashService**: Argon2id password hashing with pepper
- **TOTPService**: Time-based one-time password generation/validation

### Event Sourcing
- Built on Evercore event store
- All state changes captured as events
- Event replay capabilities
- Strong consistency guarantees

### Security Features
- End-to-end encryption
- Secure secret generation
- Two-factor authentication
- Automatic database migrations

## Getting Started

### Prerequisites
- Go 1.20+
- PostgreSQL or SQLite
- Event store (compatible with Evercore)

### Installation
```bash
go get github.com/kernelplex/ubase
```

### Configuration (via Environment Variables)

>[!IMPORTANT]
> When using sqlite, use the **?_time_format=sqlite** query parameter to ensure proper time handling.


```bash
# Required (base64-encoded values)
export PEPPER="<base64-encoded-32-bytes>"
export SECRET_KEY="<base64-encoded-32-bytes>"  # AES-256 key
export TOTP_ISSUER="YourAppName"

# Database (defaults to SQLite)
export DATABASE_CONNECTION="postgres://user:pass@localhost/dbname"
export EVENT_STORE_CONNECTION="sqlite:///var/data/events.db?_time_format=sqlite"

# Optional
export ENVIRONMENT="development"
export TOKEN_SOFT_EXPIRY_SECONDS="3600"  # 1 hour
export TOKEN_HARD_EXPIRY_SECONDS="86400" # 24 hours

# Mailer (optional)
# MAILER_TYPE controls delivery; defaults to "none".
# Supported values: none, noop, file, smtp
# For file: set MAILER_FROM and MAILER_OUTPUT_DIR
# For smtp: set MAILER_FROM, MAILER_USERNAME, MAILER_PASSWORD, MAILER_HOST
# When MAILER_TYPE is none or noop, no MAILER_* values are required
# export MAILER_TYPE="none"
# export MAILER_FROM="no-reply@example.com"
# export MAILER_OUTPUT_DIR="/var/emails"     # when MAILER_TYPE=file
# export MAILER_USERNAME="smtp-user"         # when MAILER_TYPE=smtp
# export MAILER_PASSWORD="smtp-pass"         # when MAILER_TYPE=smtp
# export MAILER_HOST="smtp.example.com:587"  # when MAILER_TYPE=smtp
```

Mail delivery is disabled by default (`MAILER_TYPE=none`).
`MAILER_FROM` and related `MAILER_*` settings are only required when a mailer is enabled:
- `file`: requires `MAILER_FROM` and `MAILER_OUTPUT_DIR`
- `smtp`: requires `MAILER_FROM`, `MAILER_USERNAME`, `MAILER_PASSWORD`, `MAILER_HOST`
- `none` or `noop`: no mailer settings required

Tip: `PEPPER` and `SECRET_KEY` are expected to be base64-encoded bytes. Use the CLI helper to generate secure values:

```bash
./build/ubase secret --length 32  # generate base64, use for PEPPER/SECRET_KEY
```

### Basic Usage
```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/kernelplex/ubase/lib/ubapp"
)

func main() {
	// Initialize app with environment config
	app := ubapp.NewUbaseAppEnvConfig()
	defer app.Shutdown()

	// Get management service
	mgmt := app.GetManagementService()

	// Create organization
	orgResp, err := mgmt.OrganizationAdd(context.Background(), ubmanage.OrganizationCreateCommand{
		Name:       "Acme Corp",
		SystemName: "acme",
		Status:     "active",
	}, "setup-script")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created org ID: %d\n", orgResp.Data.Id)

	// Create admin user
	userResp, err := mgmt.UserAdd(context.Background(), ubmanage.UserCreateCommand{
		Email:       "admin@acme.com",
		Password:    "SecurePassword123!",
		FirstName:   "Admin",
		LastName:    "User",
		DisplayName: "Admin",
		Verified:    true,
	}, "setup-script")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created user ID: %d\n", userResp.Data.Id)
}
```

## Advanced Features

### Two-Factor Authentication
```go
// Generate 2FA secret
secretResp, err := mgmt.UserGenerateTwoFactorSharedSecret(ctx, ubmanage.UserGenerateTwoFactorSharedSecretCommand{
	Id: userID,
}, "admin")

// Verify code
verifyResp, err := mgmt.UserVerifyTwoFactorCode(ctx, ubmanage.UserVerifyTwoFactorLoginCommand{
	UserId: userID,
	Code:   "123456", 
}, "admin")
```

### Event Sourcing
All state changes are automatically recorded as events in the event store. You can:
- Rebuild state from events
- Subscribe to specific event types
- Implement event handlers for side effects

## License
[MIT](https://choosealicense.com/licenses/mit/)
