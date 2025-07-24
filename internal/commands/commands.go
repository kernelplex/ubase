package commands

import (
	"github.com/kernelplex/ubase/lib/ubcli"
)

func GetCommands(name string) *ubcli.CommandLine {
	commandLine := ubcli.NewCommandLine(name)

	// Utility commands
	commandLine.Add(MigrateUpCommand())
	commandLine.Add(SecretCommand())

	// Organization commands
	commandLine.Add(OrganizationAddCommand())
	commandLine.Add(OrganizationListCommand())
	commandLine.Add(OrganizationUpdateCommand())

	// Role commands
	commandLine.Add(RoleAddCommand())
	commandLine.Add(RoleListCommand())
	commandLine.Add(RoleViewCommand())
	commandLine.Add(RoleAddPermissionsCommand())

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
