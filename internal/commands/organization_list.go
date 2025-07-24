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

func OrganizationListCommand() ubcli.Command {
	const commandName = "organization-list"

	flagset := flag.NewFlagSet(commandName, flag.ExitOnError)

	organizationList := func(args []string) error {
		app := ubapp.NewUbaseAppEnvConfig()
		defer app.Shutdown()

		service := app.GetManagementService()
		response, err := service.OrganizationList(context.Background())
		if err != nil {
			return err
		}

		if response.Status != ubstatus.Success {
			return fmt.Errorf("failed to list organizations: %s", response.Status)
		}

		// Print in a table format
		columnNames := []string{"ID", "Name", "System Name", "Status"}
		table := tablewriter.NewWriter(os.Stdout)
		table.Header(columnNames)
		for _, org := range response.Data {
			table.Append([]string{
				strconv.FormatInt(org.ID, 10),
				org.Name,
				org.SystemName,
				org.Status,
			})
		}

		table.Render()

		return nil
	}

	return ubcli.Command{
		Name:    commandName,
		Help:    "List organizations",
		Run:     organizationList,
		FlagSet: flagset,
	}
}
