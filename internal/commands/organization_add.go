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

func OrganizationAddCommand() ubcli.Command {
	const commandName = "organization-add"

	var name string
	var systemName string

	flagset := flag.NewFlagSet(commandName, flag.ExitOnError)
	flagset.StringVar(&systemName, "system-name", "", "System name of the organization")
	flagset.StringVar(&name, "name", "", "Name of the organization")

	organizationAdd := func(args []string) error {
		agent := GetAgent()

		app := ubapp.NewUbaseAppEnvConfig()
		defer app.Shutdown()
		name = maybeReadInput("Name: ", name)
		systemName = maybeReadInput("System name: ", systemName)

		organizationAddCommand := ubmanage.OrganizationCreateCommand{
			Name:       name,
			SystemName: systemName,
			Status:     "active",
		}

		service := app.GetManagementService()

		response, err := service.OrganizationAdd(context.Background(), organizationAddCommand, agent)
		if err != nil {
			return err
		}

		if response.Status != ubstatus.Success {
			return fmt.Errorf("failed to add organization: %s", response.Status)
		}
		fmt.Printf("Organization Id: %d added.\n", response.Data.Id)
		return nil
	}

	return ubcli.Command{
		Name:    commandName,
		Help:    "Add an organization",
		Run:     organizationAdd,
		FlagSet: flagset,
	}
}
