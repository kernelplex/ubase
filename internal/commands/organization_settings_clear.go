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

func OrganizationSettingsClearCommand() ubcli.Command {
    const commandName = "organization-settings-clear"

    var id int64
    var keysArg string
    flagset := flag.NewFlagSet(commandName, flag.ExitOnError)
    flagset.Int64Var(&id, "id", 0, "ID of the organization")
    flagset.StringVar(&keysArg, "keys", "", "Comma-separated setting keys to clear (e.g., key1,key2)")

    run := func(args []string) error {
        agent := GetAgent()
        if id == 0 {
            return fmt.Errorf("id is required")
        }

        keys := parseCSVKeys(keysArg)
        if len(keys) == 0 {
            keys = promptKeys()
        }

        app := ubapp.NewUbaseAppEnvConfig()
        defer app.Shutdown()
        svc := app.GetManagementService()

        resp, err := svc.OrganizationSettingsRemove(context.Background(), ubmanage.OrganizationSettingsRemoveCommand{
            Id:          id,
            SettingKeys: keys,
        }, agent)
        if err != nil {
            return err
        }
        if resp.Status != ubstatus.Success {
            return fmt.Errorf("failed to clear settings: %s", resp.Status)
        }
        fmt.Println("Organization settings cleared for provided keys.")
        return nil
    }

    return ubcli.Command{
        Name:    commandName,
        Help:    "Clear one or more organization settings by key.",
        Run:     run,
        FlagSet: flagset,
    }
}

