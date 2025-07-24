package commands

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/kernelplex/ubase/lib/ubapp"
	"github.com/kernelplex/ubase/lib/ubcli"
	"github.com/kernelplex/ubase/lib/ubstatus"
	"github.com/olekukonko/tablewriter"
)

func RoleViewCommand() ubcli.Command {
	const commandName = "role-view"

	var roleId int64

	flagset := flag.NewFlagSet(commandName, flag.ExitOnError)
	flagset.Int64Var(&roleId, "role-id", 0, "ID of the role to view")

	roleView := func(args []string) error {
		if roleId == 0 {
			return fmt.Errorf("role-id is required")
		}

		app := ubapp.NewUbaseAppEnvConfig()
		defer app.Shutdown()

		service := app.GetManagementService()
		response, err := service.RoleGetById(context.Background(), roleId)
		if err != nil {
			return err
		}

		if response.Status != ubstatus.Success {
			return fmt.Errorf("failed to get role: %s", response.Status)
		}

		// Print role details
		fmt.Printf("Role ID: %d\n", response.Data.Id)
		fmt.Printf("Name: %s\n", response.Data.State.Name)
		fmt.Printf("System Name: %s\n", response.Data.State.SystemName)
		fmt.Printf("Organization ID: %d\n", response.Data.State.OrganizationId)
		fmt.Printf("Deleted: %t\n", response.Data.State.Deleted)

		// Print permissions in a table
		if len(response.Data.State.Permissions) > 0 {
			fmt.Println("\nPermissions:")
			table := tablewriter.NewWriter(os.Stdout)
			table.Header([]string{"Permission"})
			for _, perm := range response.Data.State.Permissions {
				table.Append([]string{perm})
			}
			table.Render()
		} else {
			fmt.Println("\nNo permissions assigned")
		}

		return nil
	}

	return ubcli.Command{
		Name:    commandName,
		Help:    "View role details and permissions",
		Run:     roleView,
		FlagSet: flagset,
	}
}
