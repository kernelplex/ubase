package ubadminpanel

import (
	"embed"
	"net/http"
	"strings"

	"github.com/kernelplex/ubase/lib/ubadminpanel/templ/views"
	"github.com/kernelplex/ubase/lib/ubapp"
	"github.com/kernelplex/ubase/lib/ubmanage"
	"github.com/kernelplex/ubase/lib/ubstatus"
	"github.com/kernelplex/ubase/lib/ubwww"
)

//go:embed static
var static embed.FS

func isHTMX(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("HX-Request"), "true")
}

// hello route is a simple placeholder home page
func AdminRoute(app *ubapp.UbaseApp, mgmt ubmanage.ManagementService) ubwww.Route {

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
		adapter := app.GetDBAdapter()
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

		component := views.Hello(isHTMX(r), orgs, users, roles, recent)
		_ = component.Render(r.Context(), w)
	}

	return ubwww.Route{
		Path:               "/admin/",
		RequiresPermission: PermSystemAdmin,
		Func:               handler,
	}
}

// RegisterAdminPanelRoutes adds admin routes and static files.
// cookieManager is used to read/write the auth cookie; mgmt performs authentication.
func RegisterAdminPanelRoutes(
	app *ubapp.UbaseApp,
	web ubwww.WebService,
	mgmt ubmanage.ManagementService,
	cookieManager ubwww.AuthTokenCookieManager[*ubwww.AuthToken],
	permissions []string,
) {

	// Serve static files
	fs := http.FileServer(http.FS(static))
	web.AddRouteHandler("/admin/static/", http.StripPrefix("/admin", fs))

	// Home placeholder (can be permission-protected later)
	web.AddRoute(AdminRoute(app, mgmt))
	web.AddRoute(OrganizationsRoute(mgmt))
	web.AddRoute(OrganizationOverviewRoute(mgmt))
	web.AddRoute(OrganizationCreateRoute(mgmt))
	web.AddRoute(OrganizationEditRoute(mgmt))
	web.AddRoute(RoleOverviewRoute(app, mgmt, permissions))
	web.AddRoute(RoleCreateRoute(mgmt))
	web.AddRoute(RoleEditRoute(mgmt))
	web.AddRoute(UsersRoute(app))
	web.AddRoute(UserOverviewRoute(mgmt))
	web.AddRoute(UserCreateRoute(mgmt))
	web.AddRoute(UserEditRoute(mgmt))
	web.AddRoute(LoginRoute(app, mgmt, cookieManager))
	web.AddRoute(VerifyTwoFactorRoute(mgmt, cookieManager))
	web.AddRoute(LogoutRoute(cookieManager))
}
