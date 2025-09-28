package ubadminpanel

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/kernelplex/ubase/lib/contracts"
	"github.com/kernelplex/ubase/lib/forms"
	"github.com/kernelplex/ubase/lib/ubadminpanel/templ/views"
	"github.com/kernelplex/ubase/lib/ubdata"
	"github.com/kernelplex/ubase/lib/ubmanage"
	"github.com/kernelplex/ubase/lib/ubstatus"
)

// orgCreateForm is used to parse organization creation form fields.
type orgCreateForm struct {
	Name       string `json:"name"`
	SystemName string `json:"system_name"`
	Status     string `json:"status"`
}

// orgEditForm is used to parse organization edit form fields.
type orgEditForm struct {
	Name       string `json:"name"`
	SystemName string `json:"system_name"`
	Status     string `json:"status"`
}

// OrganizationsRoute shows a searchable list of organizations.
func OrganizationsRoute(mgmt ubmanage.ManagementService,
	adminLinkService contracts.AdminLinkService,
) contracts.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		resp, err := mgmt.OrganizationList(r.Context())
		if err != nil {
			slog.Error("organization list error", "error", err)
			http.Error(w, "Failed to load organizations", http.StatusInternalServerError)
			return
		}
		orgs := resp.Data
		if q != "" {
			qq := strings.ToLower(q)
			tmp := make([]ubdata.Organization, 0, len(orgs))
			for _, o := range orgs {
				if strings.Contains(strings.ToLower(o.Name), qq) || strings.Contains(strings.ToLower(o.SystemName), qq) {
					tmp = append(tmp, o)
				}
			}
			orgs = tmp
		}

		if isHTMX(r) {
			_ = views.OrganizationsTable(orgs).Render(r.Context(), w)
			return
		}
		_ = views.OrganizationsPage(contracts.OrganizationsPageViewModel{
			BaseViewModel: contracts.BaseViewModel{
				Fragment: false,
				Links:    adminLinkService.GetLinks(r),
			},
			Organizations: orgs,
			Query:         q,
		}).Render(r.Context(), w)
	}

	return contracts.Route{
		Path:               "GET /admin/organizations",
		RequiresPermission: PermSystemAdmin,
		Func:               handler,
	}
}

// OrganizationOverviewRoute shows a single organization's overview by ID.
func OrganizationOverviewRoute(mgmt ubmanage.ManagementService,
	adminLinkService contracts.AdminLinkService,

) contracts.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			http.NotFound(w, r)
			return
		}
		resp, err := mgmt.OrganizationGet(r.Context(), id)
		if err != nil || resp.Status != ubstatus.Success {
			slog.Error("organization get error", "error", err, "id", id, "status", resp.Status)
			http.NotFound(w, r)
			return
		}
		name := resp.Data.State.Name
		systemName := resp.Data.State.SystemName
		rolesResp, rerr := mgmt.OrganizationRolesWithUserCount(r.Context(), id)
		var roles []ubdata.ListRolesWithUserCountsRow
		if rerr != nil || rolesResp.Status != ubstatus.Success {
			slog.Error("roles list error", "error", rerr, "org", id)
			roles = []ubdata.ListRolesWithUserCountsRow{}
		} else {
			roles = rolesResp.Data
		}
		_ = views.OrganizationOverview(contracts.OrganizationOverviewViewModel{
			BaseViewModel: contracts.BaseViewModel{
				Fragment: false,
				Links:    adminLinkService.GetLinks(r),
			},
			ID:         id,
			Name:       name,
			SystemName: systemName,
			Roles:      roles,
		}).Render(r.Context(), w)
	}

	return contracts.Route{
		Path:               "GET /admin/organizations/{id}",
		RequiresPermission: PermSystemAdmin,
		Func:               handler,
	}
}

// OrganizationCreateRoute renders and processes the add organization form.
func OrganizationCreateRoute(mgmt ubmanage.ManagementService,
	adminLinkService contracts.AdminLinkService) contracts.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		_ = views.OrganizationForm(contracts.OrganizationFormViewModel{
			BaseViewModel: contracts.BaseViewModel{
				Fragment: isHTMX(r),
				Links:    adminLinkService.GetLinks(r),
			},
			IsEdit:       false,
			Organization: nil,
			Error:        "",
			FieldErrors:  nil,
		}).Render(r.Context(), w)
	}

	return contracts.Route{
		Path:               "GET /admin/organizations/new",
		RequiresPermission: PermSystemAdmin,
		Func:               handler,
	}
}

