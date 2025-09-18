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

func UserSettingsSetCommand() ubcli.Command {
    const commandName = "user-settings-set"

    var id int64
    var settingsArg string
    flagset := flag.NewFlagSet(commandName, flag.ExitOnError)
    flagset.Int64Var(&id, "id", 0, "ID of the user")
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

        resp, err := svc.UserSettingsAdd(context.Background(), ubmanage.UserSettingsAddCommand{
            Id:       id,
            Settings: settings,
        }, agent)
        if err != nil {
            return err
        }
        if resp.Status != ubstatus.Success {
            return fmt.Errorf("failed to set user settings: %s", resp.Status)
        }
        fmt.Println("User settings updated.")
        return nil
    }

    return ubcli.Command{
        Name:    commandName,
        Help:    "Set one or more user settings (merge).",
        Run:     run,
        FlagSet: flagset,
    }
}

