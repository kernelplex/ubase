package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/user"
	"strconv"

	"github.com/kernelplex/ubase/lib/ubapp"
	"github.com/kernelplex/ubase/lib/ubcli"
	"github.com/kernelplex/ubase/lib/ubmanage"
	"github.com/kernelplex/ubase/lib/ubstatus"

	"github.com/olekukonko/tablewriter"
)

func GetSystemUsername() string {
	if user, err := user.Current(); err == nil {
		return user.Username
	}
	return "unknown"
}

func GetAgent() string {
	return GetSystemUsername() + "@" + os.Getenv("HOSTNAME")
}

func OrganizationUpdateCommand() ubcli.Command {
	const name = "organization-update"
	flagset := flag.NewFlagSet(name, flag.ExitOnError)

	organizationUpdate := func(args []string) error {
		panic("not implemented")
	}

	return ubcli.Command{
		Name:    name,
		Help:    "Add an organization",
		Run:     organizationUpdate,
		FlagSet: flagset,
	}

}
func maybeReadInput(prompt string, existing string) string {
	if existing != "" {
		return existing
	}

	for true {

		fmt.Print(prompt)
		var input string
		_, err := fmt.Scanln(&input)
		if err == nil {
			return input
		}
		fmt.Println("Invalid input: " + err.Error())
	}
	return ""
}

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

func GetCommands(name string) *ubcli.CommandLine {
	commandLine := ubcli.NewCommandLine(name)

	// Utility commands
	commandLine.Add(MigrateUpCommand())
	commandLine.Add(SecretCommand())

	// Organization commands
	commandLine.Add(OrganizationAddCommand())
	commandLine.Add(OrganizationUpdateCommand())
	commandLine.Add(OrganizationListCommand())

	// Role commands

	// User commands

	/*
		commandLine.Add(ServeCommand())
		commandLine.Add(GenerateSecretCommand())
		commandLine.Add(RoleAddCommand())
		commandLine.Add(RoleListCommand())
		commandLine.Add(RoleAddPermissionCommand())
		commandLine.Add(RoleRemovePermissionCommand())
		commandLine.Add(RoleShowCommand())
		commandLine.Add(PermissionListCommand())
		commandLine.Add(UserAddCommand())
		commandLine.Add(UserGetCommand())
		commandLine.Add(UserUpdateCommand())
		commandLine.Add(UserAddRoleCommand())
	*/
	return commandLine
}
