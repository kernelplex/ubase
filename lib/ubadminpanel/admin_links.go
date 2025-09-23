package ubadminpanel

import "github.com/kernelplex/ubase/lib/contracts"

// GetAdminLinks returns the standard admin navigation links
func GetAdminLinks() []contracts.AdminLink {
	return []contracts.AdminLink{
		{Title: "Dashboard", Icon: "home", Path: "/admin/"},
		{Title: "Organizations", Icon: "building", Path: "/admin/organizations"},
		{Title: "Users", Icon: "users", Path: "/admin/users"},
		{Title: "Roles", Icon: "key", Path: "/admin/roles"},
	}
}
