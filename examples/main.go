package main

import (
	// "context"

	"context"
	"database/sql"
	"log/slog"

	evercore "github.com/kernelplex/evercore/base"
	"github.com/kernelplex/evercore/evercoresqlite"

	"github.com/kernelplex/ubase/lib"
	"github.com/kernelplex/ubase/lib/dbpostgres"
	"github.com/kernelplex/ubase/lib/dbsqlite"
	"github.com/kernelplex/ubase/lib/ubsecurity"
	"github.com/kernelplex/ubase/sql/sqlite"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/pressly/goose/v3"
)

func main() {
	x := dbsqlite.AddUserParams{
		UserID:      1,
		FirstName:   "Charles",
		LastName:    "Havez",
		DisplayName: "Charles Havez",
		Email:       "chavez@example.com",
	}

	var y dbpostgres.AddUserParams

	y = (dbpostgres.AddUserParams)(x)
	slog.Info("y", "y", y)

	eventstore_db := "./.eventstore.db"
	user_db := "./.user.db"

	edb, err := sql.Open("sqlite3", eventstore_db)
	if err != nil {
		panic(err)
	}
	if err := evercoresqlite.MigrateUp(edb); err != nil {
		panic(err)
	}

	udb, err := sql.Open("sqlite3", user_db)
	if err != nil {
		panic(err)
	}
	ubase_sqlite.MigrateUp(udb)

	storage := evercoresqlite.NewSqliteStorageEngine(edb)

	eventStore := evercore.NewEventStore(storage)
	hashService := ubsecurity.DefaultArgon2Id

	userService := ubase.CreateUserService(eventStore, hashService, udb)

	newUser := ubase.UserCreateCommand{
		Email:       "chavez@example.com",
		Password:    "SuperSecretPassword123!!!",
		FirstName:   "Charles",
		LastName:    "Havez",
		DisplayName: "Charles Havez",
	}

	ctx := context.Background()

	res, err := userService.CreateUser(ctx, newUser, "cli:example")
	slog.Info("user created", "id", res.Id, "response", res)
}
