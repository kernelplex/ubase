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

func OrganizationSettingsSetCommand() ubcli.Command {
    const commandName = "organization-settings-set"

    var id int64
    var settingsArg string
    flagset := flag.NewFlagSet(commandName, flag.ExitOnError)
    flagset.Int64Var(&id, "id", 0, "ID of the organization")
    flagset.StringVar(&settingsArg, "settings", "", "Comma-separated key=value pairs (e.g., key1=val1,key2=val2)")

    run := func(args []string) error {
        agent := GetAgent()
        if id == 0 {
            return fmt.Errorf("id is required")
        }

        settings, err := parseKeyValuePairs(settingsArg)
        if err != nil {
            return err
        }
        if len(settings) == 0 {
            settings = promptKeyValuePairs()
        }

        app := ubapp.NewUbaseAppEnvConfig()
        defer app.Shutdown()
        svc := app.GetManagementService()

        resp, err := svc.OrganizationSettingsAdd(context.Background(), ubmanage.OrganizationSettingsAddCommand{
            Id:       id,
            Settings: settings,
        }, agent)
        if err != nil {
            return err
        }
        if resp.Status != ubstatus.Success {
            return fmt.Errorf("failed to set settings: %s", resp.Status)
        }
        fmt.Println("Organization settings updated.")
        return nil
    }

    return ubcli.Command{
        Name:    commandName,
        Help:    "Set one or more organization settings (merge).",
        Run:     run,
        FlagSet: flagset,
    }
}

