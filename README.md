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
```bash
# Required
export PEPPER="your-pepper-value" 
export SECRET_KEY="32-byte-encryption-key"
export TOTP_ISSUER="YourAppName"

# Database (defaults to SQLite)
export DATABASE_CONNECTION="postgres://user:pass@localhost/dbname"
export EVENT_STORE_CONNECTION="sqlite:///var/data/events.db"

# Optional
export ENVIRONMENT="development"
export TOKEN_SOFT_EXPIRY_SECONDS="3600"  # 1 hour
export TOKEN_HARD_EXPIRY_SECONDS="86400" # 24 hours
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
