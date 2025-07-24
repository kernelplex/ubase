package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/kernelplex/ubase/lib/ubapp"
	"github.com/kernelplex/ubase/lib/ubcli"
	"github.com/kernelplex/ubase/lib/ubstatus"
	"github.com/olekukonko/tablewriter"
)

func RoleListCommand() ubcli.Command {
	const commandName = "role-list"

	var organizationId int64

	flagset := flag.NewFlagSet(commandName, flag.ExitOnError)
	flagset.Int64Var(&organizationId, "organization-id", 0, "ID of the organization to list roles for")

	roleList := func(args []string) error {
		if organizationId == 0 {
			return fmt.Errorf("organization-id is required")
		}

		app := ubapp.NewUbaseAppEnvConfig()
		defer app.Shutdown()

		service := app.GetManagementService()
		response, err := service.RoleList(context.Background(), organizationId)
		if err != nil {
			return err
		}

		if response.Status != ubstatus.Success {
			return fmt.Errorf("failed to list roles: %s", response.Status)
		}

		// Print in a table format
		columnNames := []string{"ID", "Name", "System Name"}
		table := tablewriter.NewWriter(os.Stdout)
		table.Header(columnNames)
		for _, role := range response.Data {
			table.Append([]string{
				strconv.FormatInt(role.ID, 10),
				role.Name,
				role.SystemName,
			})
		}

		table.Render()
		return nil
	}

	return ubcli.Command{
		Name:    commandName,
		Help:    "List roles for an organization",
		Run:     roleList,
		FlagSet: flagset,
	}
}