func OrganizationCreatePostRoute(
	mgmt ubmanage.ManagementService,
	adminLinkService contracts.AdminLinkService,
) contracts.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		var f orgCreateForm
		if err := forms.ParseFormToStruct(r, &f); err != nil {
			_ = views.OrganizationForm(contracts.OrganizationFormViewModel{
				BaseViewModel: contracts.BaseViewModel{
					Fragment: isHTMX(r),
					Links:    adminLinkService.GetLinks(r),
				},
				IsEdit:       false,
				Organization: nil,
				Error:        "Invalid form submission",
				FieldErrors:  nil,
			}).Render(r.Context(), w)
			return
		}
		name := strings.TrimSpace(f.Name)
		sys := strings.TrimSpace(f.SystemName)
		status := strings.TrimSpace(f.Status)
		resp, err := mgmt.OrganizationAdd(r.Context(), ubmanage.OrganizationCreateCommand{Name: name, SystemName: sys, Status: status}, "web:ubadminpanel")
		if err != nil || resp.Status != ubstatus.Success {
			if err != nil {
				slog.Error("org add error", "error", err)
			}
			errMap := resp.GetValidationMap()
			msg := resp.Message
			draft := ubdata.Organization{Name: name, SystemName: sys, Status: status}
			_ = views.OrganizationForm(contracts.OrganizationFormViewModel{
				BaseViewModel: contracts.BaseViewModel{
					Fragment: isHTMX(r),
					Links:    adminLinkService.GetLinks(r),
				},
				IsEdit:       false,
				Organization: &draft,
				Error:        msg,
				FieldErrors:  errMap,
			}).Render(r.Context(), w)
			return
		}
		dest := "/admin/organizations/" + strconv.FormatInt(resp.Data.Id, 10)
		if isHTMX(r) {
			w.Header().Set("HX-Redirect", dest)
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, dest, http.StatusSeeOther)
	}

	return contracts.Route{
		Path:               "POST /admin/organizations/new",
		RequiresPermission: PermSystemAdmin,
		Func:               handler,
	}
}

// OrganizationSettingsRoute displays the settings for an organization
func OrganizationSettingsRoute(mgmt ubmanage.ManagementService) contracts.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			http.NotFound(w, r)
			return
		}

		// Get organization settings
		orgResponse, _ := mgmt.OrganizationGet(r.Context(), id)
		if orgResponse.Status != ubstatus.Success {
			http.NotFound(w, r)
			return
		}
		settings := orgResponse.Data.State.Settings

		_ = views.OrganizationSettingsTable(id, settings).Render(r.Context(), w)
	}

	return contracts.Route{
		Path:               "GET /admin/organizations/{id}/settings",
		RequiresPermission: PermSystemAdmin,
		Func:               handler,
	}
}

// OrganizationSettingsAddRoute adds a new setting
func OrganizationSettingsAddRoute(mgmt ubmanage.ManagementService) contracts.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			http.NotFound(w, r)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		value := strings.TrimSpace(r.FormValue("value"))

		if name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}

		// Add the setting
		_, err = mgmt.OrganizationSettingsAdd(r.Context(), ubmanage.OrganizationSettingsAddCommand{
			Id:       id,
			Settings: map[string]string{name: value},
		}, "web:ubadminpanel")

		if err != nil {
			slog.Error("failed to add organization setting", "error", err)
			http.Error(w, "Failed to add setting", http.StatusInternalServerError)
			return
		}

		// Get organization settings
		orgResponse, _ := mgmt.OrganizationGet(r.Context(), id)
		if orgResponse.Status != ubstatus.Success {
			http.NotFound(w, r)
			return
		}
		settings := orgResponse.Data.State.Settings
		_ = views.OrganizationSettingsTable(id, settings).Render(r.Context(), w)
	}

	return contracts.Route{
		Path:               "POST /admin/organizations/{id}/settings/add",
		RequiresPermission: PermSystemAdmin,
		Func:               handler,
	}
}

