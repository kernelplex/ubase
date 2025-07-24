package commands

import (
	"flag"

	"github.com/kernelplex/ubase/lib/ubapp"
	"github.com/kernelplex/ubase/lib/ubcli"
)

func MigrateUpCommand() ubcli.Command {
	flagset := flag.NewFlagSet("migrate-up", flag.ExitOnError)

	migrateUp := func(args []string) error {
		app := ubapp.NewUbaseAppEnvConfig()
		defer app.Shutdown()
		return app.MigrateUp()
	}

	return ubcli.Command{
		Name:    "migrate-up",
		Help:    "Run database migrations up",
		Run:     migrateUp,
		FlagSet: flagset,
	}
}
