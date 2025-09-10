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

func UserApiKeyDeleteCommand() ubcli.Command {
    const commandName = "user-api-key-delete"

    var (
        userID int64
        apiKey string
    )

    flagset := flag.NewFlagSet(commandName, flag.ExitOnError)
    flagset.Int64Var(&userID, "user-id", 0, "User ID")
    flagset.StringVar(&apiKey, "api-key", "", "Full API key (id+secret)")

    userApiKeyDelete := func(args []string) error {
        agent := GetAgent()

        app := ubapp.NewUbaseAppEnvConfig()
        defer app.Shutdown()

        // Prompt for any missing required fields
        userID = maybeReadInt64Input("User ID: ", userID)
        apiKey = maybeReadInput("API Key: ", apiKey)

        cmd := ubmanage.UserDeleteApiKeyCommand{
            UserId: userID,
            ApiKey: apiKey,
        }

        service := app.GetManagementService()
        response, err := service.UserDeleteApiKey(context.Background(), cmd, agent)
        if err != nil {
            return err
        }

        if response.Status != ubstatus.Success {
            return fmt.Errorf("failed to delete API key: %s", response.Status)
        }

        fmt.Println("API key deleted successfully.")
        return nil
    }

    return ubcli.Command{
        Name:    commandName,
        Help:    "Delete a user's API key by providing the full API key",
        Run:     userApiKeyDelete,
        FlagSet: flagset,
    }
}

