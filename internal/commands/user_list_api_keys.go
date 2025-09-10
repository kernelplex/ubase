package commands

import (
    "context"
    "flag"
    "fmt"
    "time"

    "github.com/kernelplex/ubase/lib/ubapp"
    "github.com/kernelplex/ubase/lib/ubcli"
)

func UserApiKeyListCommand() ubcli.Command {
    const commandName = "user-api-key-list"

    var (
        userID int64
    )

    flagset := flag.NewFlagSet(commandName, flag.ExitOnError)
    flagset.Int64Var(&userID, "user-id", 0, "User ID")

    userApiKeyList := func(args []string) error {
        _ = GetAgent() // not used, kept for parity

        app := ubapp.NewUbaseAppEnvConfig()
        defer app.Shutdown()

        userID = maybeReadInt64Input("User ID: ", userID)

        adapter := app.GetDBAdapter()
        ctx := context.Background()
        apiKeys, err := adapter.UserListApiKeys(ctx, userID)
        if err != nil {
            return fmt.Errorf("failed to list API keys: %w", err)
        }

        if len(apiKeys) == 0 {
            fmt.Println("No API keys found.")
            return nil
        }

        fmt.Printf("Found %d API key(s):\n", len(apiKeys))
        fmt.Println("ID           | Name                  | OrgID | Created At          | Expires At")
        fmt.Println("-------------+-----------------------+-------+---------------------+---------------------")
        for _, k := range apiKeys {
            created := k.CreatedAt.Format(time.RFC3339)
            expires := k.ExpiresAt.Format(time.RFC3339)
            fmt.Printf("%-12s | %-21s | %-5d | %-19s | %-19s\n",
                k.Id, k.Name, k.OrganizationID, created, expires)
        }
        return nil
    }

    return ubcli.Command{
        Name:    commandName,
        Help:    "List API keys for a user (shows key id, not secret)",
        Run:     userApiKeyList,
        FlagSet: flagset,
    }
}