// OrganizationSettingsRemoveRoute removes a setting
func OrganizationSettingsRemoveRoute(mgmt ubmanage.ManagementService) contracts.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			http.NotFound(w, r)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		if name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}

		// Remove the setting
		_, err = mgmt.OrganizationSettingsRemove(r.Context(), ubmanage.OrganizationSettingsRemoveCommand{
			Id:          id,
			SettingKeys: []string{name},
		}, "web:ubadminpanel")

		if err != nil {
			slog.Error("failed to remove organization setting", "error", err)
			http.Error(w, "Failed to remove setting", http.StatusInternalServerError)
			return
		}

		// Return empty content which will remove the row via HTMX
		w.WriteHeader(http.StatusOK)
	}

	return contracts.Route{
		Path:               "POST /admin/organizations/{id}/settings/remove",
		RequiresPermission: PermSystemAdmin,
		Func:               handler,
	}
}

// OrganizationEditRoute renders and processes the edit organization form.
func OrganizationEditRoute(mgmt ubmanage.ManagementService,
	adminLinkService contracts.AdminLinkService,
) contracts.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		if id <= 0 {
			http.NotFound(w, r)
			return
		}

		if r.Method == http.MethodGet {
			oresp, err := mgmt.OrganizationGet(r.Context(), id)
			if err != nil || oresp.Status != ubstatus.Success {
				http.NotFound(w, r)
				return
			}
			org := ubdata.Organization{ID: id, Name: oresp.Data.State.Name, SystemName: oresp.Data.State.SystemName, Status: oresp.Data.State.Status}
			_ = views.OrganizationForm(contracts.OrganizationFormViewModel{
				BaseViewModel: contracts.BaseViewModel{
					Fragment: isHTMX(r),
					Links:    adminLinkService.GetLinks(r),
				},
				IsEdit:       true,
				Organization: &org,
				Error:        "",
				FieldErrors:  nil,
			}).Render(r.Context(), w)
			return
		}

		if r.Method == http.MethodPost {
			var f orgEditForm
			if err := forms.ParseFormToStruct(r, &f); err != nil {
				_ = views.OrganizationForm(contracts.OrganizationFormViewModel{
					BaseViewModel: contracts.BaseViewModel{
						Fragment: isHTMX(r),
						Links:    adminLinkService.GetLinks(r),
					},
					IsEdit:       true,
					Organization: nil,
					Error:        "Invalid form submission",
					FieldErrors:  nil,
				}).Render(r.Context(), w)
				return
			}
			name := strings.TrimSpace(f.Name)
			sys := strings.TrimSpace(f.SystemName)
			status := strings.TrimSpace(f.Status)
			cmd := ubmanage.OrganizationUpdateCommand{Id: id, Name: &name, SystemName: &sys, Status: &status}
			uresp, err := mgmt.OrganizationUpdate(r.Context(), cmd, "web:ubadminpanel")
			if err != nil || uresp.Status != ubstatus.Success {
				errMap := uresp.GetValidationMap()
				msg := uresp.Message
				draft := ubdata.Organization{ID: id, Name: name, SystemName: sys, Status: status}
				_ = views.OrganizationForm(contracts.OrganizationFormViewModel{
					BaseViewModel: contracts.BaseViewModel{
						Fragment: isHTMX(r),
						Links:    adminLinkService.GetLinks(r),
					},
					IsEdit:       true,
					Organization: &draft,
					Error:        msg,
					FieldErrors:  errMap,
				}).Render(r.Context(), w)
				return
			}
			dest := "/admin/organizations/" + strconv.FormatInt(id, 10)
			if isHTMX(r) {
				w.Header().Set("HX-Redirect", dest)
				w.WriteHeader(http.StatusOK)
				return
			}
			http.Redirect(w, r, dest, http.StatusSeeOther)
			return
		}
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}

	return contracts.Route{
		Path:               "/admin/organizations/{id}/edit",
		RequiresPermission: PermSystemAdmin,
		Func:               handler,
	}
}
