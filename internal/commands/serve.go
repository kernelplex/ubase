package commands

import (
	"flag"
	"log/slog"
	"os"

	"github.com/kernelplex/ubase/lib/contracts"
	"github.com/kernelplex/ubase/lib/ubadminpanel"
	"github.com/kernelplex/ubase/lib/ubapp"
	"github.com/kernelplex/ubase/lib/ubcli"
	"github.com/kernelplex/ubase/lib/ubwww"
)

func ServeCommand() ubcli.Command {
	flagset := flag.NewFlagSet("serve", flag.ExitOnError)
	var port uint
	flagset.UintVar(&port, "port", 8089, "Port to run the server on")

	serve := func(args []string) error {

		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug, // Example: Set to Debug to see all levels
		})
		logger := slog.New(handler)
		slog.SetDefault(logger)

		flagset.Parse(args)
		app := ubapp.NewUbaseAppEnvConfig()
		defer app.Shutdown()

		cookieManager := ubwww.NewCookieMonster(
			app.GetEncryptionService(),
			"ubase_auth_token",
			false,
			3600,
			contracts.CookieContextKey("auth_token"),
			contracts.IdentityContextKey("user_identity"),
		)
		prefectService := app.GetPrefectService()
		permissionMiddleware := ubwww.NewPermissionMiddleware(
			prefectService, cookieManager)
		web := ubwww.NewWebService(
			port,
			cookieManager,
			permissionMiddleware)

		config := app.GetConfig()
		dataAdapter := app.GetDBAdapter()

		ubadminpanel.RegisterAdminPanelRoutes(
			config.PrimaryOrganization,
			dataAdapter,
			web,
			app.GetManagementService(), cookieManager,
			[]string{ubadminpanel.PermSystemAdmin,
				"edit_article", "view_article"},
		)

		err := prefectService.Start()
		if err != nil {
			slog.Error("Failed to start prefect service", "error", err)
			return err
		}

		err = web.Start()
		if err != nil {
			slog.Error("Failed to start web server", "error", err)
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
