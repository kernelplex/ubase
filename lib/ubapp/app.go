package ubapp

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"

	evercore "github.com/kernelplex/evercore/base"
	"github.com/kernelplex/evercore/evercoreuri"
	_ "github.com/kernelplex/ubase/internal/evercoregen"
	"github.com/kernelplex/ubase/lib/contracts"
	"github.com/kernelplex/ubase/lib/ensure"
	"github.com/kernelplex/ubase/lib/ub2fa"
	"github.com/kernelplex/ubase/lib/ubadminpanel"
	"github.com/kernelplex/ubase/lib/ubconst"
	"github.com/kernelplex/ubase/lib/ubdata"
	"github.com/kernelplex/ubase/lib/ubenv"
	"github.com/kernelplex/ubase/lib/ubmailer"
	"github.com/kernelplex/ubase/lib/ubmanage"
	"github.com/kernelplex/ubase/lib/ubsecurity"
	"github.com/kernelplex/ubase/lib/ubwww"
	ubase_postgres "github.com/kernelplex/ubase/sql/postgres"
	ubase_sqlite "github.com/kernelplex/ubase/sql/sqlite"
	_ "github.com/lib/pq"
	"github.com/xo/dburl"
	_ "modernc.org/sqlite"
)

type UbaseConfig struct {
	WebPort                   uint   `env:"WEB_PORT" default:"8080"`
	DatabaseConnection        string `env:"DATABASE_CONNECTION" default:"/var/data/main.db"`
	EventStoreConnection      string `env:"EVENT_STORE_CONNECTION" default:"/var/data/main.db"`
	Pepper                    []byte `env:"PEPPER" required:"true"`
	SecretKey                 []byte `env:"SECRET_KEY" required:"true"`
	Environment               string `env:"ENVIRONMENT" default:"production"`
	TokenMaxSoftExpirySeconds int    `env:"TOKEN_SOFT_EXPIRY_SECONDS" default:"3600"`  // 1 hour
	TokenMaxHardExpirySeconds int    `env:"TOKEN_HARD_EXPIRY_SECONDS" default:"86400"` // 24 hours
	PrimaryOrganization       int64  `env:"PRIMARY_ORGANIZATION" required:"true"`
	TOTPIssuer                string `env:"TOTP_ISSUER" required:"true"`

	// Mailer
	MailerType      string `env:"MAILER_TYPE" default:"none"`
	MailerFrom      string `env:"MAILER_FROM"`
	MailerUsername  string `env:"MAILER_USERNAME"`
	MailerPassword  string `env:"MAILER_PASSWORD"`
	MailerHost      string `env:"MAILER_HOST"`
	MailerOutputDir string `env:"MAILER_OUTPUT_DIR"`
}

func UbaseConfigFromEnv() UbaseConfig {
	config := UbaseConfig{}
	err := ubenv.ConfigFromEnv(&config)
	if err != nil {
		panic(err)
	}
	return config
}

type BackgroundService interface {
	Start() error
	Stop() error
}

type UbaseApp struct {
	config                *UbaseConfig
	db                    *sql.DB
	dbtype                ubconst.DatabaseType
	dburl                 *dburl.URL
	dbadapter             ubdata.DataAdapter
	store                 *evercore.EventStore // Event store
	hashService           ubsecurity.HashGenerator
	encryptionService     ubsecurity.EncryptionService
	totpService           ub2fa.TotpService
	managementService     ubmanage.ManagementService
	mailer                ubmailer.Mailer
	backgroundMailer      *ubmailer.BackgroundMailer
	prefectService        ubmanage.PrefectService
	backgroundServices    []BackgroundService
	permissionsMiddleware *ubwww.PermissionMiddleware

	cookieManager         contracts.AuthTokenCookieManager
	webService            ubwww.WebService
	adminPanelInitialized bool
}

func NewUbaseAppEnvConfig() UbaseApp {
	config := UbaseConfigFromEnv()
	app := UbaseApp{}
	// Store loaded config on the app instance so GetConfig() reflects actual values
	app.config = &config

	return app
}

