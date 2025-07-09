package ubapp

import (
	"database/sql"
	"fmt"
	"log/slog"

	evercore "github.com/kernelplex/evercore/base"
	"github.com/kernelplex/evercore/evercoreuri"
	ubase "github.com/kernelplex/ubase/lib"
	"github.com/kernelplex/ubase/lib/ubconst"
	"github.com/kernelplex/ubase/lib/ubdata"
	"github.com/kernelplex/ubase/lib/ubenv"
	"github.com/kernelplex/ubase/lib/ubsecurity"
	"github.com/kernelplex/ubase/sql/postgres"
	"github.com/kernelplex/ubase/sql/sqlite"
	"github.com/xo/dburl"
)

type Config struct {
	DatabaseConnection        string `env:"DATABASE_CONNECTION" default:"/var/data/main.db"`
	EventStoreConnection      string `env:"EVENT_STORE_CONNECTION" default:"/var/data/main.db"`
	Pepper                    []byte `env:"PEPPER" required:"true"`
	SecretKey                 []byte `env:"SECRET_KEY" required:"true"`
	Environment               string `env:"ENVIRONMENT" default:"production"`
	TokenMaxSoftExpirySeconds int    `env:"TOKEN_SOFT_EXPIRY_SECONDS" default:"3600"`  // 1 hour
	TokenMaxHardExpirySeconds int    `env:"TOKEN_HARD_EXPIRY_SECONDS" default:"86400"` // 24 hours
}

type App struct {
	config            Config
	db                *sql.DB
	dbadapter         ubdata.DataAdapter
	store             *evercore.EventStore // Event store
	hashService       ubsecurity.HashGenerator
	encryptionService ubsecurity.EncryptionService
	permissionService ubase.PermissionService
	roleService       ubase.RoleService
	userService       ubase.UserService
}

func (app *App) SetupFromEnv() error {
	config := Config{}
	err := ubenv.ConfigFromEnv(&config)
	if err != nil {
		return err
	}
	return app.Setup(config)
}

func (app *App) GetConfig() *Config {
	return &app.config
}

func (app *App) GetDB() *sql.DB {
	return app.db
}

func (app *App) GetEventStore() *evercore.EventStore {
	return app.store
}

func (app *App) Setup(config Config) error {
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
		return err
	}

	app.store = eventStore

	// ======================================================================
	// UBase services
	// ======================================================================
	dburl, err := dburl.Parse(config.DatabaseConnection)
	if err != nil {
		return err
	}
	app.db, err = sql.Open(dburl.Driver, dburl.DSN)
	if err != nil {
		return err
	}

	var databaseType ubconst.DatabaseType

	if dburl.Driver == "postgres" {
		ubase_postgres.MigrateUp(app.db)
		databaseType = ubconst.DatabaseTypePostgres
	} else if dburl.Driver == "sqlite3" {
		ubase_sqlite.MigrateUp(app.db)
		databaseType = ubconst.DatabaseTypeSQLite
	} else {
		return fmt.Errorf("unsupported database type: %s", dburl.Driver)
	}

	dbadapter := ubdata.NewDatabase(databaseType, app.db)

	app.dbadapter = dbadapter

	app.roleService = ubase.CreateRoleService(app.store, dbadapter)
	app.userService = ubase.CreateUserService(app.store, app.hashService, dbadapter)
	app.permissionService = ubase.NewPermissionService(dbadapter, app.roleService)
	return nil
}

func NewAppFromEnv() (*App, error) {
	config := Config{}
	err := ubenv.ConfigFromEnv(&config)
	if err != nil {
		return nil, err
	}
	return NewApp(config)
}

func NewApp(config Config) (*App, error) {
	app := App{
		config: config,
	}

	return &app, nil
}

func (app *App) Shutdown() {
	err := app.db.Close()
	if err != nil {
		slog.Error("Error closing database", "error", err)
	}
	err = app.store.Close()
	if err != nil {
		slog.Error("Error closing event store", "error", err)
	}
}
