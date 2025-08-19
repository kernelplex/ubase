package integration_tests

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"

	// "github.com/kernelplex/ubase/lib/ubconst"
	evercore "github.com/kernelplex/evercore/base"
	"github.com/kernelplex/evercore/evercoresqlite"
	"github.com/kernelplex/ubase/lib/ubdata"
	"github.com/kernelplex/ubase/sql/sqlite"
)

var ENV_SQLITE_TEST_DB = "SQLITE_TEST_DB"
var ENV_TEST_SQLITE_EVENTSTORE_DB = "SQLITE_TEST_EVENTSTORE_DB"

func openDatabase(connectionString string) (*sql.DB, error) {

	db, err := sql.Open("sqlite", connectionString)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("PRAGMA synchronous=NORMAL;")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("PRAGMA jounal_size_limit=6144000;")
	if err != nil {
		panic(err)
	}
	return db, nil
}

func ReadDotEnv() {
	// Search for .env file in current directory and parent directories
	path := "."
	for _ = range 10 { // Limit to 10 levels up to prevent infinite loops
		envPath := path + "/.env"
		if _, err := os.Stat(envPath); err == nil {
			if err := godotenv.Load(envPath); err == nil {
				return
			}
		}
		path = path + "/.."
	}

	// If no .env found, continue without error since it's optional for SQLite
}

func CleanupExistingDatabases(testDbFile string, testEventstoreDbFile string) {
	os.Remove(testDbFile)
	os.Remove(testEventstoreDbFile)
}

func TestSqliteDataAdapter(t *testing.T) {
	// Print the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	ReadDotEnv()
	testDbFile, ok := os.LookupEnv(ENV_SQLITE_TEST_DB)
	if !ok {
		t.Fatalf("Missing environment variable %s", ENV_SQLITE_TEST_DB)
	}

	testEventstoreDbFile, ok := os.LookupEnv(ENV_TEST_SQLITE_EVENTSTORE_DB)
	if !ok {
		t.Fatalf("Missing environment variable %s", ENV_TEST_SQLITE_EVENTSTORE_DB)
	}

	CleanupExistingDatabases(testDbFile, testEventstoreDbFile)

	db, err := openDatabase(testDbFile)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	err = ubase_sqlite.MigrateUp(db)
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
		return
	}

	fmt.Printf("Current working directory: %s\n", cwd)

	adapter := ubdata.NewSQLiteAdapter(db)
	testSuite := NewAdapterExercises(db, adapter)

	// CleanupExistingDatabases(testDbFile, testEventstoreDbFile)

	// Run the tests
	testSuite.RunTests(t)

	/*

		edb, err := openDatabase(testEventstoreDbFile)
		if err != nil {
			t.Fatalf("Failed to open eventstore database: %v", err)
		}
		defer edb.Close()

		if err := evercoresqlite.MigrateUp(edb); err != nil {
			t.Fatalf("Failed to migrate eventstore database: %v", err)
		}

		storage := evercoresqlite.NewSqliteStorageEngine(edb)
		eventStore := evercore.NewEventStore(storage)

		// Create a new test suite
		// testSuite := NewStorageEngineTestSuite(eventStore, db, dbType)
		adapter := ubdata.NewSQLiteAdapter(db)
		testSuite := NewAdapterExercises(db, adapter)

		// Run the tests
		testSuite.RunTests(t)
	*/
}

func TestSqliteManagementService(t *testing.T) {
	// Print the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	ReadDotEnv()
	testDbFile, ok := os.LookupEnv(ENV_SQLITE_TEST_DB)
	if !ok {
		t.Fatalf("Missing environment variable %s", ENV_SQLITE_TEST_DB)
	}

	testEventstoreDbFile, ok := os.LookupEnv(ENV_TEST_SQLITE_EVENTSTORE_DB)
	if !ok {
		t.Fatalf("Missing environment variable %s", ENV_TEST_SQLITE_EVENTSTORE_DB)
	}

	CleanupExistingDatabases(testDbFile, testEventstoreDbFile)

	db, err := openDatabase(testDbFile)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()
	ubase_sqlite.MigrateUp(db)

	fmt.Printf("Current working directory: %s\n", cwd)

	adapter := ubdata.NewSQLiteAdapter(db)

	// CleanupExistingDatabases(testDbFile, testEventstoreDbFile)

	edb, err := openDatabase(testEventstoreDbFile)
	if err != nil {
		t.Fatalf("Failed to open eventstore database: %v", err)
	}
	defer edb.Close()

	if err := evercoresqlite.MigrateUp(edb); err != nil {
		t.Fatalf("Failed to migrate eventstore database: %v", err)
	}

	storage := evercoresqlite.NewSqliteStorageEngine(edb)
	eventStore := evercore.NewEventStore(storage)
	testSuite := NewManagementServiceTestSuite(eventStore, adapter)

	// Run the tests
	testSuite.RunTests(t)

}
