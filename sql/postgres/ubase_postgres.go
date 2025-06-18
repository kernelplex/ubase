package ubase_postgres

import (
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

var migrationsDir = "migrations"

//go:embed migrations/*.sql
var EmbeddedPostgresMigrations embed.FS

func MigrateUp(db *sql.DB) error {
	goose.SetDialect("postgres")
	goose.SetBaseFS(EmbeddedPostgresMigrations)
	goose.SetTableName("ubase_migrations")
	return goose.Up(db, migrationsDir)
}