func buildDatabase(app *UbaseApp, config *UbaseConfig) {
	ensure.That(len(config.DatabaseConnection) > 0, "database connection string must be set")
	dburl, err := dburl.Parse(config.DatabaseConnection)
	if err != nil {
		panic(fmt.Errorf("failed to parse database connection URL: %w", err))
	}
	app.dburl = dburl

	app.db, err = sql.Open(dburl.Driver, dburl.DSN)
	if err != nil {
		panic(fmt.Errorf("failed to open database connection: %w", err))
	}

	switch dburl.Driver {
	case "postgres":
		slog.Info("Using Postgres database")
		err := ubase_postgres.MigrateUp(app.db)
		if err != nil {
			panic(fmt.Errorf("failed to migrate database: %w", err))
		}
		app.dbtype = ubconst.DatabaseTypePostgres
	case "sqlite3":
		slog.Info("Using SQLite database")

		// Execute sane PRAGMA settings for SQLite
		_, err := app.db.Exec("PRAGMA journal_mode=WAL;")
		if err != nil {
			panic(fmt.Errorf("failed to set journal mode: %w", err))
		}
		_, err = app.db.Exec("PRAGMA synchronous=normal;")
		if err != nil {
			panic(fmt.Errorf("failed to set synchronous mode: %w", err))
		}

		_, err = app.db.Exec("PRAGMA temp_store=memory;")
		if err != nil {
			panic(fmt.Errorf("failed to set temp store: %w", err))
		}

		_, err = app.db.Exec("PRAGMA mmap_size = 30000000000;")
		if err != nil {
			panic(fmt.Errorf("failed to set mmap size: %w", err))
		}

		err = ubase_sqlite.MigrateUp(app.db)
		if err != nil {
			panic(fmt.Errorf("failed to migrate database: %w", err))
		}

		app.dbtype = ubconst.DatabaseTypeSQLite
		slog.Info("Database type set to", "type", app.dbtype)
	default:
		panic(fmt.Sprintf("unsupported database type: %s", dburl.Driver))
	}
	slog.Info("Database type set to", "type", app.dbtype)
}

func (app *UbaseApp) GetConfig() *UbaseConfig {
	if app.config == nil {
		app.config = &UbaseConfig{}
		ubenv.ConfigFromEnv(app.config)
	}
	return app.config
}

func (app *UbaseApp) GetDB() *sql.DB {
	if app.db == nil {
		config := app.GetConfig()
		buildDatabase(app, config)
		slog.Info("Database type set to", "type", app.dbtype)
	}
	return app.db
}

func (app *UbaseApp) GetDBAdapter() ubdata.DataAdapter {
	if app.dbadapter == nil {
		db := app.GetDB()
		dbadapter := ubdata.NewDatabase(app.dbtype, db)
		app.dbadapter = dbadapter
	}
	return app.dbadapter
}

func (app *UbaseApp) GetEventStore() *evercore.EventStore {
	if app.store == nil {
		config := app.GetConfig()

		ensure.That(len(config.EventStoreConnection) > 0, "event store connection string must be set")
		eventStore, err := evercoreuri.Connect(config.EventStoreConnection)
		if err != nil {
			panic(fmt.Errorf("failed to connect to event store: %w", err))
		}
		app.store = eventStore
	}

	return app.store
}

func (app *UbaseApp) GetManagementService() ubmanage.ManagementService {
	if app.managementService == nil {
		store := app.GetEventStore()
		dbadapter := app.GetDBAdapter()
		hashService := app.GetHashService()
		encryptionService := app.GetEncryptionService()
		totpService := app.GetTOTPService()

		app.managementService = ubmanage.NewManagement(store, dbadapter, hashService, encryptionService, totpService)
	}

	return app.managementService
}

func (app *UbaseApp) GetHashService() ubsecurity.HashGenerator {
	if app.hashService == nil {
		config := app.GetConfig()
		hashService := ubsecurity.DefaultArgon2Id
		hashService.Pepper = config.Pepper
		ensure.That(len(hashService.Pepper) > 0, "pepper must be set and greater than zero")
		app.hashService = hashService
	}
	return app.hashService
}

func (app *UbaseApp) GetEncryptionService() ubsecurity.EncryptionService {
	if app.encryptionService == nil {
		config := app.GetConfig()
		ensure.That(len(config.SecretKey) > 0, "secret key must be set and greater than zero")
		app.encryptionService = ubsecurity.NewEncryptionService(config.SecretKey)
	}

	return app.encryptionService
}

