package ubaseapp

import (
	"database/sql"

	evercore "github.com/kernelplex/evercore/base"
	ubase "github.com/kernelplex/ubase/lib"
	"github.com/kernelplex/ubase/lib/ubsecurity"
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
	/*
		config            *config.Config
	*/
	db                *sql.DB
	edb               *sql.DB // Event store database connection
	store             *evercore.EventStore
	userService       ubase.UserService
	hashService       ubsecurity.HashGenerator
	encryptionService ubsecurity.EncryptionService
	permissionService ubase.PermissionService
	roleService       ubase.RoleService
}
