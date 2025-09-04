package commands

import (
	"github.com/kernelplex/ubase/lib/ubcli"
)

func GetCommands(name string) *ubcli.CommandLine {
	commandLine := ubcli.NewCommandLine(name)

	// Utility commands
	commandLine.Add(MigrateUpCommand())
	commandLine.Add(SecretCommand())
	commandLine.Add(TotpGenerateCommand())

	// GenerateTotpCommandOrganization commands
	commandLine.Add(OrganizationAddCommand())
	commandLine.Add(OrganizationListCommand())
	commandLine.Add(OrganizationUpdateCommand())

	// Role commands
	commandLine.Add(RoleAddCommand())
	commandLine.Add(RoleListCommand())
	commandLine.Add(RoleViewCommand())
	commandLine.Add(RoleAddPermissionsCommand())

	// User commands
	commandLine.Add(UserAddCommand())
	commandLine.Add(UserUpdateCommand())
	commandLine.Add(UserViewCommand())
	commandLine.Add(UserVerifyCommand())
	commandLine.Add(UserAddRoleCommand())
	commandLine.Add(UserRemoveRoleCommand())
	commandLine.Add(UserDisableCommand())
	commandLine.Add(UserEnableCommand())
	commandLine.Add(UserSetTwoFactorSharedSecretCommand())

	return commandLine
}
