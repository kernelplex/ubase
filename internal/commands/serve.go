package commands

import (
	"flag"
	"log/slog"
	"os"

	//"github.com/kernelplex/ubase/lib/contracts"
	"github.com/kernelplex/ubase/lib/ubadminpanel"
	"github.com/kernelplex/ubase/lib/ubapp"
	"github.com/kernelplex/ubase/lib/ubcli"
	//"github.com/kernelplex/ubase/lib/ubwww"
)

func ServeCommand() ubcli.Command {
	flagset := flag.NewFlagSet("serve", flag.ExitOnError)

	serve := func(args []string) error {

		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug, // Example: Set to Debug to see all levels
		})
		logger := slog.New(handler)
		slog.SetDefault(logger)
		app := ubapp.NewUbaseAppEnvConfig()

		defer app.Shutdown()

		permissions := []string{
			ubadminpanel.PermSystemAdmin,
		}
		app.WithAdminPanel(permissions)
		err := app.StartServices()
		if err != nil {
			slog.Error("Failed to start services", "error", err)
			return err
		}
		// Block forever
		select {}
	}

	return ubcli.Command{
		Name:    "serve",
		Help:    "Run an admin server.",
		Run:     serve,
		FlagSet: flagset,
	}
}
