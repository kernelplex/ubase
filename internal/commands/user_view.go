package commands

import (
	"context"
	"flag"
	"fmt"

	"github.com/kernelplex/ubase/lib/ubapp"
	"github.com/kernelplex/ubase/lib/ubcli"
	"github.com/kernelplex/ubase/lib/ubstatus"
)

func UserViewCommand() ubcli.Command {
	const commandName = "user-view"

	var userId int64

	flagset := flag.NewFlagSet(commandName, flag.ExitOnError)
	flagset.Int64Var(&userId, "user-id", 0, "ID of the user to view")

	userView := func(args []string) error {
		if userId == 0 {
			return fmt.Errorf("user-id is required")
		}

		app := ubapp.NewUbaseAppEnvConfig()
		defer app.Shutdown()

		service := app.GetManagementService()

		response, err := service.UserGetById(context.Background(), userId)
		if err != nil {
			return err
		}

		if response.Status != ubstatus.Success {
			return fmt.Errorf("failed to get user: %s", response.Status)
		}

		// Print user details
		fmt.Printf("User ID: %d\n", response.Data.Id)
		fmt.Printf("Email: %s\n", response.Data.State.Email)
		fmt.Printf("First Name: %s\n", response.Data.State.FirstName)
		fmt.Printf("Last Name: %s\n", response.Data.State.LastName)
		fmt.Printf("Display Name: %s\n", response.Data.State.DisplayName)
		fmt.Printf("Verified: %t\n", response.Data.State.Verified)

		/*
			// Print roles in a table if available
			if len(response.Data.State.Roles) > 0 {
				fmt.Println("\nRoles:")
				table := tablewriter.NewWriter(os.Stdout)
				table.Header([]string{"Role ID", "Name", "System Name"})
				for _, role := range response.Data.State.Roles {
					table.Append([]string{
						fmt.Sprintf("%d", role.),
						role.Name,
						role.SystemName,
					})
				}
				table.Render()
			} else {
				fmt.Println("\nNo roles assigned")
			}
		*/

		return nil
	}

	return ubcli.Command{
		Name:    commandName,
		Help:    "View user details and roles",
		Run:     userView,
		FlagSet: flagset,
	}
}
