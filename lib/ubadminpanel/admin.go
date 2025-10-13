package ubadminpanel

import (
	"embed"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/kernelplex/ubase/lib/contracts"
	"github.com/kernelplex/ubase/lib/ensure"
	"github.com/kernelplex/ubase/lib/ubadminpanel/templ/layouts"
	"github.com/kernelplex/ubase/lib/ubadminpanel/templ/views"
	"github.com/kernelplex/ubase/lib/ubdata"
	"github.com/kernelplex/ubase/lib/ubmanage"
	"github.com/kernelplex/ubase/lib/ubstatus"
)

//go:embed static
var Static embed.FS

func isHTMX(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("HX-Request"), "true")
}

// IsHTMX is an exported helper for consumers to detect HTMX requests.
func IsHTMX(r *http.Request) bool { return isHTMX(r) }

type AdminRendererImpl struct {
	adminLinkService contracts.AdminLinkService
	stylesheets      []string
}

func NewAdminRenderer(adminLinkService contracts.AdminLinkService) contracts.AdminRenderer {
	ensure.That(adminLinkService != nil, "admin link service is required")
	return &AdminRendererImpl{adminLinkService: adminLinkService}
}

func (ar *AdminRendererImpl) AddStyle(css string) {
	ar.stylesheets = append(ar.stylesheets, css)
}

func (ar *AdminRendererImpl) Render(w http.ResponseWriter, r *http.Request, component templ.Component) {

	fragment := isHTMX(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	cmp := layouts.RenderComponent(fragment, true, ar.adminLinkService.GetLinks(r), ar.stylesheets, component)
	_ = cmp.Render(r.Context(), w)
}

func AdminBasicRoute(prefectService ubmanage.PrefectService,
	cookieManager contracts.AuthTokenCookieManager,
	adminLinkService contracts.AdminLinkService) contracts.Route {
	ensure.That(prefectService != nil, "prefect service is required")
	ensure.That(cookieManager != nil, "cookie manager is required")

	handler := func(w http.ResponseWriter, r *http.Request) {
		// Redirect to admin/panel if they have SystemAdmin permission
		identity, found := cookieManager.IdentityFromContext(r.Context())
		if identity.UserID == 0 || !found {
			// If unauthenticated, redirect to login. Support HTMX fragments.
			if strings.EqualFold(r.Header.Get("HX-Request"), "true") {
				w.Header().Set("HX-Redirect", "/admin/login")
				w.WriteHeader(http.StatusOK)
			} else {
				http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			}
			return
		}

		if found && identity.UserID != 0 && identity.OrganizationID != 0 {
			hasPermission, err := prefectService.UserHasPermission(r.Context(), identity.UserID, identity.OrganizationID, PermSystemAdmin)
			if err == nil && hasPermission {
				// Include hx-recirect
				w.Header().Set("HX-Redirect", "/admin/panel")
				http.Redirect(w, r, "/admin/panel", http.StatusSeeOther)
				return
			}
		}

		vm := contracts.BaseViewModel{
			Fragment: isHTMX(r),
			Links:    adminLinkService.GetLinks(r),
		}

		component := views.AdminBlank(vm)
		_ = component.Render(r.Context(), w)
	}

	return contracts.Route{
		Path: "/admin/",
		Func: handler,
	}
}

func AdminPanelRoute(
	adapter ubdata.DataAdapter,
	mgmt ubmanage.ManagementService,
	adminLinkService contracts.AdminLinkService) contracts.Route {

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
		recent := make([]contracts.RecentUser, 0, len(ids))
		for _, uid := range ids {
			uref, _ := mgmt.UserGetById(r.Context(), uid)
			if uref.Status != ubstatus.Success {
				continue
			}
			st := uref.Data.State
			recent = append(recent, contracts.RecentUser{ID: uid, DisplayName: st.DisplayName, Email: st.Email, LastLogin: st.LastLogin})
		}

		vm := contracts.AdminPanelViewModel{
			BaseViewModel: contracts.BaseViewModel{
				Fragment: isHTMX(r),
				Links:    adminLinkService.GetLinks(r),
			},
			OrgCount:  orgs,
			UserCount: users,
			RoleCount: roles,
			Recent:    recent,
		}

		component := views.AdminPanel(vm)
		_ = component.Render(r.Context(), w)
	}

	return contracts.Route{
		Path:               "/admin/panel",
		RequiresPermission: PermSystemAdmin,
		Func:               handler,
	}
}
