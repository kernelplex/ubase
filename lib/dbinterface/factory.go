package dbinterface

import (
	"github.com/kernelplex/ubase/lib/dbsqlite"
)

type DatabaseType string

const (
	DatabaseTypeSQLite   DatabaseType = "sqlite"
	DatabaseTypePostgres DatabaseType = "postgres"
)

func NewDatabase(dbType DatabaseType, db dbsqlite.DBTX) Database {
	switch dbType {
	case DatabaseTypePostgres:
		return NewPostgresAdapter(db)
	case DatabaseTypeSQLite:
		return NewSQLiteAdapter(db)
	default:
		panic("unknown database type")
	}
}
