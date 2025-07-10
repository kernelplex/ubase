package integration_tests

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	evercore "github.com/kernelplex/evercore/base"
	pg "github.com/kernelplex/evercore/evercorepostgres"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/kernelplex/ubase/lib/ubconst"
	"github.com/kernelplex/ubase/sql/postgres"
)

func TestPostgresqlStorageEngine(t *testing.T) {
	ctx := t.Context()

	postgresContainer, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
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

	pg.MigrateUp(db)
	ubase_postgres.MigrateUp(db)

	storage := pg.NewPostgresStorageEngine(db)
	eventStore := evercore.NewEventStore(storage)

	// Create a new test suite
	testSuite := NewStorageEngineTestSuite(eventStore, db, ubconst.DatabaseTypePostgres)

	// Run the tests
	testSuite.RunTests(t)

}
