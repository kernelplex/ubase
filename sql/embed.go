package sqlpack

import (
	"embed"
)

//go:embed sqlite/migrations/*.sql
var EmbeddedSqliteMigrations embed.FS

//go:embed postgres/migrations/*.sql
var EmbeddedPostgresMigrations embed.FS
