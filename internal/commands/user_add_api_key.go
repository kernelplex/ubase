package commands

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/kernelplex/ubase/lib/ubapp"
	"github.com/kernelplex/ubase/lib/ubcli"
	"github.com/kernelplex/ubase/lib/ubmanage"
	"github.com/kernelplex/ubase/lib/ubstatus"
)

func UserApiKeyAddCommand() ubcli.Command {
	const commandName = "user-api-key-add"

	var (
		userID         int64
		organizationID int64
		name           string
		expiryDays     int64
	)

	flagset := flag.NewFlagSet(commandName, flag.ExitOnError)
	flagset.Int64Var(&userID, "user-id", 0, "User ID")
	flagset.Int64Var(&organizationID, "organization-id", 0, "Organization ID")
	flagset.StringVar(&name, "name", "", "API key name")
	flagset.Int64Var(&expiryDays, "expiry-days", 365, "Number of days until the API key expires")

	userApiKeyAdd := func(args []string) error {
		agent := GetAgent()

		app := ubapp.NewUbaseAppEnvConfig()
		defer app.Shutdown()

		// Prompt for any missing required fields
		userID = maybeReadInt64Input("User ID: ", userID)
		organizationID = maybeReadInt64Input("Organization ID: ", organizationID)
		name = maybeReadInput("API Key Name: ", name)

		// Calculate expiration time
		expiresAt := time.Now().Add(time.Duration(expiryDays) * 24 * time.Hour)

		command := ubmanage.UserGenerateApiKeyCommand{
			UserId:         userID,
			Name:           name,
			OrganizationId: organizationID,
			ExpiresAt:      expiresAt,
		}

		service := app.GetManagementService()
		response, err := service.UserGenerateApiKey(context.Background(), command, agent)
		if err != nil {
			return err
		}

		if response.Status != ubstatus.Success {
			return fmt.Errorf("failed to generate API key: %s", response.Status)
		}

		fmt.Printf("API Key generated: %s\n", response.Data)
		fmt.Println("Note: This is the only time the full API key will be shown. Make sure to save it securely.")
		return nil
	}

	return ubcli.Command{
		Name:    commandName,
		Help:    "Generate an API key for a user on an organization",
		Run:     userApiKeyAdd,
		FlagSet: flagset,
	}
}
