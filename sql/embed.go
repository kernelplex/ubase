package sqlpack

import (
	"embed"
)

//go:embed sqlite/migrations/*.sql
var EmbeddedSqliteMigrations embed.FS
