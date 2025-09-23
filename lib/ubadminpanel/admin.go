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
)

//go:embed static
var Static embed.FS

func isHTMX(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("HX-Request"), "true")
}

// IsHTMX is an exported helper for consumers to detect HTMX requests.
func IsHTMX(r *http.Request) bool { return isHTMX(r) }

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
				Links:    GetAdminLinks(),
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
		Path:               "/admin/",
		RequiresPermission: PermSystemAdmin,
		Func:               handler,
	}
}