func (app *UbaseApp) GetTOTPService() ub2fa.TotpService {
	if app.totpService == nil {
		config := app.GetConfig()
		ensure.That(len(config.TOTPIssuer) > 0, "TOTP issuer must be set and greater than zero")
		app.totpService = ub2fa.NewTotpService(config.TOTPIssuer)
	}
	return app.totpService
}

func (app *UbaseApp) GetMailer() ubmailer.Mailer {
	if app.mailer == nil {
		config := app.GetConfig()
		app.mailer = ubmailer.MaybeNewMailer(ubmailer.MailerConfig{
			Type:      ubmailer.MailerType(config.MailerType),
			From:      config.MailerFrom,
			OutputDir: config.MailerOutputDir,
		})
		ensure.That(app.mailer != nil, "mailer cannot be nil, check MAILER_TYPE configuration")
	}
	return app.mailer
}

func (app *UbaseApp) GetBackgroundMailer() *ubmailer.BackgroundMailer {
	if app.backgroundMailer == nil {
		mailer := app.GetMailer()
		app.backgroundMailer = ubmailer.NewBackgroundMailer(mailer)
		app.RegisterService(app.backgroundMailer)
		app.backgroundMailer.Start()

	}

	if app.mailer != nil {
		slog.Info("Mailer enabled")
	} else {
		slog.Info("Mailer disabled")
	}
	return app.backgroundMailer
}

func (app *UbaseApp) Shutdown() {

	if app.db != nil {
		err := app.db.Close()
		if err != nil {
			slog.Error("Error closing database", "error", err)
		}
	}
	if app.store != nil {
		err := app.store.Close()
		if err != nil {
			slog.Error("Error closing event store", "error", err)
		}
	}

	if app.backgroundMailer != nil {
		app.backgroundMailer.Stop()
	}
}

// Runs the migrations for the ubase database.
func (app *UbaseApp) MigrateUp() error {
	switch app.dburl.Driver {
	case "postgres":
		return ubase_postgres.MigrateUp(app.db)
	case "sqlite3":
		return ubase_sqlite.MigrateUp(app.db)
	default:
		return fmt.Errorf("unsupported database driver: %s", app.dburl.Driver)
	}
}

func (app *UbaseApp) GetPrefectService() ubmanage.PrefectService {

	if app.prefectService == nil {
		managementService := app.GetManagementService()
		eventStore := app.GetEventStore()
		app.prefectService = ubmanage.NewPrefectService(managementService, eventStore, 100, 100)
		app.RegisterService(app.prefectService)
	}

	return app.prefectService
}

func (app *UbaseApp) RegisterService(service BackgroundService) {
	// Check to see if the service is already registered
	for _, s := range app.backgroundServices {
		if fmt.Sprintf("%T", s) == fmt.Sprintf("%T", service) {
			slog.Warn("Service already registered, skipping", "service", fmt.Sprintf("%T", service))
			return
		}
	}
	slog.Info("Registering service", "service", fmt.Sprintf("%T", service))
	app.backgroundServices = append(app.backgroundServices, service)
}

func (app *UbaseApp) StartServices() error {
	for _, service := range app.backgroundServices {
		slog.Info("Starting service...", "service", fmt.Sprintf("%T", service))
		err := service.Start()
		if err != nil {
			return fmt.Errorf("failed to start service: %w", err)
		}
	}
	return nil
}

func (app *UbaseApp) GetCookieManager() contracts.AuthTokenCookieManager {
	if app.cookieManager == nil {
		config := app.GetConfig()
		encryptionService := app.GetEncryptionService()
		secure := config.Environment == "production"

		cookieManager := ubwww.NewCookieMonster(
			encryptionService,
			"auth_token",
			secure,
			int64(config.TokenMaxSoftExpirySeconds),
			contracts.CookieContextKey("auth_token"),
			contracts.IdentityContextKey("user_identity"),
		)
		app.cookieManager = cookieManager
	}
	return app.cookieManager
}

