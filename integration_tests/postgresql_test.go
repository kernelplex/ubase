//go:build postgresql

package integration_tests

import (
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	pg "github.com/kernelplex/evercore/evercorepostgres"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/kernelplex/ubase/lib/ubdata"
	ubase_postgres "github.com/kernelplex/ubase/sql/postgres"
)

func TestPostgresqlStorageEngine(t *testing.T) {
	ctx := t.Context()

	postgresContainer, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.BasicWaitStrategies(),
	)

	if err != nil {
		t.Fatalf("failed to start postgres container: %s", err)
		return
	}
	defer func() {
		err := postgresContainer.Terminate(ctx)
		if err != nil {
			t.Fatalf("failed to terminate postgres container: %s", err)
		}
	}()

	err = postgresContainer.Start(ctx)
	if err != nil {
		t.Fatalf("failed to start postgres container: %s", err)
		return
	}

	connectionString, err := postgresContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get postgres connection string: %s", err)
		return
	}

	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		t.Fatalf("failed to open postgres connection: %s", err)
		return
	}

	err = pg.MigrateUp(db)
	if err != nil {
		t.Fatalf("failed to migrate postgres database: %s", err)
		return
	}
	err = ubase_postgres.MigrateUp(db)
	if err != nil {
		t.Fatalf("failed to migrate postgres database: %s", err)
		return
	}

	adapter := ubdata.NewPostgresAdapter(db)
	testSuite := NewAdapterExercises(db, adapter)

	// Run the tests
	testSuite.RunTests(t)

	/*
		storage := pg.NewPostgresStorageEngine(db)
		eventStore := evercore.NewEventStore(storage)

		// Create a new test suite
		testSuite := NewStorageEngineTestSuite(eventStore, db, ubconst.DatabaseTypePostgres)

		// Run the tests
		testSuite.RunTests(t)
	*/

}
