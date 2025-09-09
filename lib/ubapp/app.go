package ubapp

import (
	"database/sql"
	"fmt"
	"log/slog"

	evercore "github.com/kernelplex/evercore/base"
	"github.com/kernelplex/evercore/evercoreuri"
	_ "github.com/kernelplex/ubase/internal/evercoregen"
	"github.com/kernelplex/ubase/lib/ub2fa"
	"github.com/kernelplex/ubase/lib/ubconst"
	"github.com/kernelplex/ubase/lib/ubdata"
	"github.com/kernelplex/ubase/lib/ubenv"
	"github.com/kernelplex/ubase/lib/ubmailer"
	"github.com/kernelplex/ubase/lib/ubmanage"
	"github.com/kernelplex/ubase/lib/ubsecurity"
	ubase_postgres "github.com/kernelplex/ubase/sql/postgres"
	ubase_sqlite "github.com/kernelplex/ubase/sql/sqlite"
	_ "github.com/lib/pq"
	"github.com/xo/dburl"
	_ "modernc.org/sqlite"
)

type UbaseConfig struct {
    DatabaseConnection        string `env:"DATABASE_CONNECTION" default:"/var/data/main.db"`
    EventStoreConnection      string `env:"EVENT_STORE_CONNECTION" default:"/var/data/main.db"`
    Pepper                    []byte `env:"PEPPER" required:"true"`
    SecretKey                 []byte `env:"SECRET_KEY" required:"true"`
    Environment               string `env:"ENVIRONMENT" default:"production"`
    TokenMaxSoftExpirySeconds int    `env:"TOKEN_SOFT_EXPIRY_SECONDS" default:"3600"`  // 1 hour
    TokenMaxHardExpirySeconds int    `env:"TOKEN_HARD_EXPIRY_SECONDS" default:"86400"` // 24 hours
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

type UbaseApp struct {
	config            UbaseConfig
	db                *sql.DB
	dburl             *dburl.URL
	dbadapter         ubdata.DataAdapter
	store             *evercore.EventStore // Event store
	hashService       ubsecurity.HashGenerator
	encryptionService ubsecurity.EncryptionService
	totpService       ub2fa.TotpService
	managementService ubmanage.ManagementService
	mailer            ubmailer.Mailer
	backgroundMailer  *ubmailer.BackgroundMailer
	prefectService    ubmanage.PrefectService
}

func NewUbaseAppEnvConfig() UbaseApp {
    config := UbaseConfigFromEnv()
    app := UbaseApp{}
    // Store loaded config on the app instance so GetConfig() reflects actual values
    app.config = config

	// ======================================================================
	// Hashing service
	// ======================================================================
	hashService := ubsecurity.DefaultArgon2Id
	hashService.Pepper = config.Pepper
	app.hashService = hashService

	// ======================================================================
	// Database connections
	// ====================================================st==================
	eventStore, err := evercoreuri.Connect(config.EventStoreConnection)
	if err != nil {
		panic(fmt.Errorf("failed to connect to event store: %w", err))
	}

	app.store = eventStore

	// ======================================================================
	// UBase database
	// ======================================================================
	dburl, err := dburl.Parse(config.DatabaseConnection)
	if err != nil {
		panic(fmt.Errorf("failed to parse database connection URL: %w", err))
	}
	app.dburl = dburl

	app.db, err = sql.Open(dburl.Driver, dburl.DSN)
	if err != nil {
		panic(fmt.Errorf("failed to open database connection: %w", err))
	}

	var databaseType ubconst.DatabaseType

	switch dburl.Driver {
	case "postgres":
		err := ubase_postgres.MigrateUp(app.db)
		if err != nil {
			panic(fmt.Errorf("failed to migrate database: %w", err))
		}
		databaseType = ubconst.DatabaseTypePostgres
	case "sqlite3":

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

		databaseType = ubconst.DatabaseTypeSQLite
	default:
		panic(fmt.Sprintf("unsupported database type: %s", dburl.Driver))
	}

	dbadapter := ubdata.NewDatabase(databaseType, app.db)
	app.dbadapter = dbadapter

	// ======================================================================
	// UBase TOTP
	// ======================================================================

	app.totpService = ub2fa.NewTotpService(config.TOTPIssuer)

	// ======================================================================
	// UBase mailer
	// ======================================================================
	app.mailer = ubmailer.MaybeNewMailer(ubmailer.MailerConfig{
		Type:      ubmailer.MailerType(config.MailerType),
		From:      config.MailerFrom,
		OutputDir: config.MailerOutputDir,
	})
	if app.mailer != nil {
		slog.Info("Mailer enabled")
		app.backgroundMailer = ubmailer.NewBackgroundMailer(app.mailer)
		app.backgroundMailer.Start()
	} else {
		slog.Info("Mailer disabled")
	}

	// ======================================================================
	// UBase services
	// ======================================================================

	app.encryptionService = ubsecurity.NewEncryptionService(config.SecretKey)

	app.managementService = ubmanage.NewManagement(app.store, dbadapter, app.hashService, app.encryptionService, app.totpService)

	app.prefectService = ubmanage.NewPrefectService(app.managementService, 100, 100)

	app.dbadapter = dbadapter

	return app
}

func (app *UbaseApp) GetConfig() *UbaseConfig {
	return &app.config
}

func (app *UbaseApp) GetDB() *sql.DB {
	return app.db
}

func (app *UbaseApp) GetDBAdapter() ubdata.DataAdapter {
	return app.dbadapter
}

func (app *UbaseApp) GetEventStore() *evercore.EventStore {
	return app.store
}

func (app *UbaseApp) GetManagementService() ubmanage.ManagementService {
	return app.managementService
}

func (app *UbaseApp) GetHashService() ubsecurity.HashGenerator {
	return app.hashService
}

func (app *UbaseApp) GetEncryptionService() ubsecurity.EncryptionService {
	return app.encryptionService
}

func (app *UbaseApp) GetTOTPService() ub2fa.TotpService {
	return app.totpService
}

func (app *UbaseApp) GetMailer() ubmailer.Mailer {
	return app.mailer
}

func (app *UbaseApp) GetBackgroundMailer() *ubmailer.BackgroundMailer {
	return app.backgroundMailer
}

func (app *UbaseApp) Shutdown() {
	err := app.db.Close()
	if err != nil {
		slog.Error("Error closing database", "error", err)
	}
	err = app.store.Close()
	if err != nil {
		slog.Error("Error closing event store", "error", err)
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
	return app.prefectService
}
