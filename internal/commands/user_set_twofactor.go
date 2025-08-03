package commands

import (
	"context"
	"flag"
	"fmt"

	"github.com/kernelplex/ubase/lib/ubapp"
	"github.com/kernelplex/ubase/lib/ubcli"
	"github.com/kernelplex/ubase/lib/ubmanage"
	"github.com/kernelplex/ubase/lib/ubstatus"
)

func UserSetTwoFactorSharedSecretCommand() ubcli.Command {
	const commandName = "user-set-twofactor"

	var (
		userId int64
		uri    string
	)

	flagset := flag.NewFlagSet(commandName, flag.ExitOnError)
	flagset.Int64Var(&userId, "user-id", 0, "User ID")
	flagset.StringVar(&uri, "uri", "", "Full totp uri")

	userAdd := func(args []string) error {
		agent := GetAgent()

		app := ubapp.NewUbaseAppEnvConfig()
		defer app.Shutdown()

		if userId == 0 {
			return fmt.Errorf("user-id is required")
		}

		if uri == "" {
			return fmt.Errorf("uri is required")
		}

		command := ubmanage.UserSetTwoFactorSharedSecretCommand{
			Id:     userId,
			Secret: uri,
		}

		service := app.GetManagementService()
		response, err := service.UserSetTwoFactorSharedSecret(context.Background(), command, agent)
		if err != nil {
			return err
		}

		if response.Status != ubstatus.Success {
			return fmt.Errorf("failed to set two factor secret: %s", response.Status)
		}

		fmt.Printf("Two factor secret set.\n")
		return nil
	}

	return ubcli.Command{
		Name:    commandName,
		Help:    "Add a new user",
		Run:     userAdd,
		FlagSet: flagset,
	}
}
