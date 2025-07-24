package commands

import (
	"context"
	"flag"
	"fmt"
	"strconv"

	"github.com/kernelplex/ubase/lib/ubapp"
	"github.com/kernelplex/ubase/lib/ubcli"
	"github.com/kernelplex/ubase/lib/ubmanage"
	"github.com/kernelplex/ubase/lib/ubstatus"
)

func RoleAddCommand() ubcli.Command {
	const commandName = "role-add"

	var name string
	var systemName string

	var organizationId int64

	flagset := flag.NewFlagSet(commandName, flag.ExitOnError)
	flagset.StringVar(&systemName, "system-name", "", "System name of the role")
	flagset.StringVar(&name, "name", "", "Name of the role")
	flagset.Int64Var(&organizationId, "organization-id", 0, "ID of the organization this role belongs to")

	roleAdd := func(args []string) error {
		agent := GetAgent()

		app := ubapp.NewUbaseAppEnvConfig()
		defer app.Shutdown()
		name = maybeReadInput("Name: ", name)
		systemName = maybeReadInput("System name: ", systemName)

		if organizationId == 0 {
			orgIdStr := maybeReadInput("Organization ID: ", "")
			var err error
			organizationId, err = strconv.ParseInt(orgIdStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid organization ID: must be a number")
			}
			if organizationId <= 0 {
				return fmt.Errorf("organization ID must be positive")
			}
		}

		roleAddCommand := ubmanage.RoleCreateCommand{
			Name:           name,
			SystemName:     systemName,
			OrganizationId: organizationId,
		}

		service := app.GetManagementService()

		response, err := service.RoleAdd(context.Background(), roleAddCommand, agent)
		if err != nil {
			return err
		}

		if response.Status != ubstatus.Success {
			return fmt.Errorf("failed to add role: %s", response.Status)
		}
		fmt.Printf("Role Id: %d added.\n", response.Data.Id)
		return nil
	}

	return ubcli.Command{
		Name:    commandName,
		Help:    "Add a role",
		Run:     roleAdd,
		FlagSet: flagset,
	}
}
