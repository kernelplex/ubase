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

func OrganizationUpdateCommand() ubcli.Command {
	const commandName = "organization-update"
	flagset := flag.NewFlagSet(commandName, flag.ExitOnError)

	var id int64
	var name string
	var systemName string

	flagset.Int64Var(&id, "id", 0, "ID of the organization")
	flagset.StringVar(&name, "name", "", "Name of the organization")
	flagset.StringVar(&systemName, "system-name", "", "System name of the organization")

	organizationUpdate := func(args []string) error {
		agent := GetAgent()
		if id == 0 {
			return fmt.Errorf("id is required")
		}

		app := ubapp.NewUbaseAppEnvConfig()
		defer app.Shutdown()
		service := app.GetManagementService()

		organizationUpdateCommand := ubmanage.OrganizationUpdateCommand{
			Id:         id,
			Name:       maybeReadOptionalInput("Name: ", name),
			SystemName: maybeReadOptionalInput("System name: ", systemName),
		}

		response, err := service.OrganizationUpdate(context.Background(), organizationUpdateCommand, agent)

		if err != nil {
			return err
		}

		if response.Status != ubstatus.Success {
			return fmt.Errorf("failed to update organization: %s", response.Status)
		}
		fmt.Printf("Organization Id: %d updated.\n", id)
		return nil
	}

	return ubcli.Command{
		Name:    commandName,
		Help:    "Update an organization",
		Run:     organizationUpdate,
		FlagSet: flagset,
	}
}
