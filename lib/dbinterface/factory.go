package dbinterface

import (
	"github.com/kernelplex/ubase/lib/dbsqlite"
	"github.com/kernelplex/ubase/lib/ubconst"
)

func NewDatabase(dbType ubconst.DatabaseType, db dbsqlite.DBTX) Database {
	switch dbType {
	case ubconst.DatabaseTypePostgres:
		return NewPostgresAdapter(db)
	case ubconst.DatabaseTypeSQLite:
		return NewSQLiteAdapter(db)
	default:
		panic("unknown database type")
	}
}
