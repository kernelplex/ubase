package commands

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/kernelplex/ubase/lib/ubapp"
	"github.com/kernelplex/ubase/lib/ubcli"
	"github.com/kernelplex/ubase/lib/ubmanage"
	"github.com/kernelplex/ubase/lib/ubstatus"
)

func readInputWithEmptyAllowed(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func RoleAddPermissionsCommand() ubcli.Command {
	const commandName = "role-add-permissions"

	var roleId int64
	var permissions string

	flagset := flag.NewFlagSet(commandName, flag.ExitOnError)
	flagset.Int64Var(&roleId, "role-id", 0, "ID of the role")
	flagset.StringVar(&permissions, "permissions", "", "Comma-separated list of permissions to add")

	roleAddPermissions := func(args []string) error {
		agent := GetAgent()
		if roleId == 0 {
			return fmt.Errorf("role-id is required")
		}

		// Initialize permission list
		permissionList := make([]string, 0)

		if permissions == "" {
			// Interactive mode - prompt for permissions until empty input
			for {
				perm := readInputWithEmptyAllowed("Permission (leave empty when done): ")
				if perm == "" {
					if len(permissionList) == 0 {
						return fmt.Errorf("at least one permission is required")
					}
					break
				}
				permissionList = append(permissionList, perm)
			}
		} else {
			// Split permissions by comma and trim whitespace
			permissionList = strings.Split(permissions, ",")
			for i, p := range permissionList {
				permissionList[i] = strings.TrimSpace(p)
			}
		}

		app := ubapp.NewUbaseAppEnvConfig()
		defer app.Shutdown()
		service := app.GetManagementService()
		if permissions != "" {
			// Split permissions by comma and trim whitespace
			permissionList = strings.Split(permissions, ",")
		}
		for i, p := range permissionList {
			permissionList[i] = strings.TrimSpace(p)
		}

		// Add each permission
		for _, perm := range permissionList {
			if perm == "" {
				continue
			}

			cmd := ubmanage.RolePermissionAddCommand{
				Id:         roleId,
				Permission: perm,
			}

			response, err := service.RolePermissionAdd(context.Background(), cmd, agent)
			if err != nil {
				return fmt.Errorf("failed to add permission %s: %w", perm, err)
			}

			if response.Status != ubstatus.Success {
				return fmt.Errorf("failed to add permission %s: %s", perm, response.Status)
			}
			fmt.Printf("Added permission: %s\n", perm)
		}

		return nil
	}

	return ubcli.Command{
		Name:    commandName,
		Help:    "Add multiple permissions to a role",
		Run:     roleAddPermissions,
		FlagSet: flagset,
	}
}
