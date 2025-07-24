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

func UserAddRoleCommand() ubcli.Command {
	const commandName = "user-add-role"

	var (
		userId int64
		roleId int64
	)

	flagset := flag.NewFlagSet(commandName, flag.ExitOnError)
	flagset.Int64Var(&userId, "user-id", 0, "ID of the user to add to role")
	flagset.Int64Var(&roleId, "role-id", 0, "ID of the role to add user to")

	userAddRole := func(args []string) error {
		agent := GetAgent()

		// Prompt for missing required fields
		userId = maybeReadInt64Input("User ID: ", userId)
		roleId = maybeReadInt64Input("Role ID: ", roleId)

		app := ubapp.NewUbaseAppEnvConfig()
		defer app.Shutdown()

		command := ubmanage.UserAddToRoleCommand{
			UserId: userId,
			RoleId: roleId,
		}

		service := app.GetManagementService()
		response, err := service.UserAddToRole(context.Background(), command, agent)
		if err != nil {
			return err
		}

		if response.Status != ubstatus.Success {
			return fmt.Errorf("failed to add user to role: %s", response.Status)
		}

		fmt.Printf("Successfully added user %d to role %d\n", userId, roleId)
		return nil
	}

	return ubcli.Command{
		Name:    commandName,
		Help:    "Add a user to a role",
		Run:     userAddRole,
		FlagSet: flagset,
	}
}
