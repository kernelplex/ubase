# UBase - Modern Identity and Access Management Framework

UBase is a Go toolkit for building identity, authentication, and authorization flows backed by an Evercore event store. It combines a production-ready management service, a CLI and admin panel, and a library of security primitives (Argon2id hashing, AES-256 encryption, and TOTP) that can be embedded into other Go programs or run as a stand‑alone service.

## Feature Highlights
- **Event-sourced core** – Every state change is written through Evercore so you can replay history, emit projections, or subscribe to domain events.
- **Dual persistence** – Supports PostgreSQL and SQLite with auto-migrations and driver-specific safety defaults (WAL, PRAGMAs).
- **Batteries included security** – HashService (Argon2id + pepper), EncryptionService (AES-256-GCM), and TOTPService ship out of the box.
- **Management API + CLI** – `ubmanage.ManagementService` powers both application code and the CLI for organizations, users, roles, and secrets.
- **Admin panel** – `./build/ubase serve` exposes the `ubadminpanel` UI protected by permission middleware.
- **Pluggable mailers** – Configure SMTP, write-to-disk, or noop providers without code changes.

## Project Layout
| Path | Description |
| --- | --- |
| `cmd/ubase` | CLI entry point; loads `.env` and runs subcommands. |
| `internal/commands` | All CLI commands: migrations, secrets, serve, organization/role/user management. |
| `lib/ubapp` | Application wiring (config, DB/event store init, background services, admin panel setup). |
| `lib/ubmanage` | Domain services consumed by CLI, web, and external programs. |
| `lib/ubwww`, `lib/ubadminpanel` | Web server + admin UI components. |
| `lib/ubsecurity`, `lib/ub2fa`, `lib/ubmailer` | Security primitives (hashing, encryption, TOTP, token/cookie helpers, mailers). |
| `sql/` | Migration scripts for PostgreSQL and SQLite (invoked via CLI). |
| `integration_tests/` | Cross-database integration suite (covers management service behavior). |

## Requirements
- Go 1.20+
- PostgreSQL 13+ **or** SQLite 3 (when using SQLite include `?_time_format=sqlite` in DSNs)
- An accessible Evercore event store backend (`DATABASE_CONNECTION` and `EVENT_STORE_CONNECTION` can both point to SQLite or Postgres)
- `make`, `sqlc`, `templ`, and the Evercore generator installed:
  ```bash
  go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
  go install github.com/a-h/templ/cmd/templ@latest
  go install github.com/kernelplex/evercore/cmd/evercoregen@latest
  ```

## Quick Start
1. **Generate code**
   ```bash
   make evercoregen sqlc   # required before builds/tests
   ```
2. **Build the CLI**
   ```bash
   make main               # outputs ./build/ubase
   ```
3. **Create a `.env` file** (all required values are read by `cmd/ubase`):
   ```bash
   PEPPER=$(./build/ubase secret --length 32)
   SECRET_KEY=$(./build/ubase secret --length 32)
   PRIMARY_ORGANIZATION=1
   TOTP_ISSUER=UBaseDemo
   DATABASE_CONNECTION=sqlite:///var/data/main.db?_time_format=sqlite
   EVENT_STORE_CONNECTION=sqlite:///var/data/events.db?_time_format=sqlite
   WEB_PORT=8080
   ENVIRONMENT=development
   MAILER_TYPE=none
   ```
   `PEPPER` and `SECRET_KEY` must be base64-encoded strings; the `secret` command returns ready-to-use values. When pointing to SQLite files remember to include the `?_time_format=sqlite` parameter so timestamps round-trip correctly.
4. **Run database migrations**
   ```bash
   ./build/ubase migrate-up
   ```
5. **Start the admin server**
   ```bash
   ./build/ubase serve
   ```
   The server listens on `WEB_PORT` (defaults to `8080`). Permissions are enforced through `ubadminpanel` so only roles with `ubadminpanel.PermSystemAdmin` can use the UI.

## Command-Line Interface
UBase ships with a command for every major workflow exposed by `ubmanage`. Run `./build/ubase help` to see the complete list.

### Utility commands
- `migrate-up` – applies SQL migrations for the configured database.
- `secret` – prints base64 secrets; perfect for `PEPPER`, `SECRET_KEY`, or API keys.
- `totp-generate` – generates TOTP seeds and example codes for audits or manual MFA setup.

### Organization & role management
- `organization-add`, `organization-update`, `organization-list`, `organization-settings-set/clear`
- `role-add`, `role-list`, `role-view`, `role-add-permissions`

