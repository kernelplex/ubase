package ubdata

import (
	"database/sql"
	"fmt"

	"github.com/kernelplex/ubase/lib/ubconst"
)

func NewDatabase(dbType ubconst.DatabaseType, db *sql.DB) DataAdapter {
	switch dbType {
	case ubconst.DatabaseTypePostgres:
		return NewPostgresAdapter(db)
	case ubconst.DatabaseTypeSQLite:
		return NewSQLiteAdapter(db)
	default:
		panic(fmt.Sprintf("unsupported database type: '%s'", dbType))
	}
}
