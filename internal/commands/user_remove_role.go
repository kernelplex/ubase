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

func UserRemoveRoleCommand() ubcli.Command {
	const commandName = "user-remove-role"

	var (
		userId int64
		roleId int64
	)

	flagset := flag.NewFlagSet(commandName, flag.ExitOnError)
	flagset.Int64Var(&userId, "user-id", 0, "ID of the user to remove from role")
	flagset.Int64Var(&roleId, "role-id", 0, "ID of the role to remove user from")

	userRemoveRole := func(args []string) error {
		agent := GetAgent()

		// Prompt for missing required fields
		userId = maybeReadInt64Input("User ID: ", userId)
		roleId = maybeReadInt64Input("Role ID: ", roleId)

		app := ubapp.NewUbaseAppEnvConfig()
		defer app.Shutdown()

		command := ubmanage.UserRemoveFromRoleCommand{
			UserId: userId,
			RoleId: roleId,
		}

		service := app.GetManagementService()
		response, err := service.UserRemoveFromRole(context.Background(), command, agent)
		if err != nil {
			return err
		}

		if response.Status != ubstatus.Success {
			return fmt.Errorf("failed to remove user from role: %s", response.Status)
		}

		fmt.Printf("Successfully removed user %d from role %d\n", userId, roleId)
		return nil
	}

	return ubcli.Command{
		Name:    commandName,
		Help:    "Remove a user from a role",
		Run:     userRemoveRole,
		FlagSet: flagset,
	}
}