func (app *UbaseApp) GetPermissionsMiddleware() *ubwww.PermissionMiddleware {
	if app.permissionsMiddleware == nil {
		prefectService := app.GetPrefectService()
		cookieManager := app.GetCookieManager()
		permissionMiddleware := ubwww.NewPermissionMiddleware(
			prefectService,
			cookieManager)
		app.permissionsMiddleware = permissionMiddleware
	}
	return app.permissionsMiddleware
}

func (app *UbaseApp) WithAdminPanel(permissions []string) {
	if !app.adminPanelInitialized {
		adapter := app.GetDBAdapter()
		managementService := app.GetManagementService()
		cookieManager := app.GetCookieManager()
		primaryOrganization := app.GetConfig().PrimaryOrganization

		ws := app.GetWebService()
		fs := http.FileServer(http.FS(ubadminpanel.Static))
		ws.AddRouteHandler("/admin/static/", http.StripPrefix("/admin", fs))
		ws.AddRoute(ubadminpanel.AdminRoute(adapter, managementService))
		ws.AddRoute(ubadminpanel.OrganizationsRoute(managementService))
		ws.AddRoute(ubadminpanel.OrganizationOverviewRoute(managementService))
		ws.AddRoute(ubadminpanel.OrganizationCreateRoute(managementService))
		ws.AddRoute(ubadminpanel.OrganizationCreatePostRoute(managementService))
		ws.AddRoute(ubadminpanel.OrganizationEditRoute(managementService))
		ws.AddRoute(ubadminpanel.RoleOverviewRoute(adapter, managementService, permissions))
		ws.AddRoute(ubadminpanel.RoleUsersListRoute(adapter))
		ws.AddRoute(ubadminpanel.RoleUsersAddRoute(adapter, managementService))
		ws.AddRoute(ubadminpanel.RoleUsersRemoveRoute(adapter, managementService))
		ws.AddRoute(ubadminpanel.RolePermissionsListRoute(adapter, permissions))
		ws.AddRoute(ubadminpanel.RolePermissionsAddRoute(adapter, managementService))
		ws.AddRoute(ubadminpanel.RolePermissionsRemoveRoute(adapter, managementService))
		ws.AddRoute(ubadminpanel.RoleCreateRoute(managementService))
		ws.AddRoute(ubadminpanel.RoleCreatePostRoute(managementService))
		ws.AddRoute(ubadminpanel.RoleEditRoute(managementService))
		ws.AddRoute(ubadminpanel.RoleEditPostRoute(managementService))
		ws.AddRoute(ubadminpanel.UsersListRoute(adapter))
		ws.AddRoute(ubadminpanel.UserOverviewRoute(managementService))
		ws.AddRoute(ubadminpanel.UserRolesListRoute(managementService))
		ws.AddRoute(ubadminpanel.UserRolesAddRoute(managementService))
		ws.AddRoute(ubadminpanel.UserRolesRemoveRoute(managementService))
		ws.AddRoute(ubadminpanel.UserCreateRoute(managementService))
		ws.AddRoute(ubadminpanel.UserCreatePostRoute(managementService))
		ws.AddRoute(ubadminpanel.UserEditRoute(managementService))
		ws.AddRoute(ubadminpanel.LoginRoute(primaryOrganization, managementService, cookieManager))
		ws.AddRoute(ubadminpanel.VerifyTwoFactorRoute(managementService, cookieManager))
		ws.AddRoute(ubadminpanel.LogoutRoute(cookieManager))

		app.adminPanelInitialized = true
	}
}

func (app *UbaseApp) GetWebService() ubwww.WebService {
	if app.webService == nil {
		config := app.GetConfig()
		ensure.That(config.PrimaryOrganization > 0, "primary organization must be set and greater than zero")
		ensure.That(config.WebPort > 0, "web port must be set and greater than zero")
		cookieManager := app.GetCookieManager()
		permissionMiddleware := app.GetPermissionsMiddleware()

		webService := ubwww.NewWebService(
			config.WebPort,
			config.PrimaryOrganization,
			cookieManager,
			permissionMiddleware)
		app.webService = webService

		app.RegisterService(webService)

	}
	return app.webService
}

func (app *UbaseApp) StopServices() error {
	for _, service := range app.backgroundServices {
		err := service.Stop()
		if err != nil {
			slog.Error("Error stopping service", "error", err)
		}
	}
	return nil
}
