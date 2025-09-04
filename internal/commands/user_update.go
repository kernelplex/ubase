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

func UserVerifyCommand() ubcli.Command {
	const commandName = "user-verify"

	var (
		id int64
	)

	flagset := flag.NewFlagSet(commandName, flag.ExitOnError)
	flagset.Int64Var(&id, "id", 0, "User ID to verify")

	userVerify := func(args []string) error {
		agent := GetAgent()

		id = maybeReadInt64Input("User ID: ", id)

		app := ubapp.NewUbaseAppEnvConfig()
		defer app.Shutdown()

		verified := true
		command := ubmanage.UserUpdateCommand{
			Id:       id,
			Verified: &verified,
		}

		service := app.GetManagementService()
		response, err := service.UserUpdate(context.Background(), command, agent)
		if err != nil {
			return err
		}

		if response.Status != ubstatus.Success {
			return fmt.Errorf("failed to verify user: %s", response.Status)
		}

		fmt.Printf("Successfully verified user with ID: %d\n", id)
		return nil
	}

	return ubcli.Command{
		Name:    commandName,
		Help:    "Verify a user's credentials",
		Run:     userVerify,
		FlagSet: flagset,
	}
}

func UserUpdateCommand() ubcli.Command {
	const commandName = "user-update"

	var (
		id          int64
		email       string
		password    string
		firstName   string
		lastName    string
		displayName string
	)

	flagset := flag.NewFlagSet(commandName, flag.ExitOnError)
	flagset.Int64Var(&id, "id", 0, "User ID to update")
	flagset.StringVar(&email, "email", "", "User's email address")
	flagset.StringVar(&password, "password", "", "User's new password (optional)")
	flagset.StringVar(&firstName, "first-name", "", "User's first name")
	flagset.StringVar(&lastName, "last-name", "", "User's last name")
	flagset.StringVar(&displayName, "display-name", "", "User's display name")

	userUpdate := func(args []string) error {
		agent := GetAgent()

		app := ubapp.NewUbaseAppEnvConfig()
		defer app.Shutdown()

		// Prompt for any missing required fields

		id = maybeReadInt64Input("User ID: ", id)
		email = maybeReadInput("Email: ", email)
		firstName = maybeReadInput("First Name: ", firstName)
		lastName = maybeReadInput("Last Name: ", lastName)
		displayName = maybeReadInput("Display Name: ", displayName)

		pPassword, err := maybeReadPasswordInput("Password: ", password, false)
		if err != nil {
			return err
		}

		command := ubmanage.UserUpdateCommand{
			Id:          id,
			Email:       maybeReadOptionalInput("Email: ", email),
			Password:    pPassword,
			FirstName:   maybeReadOptionalInput("First Name: ", firstName),
			LastName:    maybeReadOptionalInput("Last Name: ", lastName),
			DisplayName: maybeReadOptionalInput("Display Name: ", displayName),
		}

		service := app.GetManagementService()
		response, err := service.UserUpdate(context.Background(), command, agent)
		if err != nil {
			return err
		}

		if response.Status != ubstatus.Success {
			return fmt.Errorf("failed to update user: %s", response.Status)
		}

		fmt.Printf("Successfully updated user with ID: %d\n", id)
		return nil
	}

	return ubcli.Command{
		Name:    commandName,
		Help:    "Update an existing user",
		Run:     userUpdate,
		FlagSet: flagset,
	}
}
