package ubadminpanel

import (
	"embed"
	"net/http"
	"strings"

	"github.com/kernelplex/ubase/lib/ubadminpanel/templ/views"
	"github.com/kernelplex/ubase/lib/ubapp"
	"github.com/kernelplex/ubase/lib/ubmanage"
	"github.com/kernelplex/ubase/lib/ubwww"
)

//go:embed static
var static embed.FS

func isHTMX(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("HX-Request"), "true")
}

// hello route is a simple placeholder home page
func AdminRoute() ubwww.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		component := views.Hello("Buddy", isHTMX(r))
		_ = component.Render(r.Context(), w)
	}

	return ubwww.Route{
		Path:               "/admin",
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
	web.AddRoute(AdminRoute())
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