### User management
- `user-add`, `user-update`, `user-verify`
- Role binding: `user-add-role`, `user-remove-role`
- API keys: `user-add-api-key`, `user-delete-api-key`, `user-list-api-keys`
- Lifecycle helpers: `user-enable`, `user-disable`
- Preferences & security: `user-settings-set/clear`, `user-set-twofactor`

### Example bootstrap workflow
```bash
./build/ubase organization-add --system-name acme --name "Acme Corp"
./build/ubase user-add --email admin@acme.com --password "SecurePassword123!" --display-name "Admin"
./build/ubase role-add --system-name admin --name "Organization Admin" --organization-id 1
./build/ubase role-add-permissions --role-id 1 --permissions system_admin
./build/ubase user-add-role --user-id 1 --role-id 1
```

## Embedding in Go
Use `ubapp.UbaseApp` to embed the management service directly inside another Go service:

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/kernelplex/ubase/lib/ubapp"
	"github.com/kernelplex/ubase/lib/ubmanage"
)

func main() {
	app := ubapp.NewUbaseAppEnvConfig()
	defer app.Shutdown()

	mgmt := app.GetManagementService()

	orgResp, err := mgmt.OrganizationAdd(context.Background(), ubmanage.OrganizationCreateCommand{
		Name:       "Acme Corp",
		SystemName: "acme",
		Status:     "active",
	}, "setup-script")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created org ID: %d\n", orgResp.Data.Id)

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

## Configuration Reference
| Variable | Required | Default | Notes |
| --- | --- | --- | --- |
| `WEB_PORT` | No | `8080` | Port for the admin/web server. |
| `DATABASE_CONNECTION` | Yes (practically) | `/var/data/main.db` | PostgreSQL DSN or SQLite path; migrations target this connection. |
| `EVENT_STORE_CONNECTION` | Yes | `/var/data/main.db` | Evercore connection string; add `?_time_format=sqlite` when using SQLite. |
| `PEPPER` | Yes | – | Base64-encoded 32 bytes for Argon2id hashing. |
| `SECRET_KEY` | Yes | – | Base64-encoded 32 bytes for AES-256-GCM encryption. |
| `ENVIRONMENT` | No | `production` | Used for logging/feature toggles. |
| `TOKEN_SOFT_EXPIRY_SECONDS` | No | `3600` | Max soft expiry for issued tokens. |
| `TOKEN_HARD_EXPIRY_SECONDS` | No | `86400` | Hard cap on token lifetime. |
| `PRIMARY_ORGANIZATION` | Yes | – | ID used by the admin panel and defaults. |
| `TOTP_ISSUER` | Yes | – | Issuer label used when generating TOTP secrets. |
| `MAILER_TYPE` | No | `none` | One of `none`, `noop`, `file`, `smtp`. |
| `MAILER_FROM` | Conditionally | – | Required for `file` and `smtp`. |
| `MAILER_USERNAME` | Conditionally | – | Required for `smtp`. |
| `MAILER_PASSWORD` | Conditionally | – | Required for `smtp`. |
| `MAILER_HOST` | Conditionally | – | Required for `smtp` (host:port). |
| `MAILER_OUTPUT_DIR` | Conditionally | – | Required for `file`; emails are saved to disk. |

Mail delivery defaults to `MAILER_TYPE=none`; when the mailer is disabled no other `MAILER_*` variables are needed.

## Development Workflow
- Generate code: `make evercoregen sqlc`
- Build binary: `make main` (outputs `build/ubase`)
- Run migrations locally: `./build/ubase migrate-up`
- Run the admin server for UI-driven workflows: `./build/ubase serve`
- Run the linter: `make lint`

## Testing
- Run the full suite (unit + integration): `make test`
- Only Postgres-backed integration tests: `make test-postgresql`
- Only SQLite-backed integration tests: `make test-sqlite`
- To execute a single integration test: `go test -run TestName -tags sqlite,postgresql integration_tests/*.go`

## Advanced Features
### Two-Factor Authentication
```go
secretResp, err := mgmt.UserGenerateTwoFactorSharedSecret(ctx, ubmanage.UserGenerateTwoFactorSharedSecretCommand{
	Id: userID,
}, "admin")

verifyResp, err := mgmt.UserVerifyTwoFactorCode(ctx, ubmanage.UserVerifyTwoFactorLoginCommand{
	UserId: userID,
	Code:   "123456",
}, "admin")
```

### Event Sourcing
All state transitions are persisted through Evercore. You can rebuild read models, subscribe to specific event types, or plug in custom background services by registering them on `ubapp.UbaseApp`.

## License
[MIT](https://choosealicense.com/licenses/mit/)
