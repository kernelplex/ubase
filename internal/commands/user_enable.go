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

func UserEnableCommand() ubcli.Command {
	const commandName = "user-enable"

	var userId int64

	flagset := flag.NewFlagSet(commandName, flag.ExitOnError)
	flagset.Int64Var(&userId, "user-id", 0, "ID of the user to enable")

	userEnable := func(args []string) error {
		agent := GetAgent()

		// Prompt for missing required field
		userId = maybeReadInt64Input("User ID: ", userId)

		app := ubapp.NewUbaseAppEnvConfig()
		defer app.Shutdown()

		command := ubmanage.UserEnableCommand{
			Id: userId,
		}

		service := app.GetManagementService()
		response, err := service.UserEnable(context.Background(), command, agent)
		if err != nil {
			return err
		}

		if response.Status != ubstatus.Success {
			return fmt.Errorf("failed to enable user: %s", response.Status)
		}

		fmt.Printf("Successfully enabled user %d\n", userId)
		return nil
	}

	return ubcli.Command{
		Name:    commandName,
		Help:    "Enable a disabled user account",
		Run:     userEnable,
		FlagSet: flagset,
	}
}
