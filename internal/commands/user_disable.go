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

func UserDisableCommand() ubcli.Command {
	const commandName = "user-disable"

	var userId int64

	flagset := flag.NewFlagSet(commandName, flag.ExitOnError)
	flagset.Int64Var(&userId, "user-id", 0, "ID of the user to disable")

	userDisable := func(args []string) error {
		agent := GetAgent()

		// Prompt for missing required field
		userId = maybeReadInt64Input("User ID: ", userId)

		app := ubapp.NewUbaseAppEnvConfig()
		defer app.Shutdown()

		command := ubmanage.UserDisableCommand{
			Id: userId,
		}

		service := app.GetManagementService()
		response, err := service.UserDisable(context.Background(), command, agent)
		if err != nil {
			return err
		}

		if response.Status != ubstatus.Success {
			return fmt.Errorf("failed to disable user: %s", response.Status)
		}

		fmt.Printf("Successfully disabled user %d\n", userId)
		return nil
	}

	return ubcli.Command{
		Name:    commandName,
		Help:    "Disable a user account",
		Run:     userDisable,
		FlagSet: flagset,
	}
}
