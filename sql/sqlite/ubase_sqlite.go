package ubase_sqlite

import (
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

var migrationsDir = "migrations"

//go:embed migrations/*.sql
var EmbeddedSqliteMigrations embed.FS

func MigrateUp(db *sql.DB) error {
	goose.SetDialect("sqlite3")
	goose.SetBaseFS(EmbeddedSqliteMigrations)
	goose.SetTableName("ubase_migrations")
	return goose.Up(db, migrationsDir)
}
