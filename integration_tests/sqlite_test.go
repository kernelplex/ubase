// //go:build sqlite

package integration_tests

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/xo/dburl"
	"modernc.org/sqlite"

	// "github.com/kernelplex/ubase/lib/ubconst"
	evercore "github.com/kernelplex/evercore/base"
	"github.com/kernelplex/evercore/evercoresqlite"
	"github.com/kernelplex/ubase/lib/ubdata"
	"github.com/kernelplex/ubase/sql/sqlite"
)

var ENV_SQLITE_TEST_DB = "SQLITE_TEST_DB"
var ENV_TEST_SQLITE_EVENTSTORE_DB = "SQLITE_TEST_EVENTSTORE_DB"

var modernc_sqlite_registered = false

func RegisterModerncSqlite() {
	if !modernc_sqlite_registered {
		sql.Register("sqlite3", &sqlite.Driver{})
		modernc_sqlite_registered = true
	}
}

func openDatabase(connectionString string) (*sql.DB, error) {
	RegisterModerncSqlite()
	connection, err := dburl.Parse(connectionString)
	if err != nil {
		panic(err)
	}
	if connection.Driver != "sqlite3" {
		return nil, fmt.Errorf("invalid driver: %s", connection.Driver)
	}

	db, err := dburl.Open(connectionString)
	if err != nil {
		return nil, err
	}

	/*
		db, err := sql.Open("sqlite", connectionString)
	*/

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
	for range 10 { // Limit to 10 levels up to prevent infinite loops
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

func mustGetSqliteFilename(dbUrl string) string {
	parsedUrl, err := dburl.Parse(dbUrl)
	if err != nil {
		panic(err)
	}
	if parsedUrl.Driver != "sqlite3" {
		panic("Invalid driver for test database: " + parsedUrl.Driver)
	}

	filename := parsedUrl.DSN
	if idx := strings.IndexRune(filename, '?'); idx != -1 {
		filename = filename[:idx]
	}
	return filename
}

func CleanupExistingDatabases(testDbUrl string, testEventstoreDbUrl string) {
	testDbUrl = mustGetSqliteFilename(testDbUrl)
	testEventstoreDbUrl = mustGetSqliteFilename(testEventstoreDbUrl)

	os.Remove(testDbUrl)
	os.Remove(testEventstoreDbUrl)
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
