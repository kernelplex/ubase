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

func UserAddCommand() ubcli.Command {
	const commandName = "user-add"

	var (
		email       string
		password    string
		firstName   string
		lastName    string
		displayName string
		verified    bool = true
	)

	flagset := flag.NewFlagSet(commandName, flag.ExitOnError)
	flagset.StringVar(&email, "email", "", "User's email address")
	flagset.StringVar(&password, "password", "", "User's password")
	flagset.StringVar(&firstName, "first-name", "", "User's first name")
	flagset.StringVar(&lastName, "last-name", "", "User's last name")
	flagset.StringVar(&displayName, "display-name", "", "User's display name")

	userAdd := func(args []string) error {
		agent := GetAgent()

		app := ubapp.NewUbaseAppEnvConfig()
		defer app.Shutdown()

		// Prompt for any missing required fields
		email = maybeReadInput("Email: ", email)
		password, err := maybeReadPasswordInput("Password: ", password, true)
		if err != nil {
			return err
		}
		firstName = maybeReadInput("First Name: ", firstName)
		lastName = maybeReadInput("Last Name: ", lastName)
		displayName = maybeReadInput("Display Name: ", displayName)

		command := ubmanage.UserCreateCommand{
			Email:       email,
			Password:    *password,
			FirstName:   firstName,
			LastName:    lastName,
			DisplayName: displayName,
			Verified:    verified,
		}

		service := app.GetManagementService()
		response, err := service.UserAdd(context.Background(), command, agent)
		if err != nil {
			return err
		}

		if response.Status != ubstatus.Success {
			return fmt.Errorf("failed to add user: %s", response.Status)
		}

		fmt.Printf("User created with ID: %d\n", response.Data.Id)
		return nil
	}

	return ubcli.Command{
		Name:    commandName,
		Help:    "Add a new user",
		Run:     userAdd,
		FlagSet: flagset,
	}
}
