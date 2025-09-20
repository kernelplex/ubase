package ubadminpanel

import (
	"embed"
	"net/http"
	"strings"

	"github.com/kernelplex/ubase/lib/contracts"
	"github.com/kernelplex/ubase/lib/ensure"
	"github.com/kernelplex/ubase/lib/ubadminpanel/templ/views"
	"github.com/kernelplex/ubase/lib/ubdata"
	"github.com/kernelplex/ubase/lib/ubmanage"
	"github.com/kernelplex/ubase/lib/ubstatus"
	"github.com/kernelplex/ubase/lib/ubwww"
)

//go:embed static
var Static embed.FS

func isHTMX(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("HX-Request"), "true")
}

// hello route is a simple placeholder home page
func AdminRoute(adapter ubdata.DataAdapter, mgmt ubmanage.ManagementService) contracts.Route {
	ensure.That(adapter != nil, "data adapter is required")

	handler := func(w http.ResponseWriter, r *http.Request) {
		orgCountResp, _ := mgmt.OrganizationsCount(r.Context())
		userCountResp, _ := mgmt.UsersCount(r.Context())
		var orgs int64
		if orgCountResp.Status == ubstatus.Success {
			orgs = orgCountResp.Data
		}
		var users int64
		if userCountResp.Status == ubstatus.Success {
			users = userCountResp.Data
		}

		// Compute total roles across all organizations
		var roles int64
		orgsResp, _ := mgmt.OrganizationList(r.Context())
		if orgsResp.Status == ubstatus.Success {
			for _, o := range orgsResp.Data {
				rolesResp, _ := mgmt.RoleList(r.Context(), o.ID)
				if rolesResp.Status == ubstatus.Success {
					roles += int64(len(rolesResp.Data))
				}
			}
		}

		// Build recent users (top 5 by last_login)
		ids, _ := adapter.ListRecentUserIds(r.Context(), 5)
		recent := make([]views.RecentUser, 0, len(ids))
		for _, uid := range ids {
			uref, _ := mgmt.UserGetById(r.Context(), uid)
			if uref.Status != ubstatus.Success {
				continue
			}
			st := uref.Data.State
			recent = append(recent, views.RecentUser{ID: uid, DisplayName: st.DisplayName, Email: st.Email, LastLogin: st.LastLogin})
		}

		component := views.AdminPanel(isHTMX(r), orgs, users, roles, recent)
		_ = component.Render(r.Context(), w)
	}

	return contracts.Route{
		Path:               "/admin/",
		RequiresPermission: PermSystemAdmin,
		Func:               handler,
	}
}

// RegisterAdminPanelRoutes adds admin routes and static files.
// cookieManager is used to read/write the auth cookie; mgmt performs authentication.
func RegisterAdminPanelRoutes(
	primaryOrganization int64,
	adapter ubdata.DataAdapter,
	web ubwww.WebService,
	mgmt ubmanage.ManagementService,
	cookieManager contracts.AuthTokenCookieManager,
	permissions []string,
) {
	ensure.That(primaryOrganization > 0, "primary organization must be set and greater than zero")
	ensure.That(adapter != nil, "data adapter is required")

	// Serve static files
	fs := http.FileServer(http.FS(Static))
	web.AddRouteHandler("/admin/static/", http.StripPrefix("/admin", fs))

	// Home placeholder (can be permission-protected later)
	web.AddRoute(AdminRoute(adapter, mgmt))
	web.AddRoute(OrganizationsRoute(mgmt))
	web.AddRoute(OrganizationOverviewRoute(mgmt))
	web.AddRoute(OrganizationCreateRoute(mgmt))
	web.AddRoute(OrganizationCreatePostRoute(mgmt))
	web.AddRoute(OrganizationEditRoute(mgmt))
	web.AddRoute(RoleOverviewRoute(adapter, mgmt, permissions))
	web.AddRoute(RoleUsersListRoute(adapter))
	web.AddRoute(RoleUsersAddRoute(adapter, mgmt))
	web.AddRoute(RoleUsersRemoveRoute(adapter, mgmt))
	web.AddRoute(RolePermissionsListRoute(adapter, permissions))
	web.AddRoute(RolePermissionsAddRoute(adapter, mgmt))
	web.AddRoute(RolePermissionsRemoveRoute(adapter, mgmt))
	web.AddRoute(RoleCreateRoute(mgmt))
	web.AddRoute(RoleCreatePostRoute(mgmt))
	web.AddRoute(RoleEditRoute(mgmt))
	web.AddRoute(RoleEditPostRoute(mgmt))
	web.AddRoute(UsersListRoute(adapter))
	web.AddRoute(UserOverviewRoute(mgmt))
	web.AddRoute(UserRolesListRoute(mgmt))
	web.AddRoute(UserRolesAddRoute(mgmt))
	web.AddRoute(UserRolesRemoveRoute(mgmt))
	web.AddRoute(UserCreateRoute(mgmt))
	web.AddRoute(UserCreatePostRoute(mgmt))
	web.AddRoute(UserEditRoute(mgmt))
	web.AddRoute(LoginRoute(primaryOrganization, mgmt, cookieManager))
	web.AddRoute(VerifyTwoFactorRoute(mgmt, cookieManager))
	web.AddRoute(LogoutRoute(cookieManager))
}
