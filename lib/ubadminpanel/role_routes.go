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

// roleEditForm is used to parse the role edit form fields.
type roleEditForm struct {
	Name       string `json:"name"`
	SystemName string `json:"system_name"`
}

// roleCreateForm is used to parse the role creation form fields.
type roleCreateForm struct {
	Name           string `json:"name"`
	SystemName     string `json:"system_name"`
	OrganizationId int64  `json:"organization_id"`
}

func RoleOverviewRoute(adapter ubdata.DataAdapter, mgmt ubmanage.ManagementService, permissions []string) contracts.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			http.NotFound(w, r)
			return
		}
		resp, err := mgmt.RoleGetById(r.Context(), id)
		if err != nil || resp.Status != ubstatus.Success {
			slog.Error("role get error", "error", err, "id", id, "status", resp.Status)
			http.NotFound(w, r)
			return
		}
		state := resp.Data.State
		_ = views.RoleOverview(false, id, state.Name, state.SystemName, state.OrganizationId).Render(r.Context(), w)
	}

    return contracts.Route{
        Path:               "GET /admin/roles/{id}",
        RequiresPermission: PermSystemAdmin,
        Func:               handler,
    }
}

func RoleUsersListRoute(adapter ubdata.DataAdapter) contracts.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			http.NotFound(w, r)
			return
		}
		members, merr := adapter.GetUsersInRole(r.Context(), id)
		if merr != nil {
			slog.Error("list users in role error", "error", merr, "role", id)
			members = []ubdata.User{}
		}
		memberSet := make(map[int64]bool, len(members))
		for _, u := range members {
			memberSet[u.UserID] = true
		}
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		var users []ubdata.User
		if q == "" {
			users = members
		} else {
			const limit = 25
			var serr error
			users, serr = adapter.SearchUsers(r.Context(), q, limit, 0)
			if serr != nil {
				slog.Error("search users error", "error", serr)
				users = []ubdata.User{}
			}
		}
		_ = views.RoleUsersTable(users, memberSet, id).Render(r.Context(), w)
	}
    return contracts.Route{
        Path:               "GET /admin/roles/{id}/users",
        RequiresPermission: PermSystemAdmin,
        Func:               handler,
    }
}

func RoleUsersAddRoute(adapter ubdata.DataAdapter, mgmt ubmanage.ManagementService) contracts.Route {
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
		uidStr := r.FormValue("user_id")
		uid, _ := strconv.ParseInt(uidStr, 10, 64)
		if uid <= 0 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		_, _ = mgmt.UserAddToRole(r.Context(), ubmanage.UserAddToRoleCommand{UserId: uid, RoleId: id}, "web:ubadminpanel")
		members, _ := adapter.GetUsersInRole(r.Context(), id)
		memberSet := make(map[int64]bool, len(members))
		for _, u := range members {
			memberSet[u.UserID] = true
		}
		user, gerr := adapter.GetUser(r.Context(), uid)
		if gerr != nil {
			slog.Error("get user error", "error", gerr, "user", uid)
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		_ = views.RoleUserRow(id, user, memberSet[uid]).Render(r.Context(), w)
	}
    return contracts.Route{
        Path:               "POST /admin/roles/{id}/users/add",
        RequiresPermission: PermSystemAdmin,
        Func:               handler,
    }
}

func RoleUsersRemoveRoute(adapter ubdata.DataAdapter, mgmt ubmanage.ManagementService) contracts.Route {
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
		uidStr := r.FormValue("user_id")
		uid, _ := strconv.ParseInt(uidStr, 10, 64)
		if uid <= 0 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		_, _ = mgmt.UserRemoveFromRole(r.Context(), ubmanage.UserRemoveFromRoleCommand{UserId: uid, RoleId: id}, "web:ubadminpanel")
		members, _ := adapter.GetUsersInRole(r.Context(), id)
		memberSet := make(map[int64]bool, len(members))
		for _, u := range members {
			memberSet[u.UserID] = true
		}
		user, gerr := adapter.GetUser(r.Context(), uid)
		if gerr != nil {
			slog.Error("get user error", "error", gerr, "user", uid)
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		_ = views.RoleUserRow(id, user, memberSet[uid]).Render(r.Context(), w)
	}
    return contracts.Route{
        Path:               "POST /admin/roles/{id}/users/remove",
        RequiresPermission: PermSystemAdmin,
        Func:               handler,
    }
}

func RolePermissionsListRoute(adapter ubdata.DataAdapter, permissions []string) contracts.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			http.NotFound(w, r)
			return
		}
		assigned, aerr := adapter.GetRolePermissions(r.Context(), id)
		if aerr != nil {
			slog.Error("get role permissions error", "error", aerr, "role", id)
			assigned = []string{}
		}
		memberSet := make(map[string]bool, len(assigned))
		for _, p := range assigned {
			memberSet[p] = true
		}
		q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
		filtered := make([]string, 0, len(permissions))
		if q == "" {
			filtered = append(filtered, permissions...)
		} else {
			for _, p := range permissions {
				if strings.Contains(strings.ToLower(p), q) {
					filtered = append(filtered, p)
				}
			}
		}
		_ = views.RolePermissionsTable(filtered, memberSet, id).Render(r.Context(), w)
	}
    return contracts.Route{
        Path:               "GET /admin/roles/{id}/permissions",
        RequiresPermission: PermSystemAdmin,
        Func:               handler,
    }
}

func RolePermissionsAddRoute(adapter ubdata.DataAdapter, mgmt ubmanage.ManagementService) contracts.Route {
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
		perm := strings.TrimSpace(r.FormValue("permission"))
		if perm == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		_, _ = mgmt.RolePermissionAdd(r.Context(), ubmanage.RolePermissionAddCommand{Id: id, Permission: perm}, "web:ubadminpanel")
		assigned, _ := adapter.GetRolePermissions(r.Context(), id)
		memberSet := make(map[string]bool, len(assigned))
		for _, p := range assigned {
			memberSet[p] = true
		}
		_ = views.RolePermissionRow(id, perm, memberSet[perm]).Render(r.Context(), w)
	}
    return contracts.Route{
        Path:               "POST /admin/roles/{id}/permissions/add",
        RequiresPermission: PermSystemAdmin,
        Func:               handler,
    }
}

func RolePermissionsRemoveRoute(adapter ubdata.DataAdapter, mgmt ubmanage.ManagementService) contracts.Route {
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
		perm := strings.TrimSpace(r.FormValue("permission"))
		if perm == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		_, _ = mgmt.RolePermissionRemove(r.Context(), ubmanage.RolePermissionRemoveCommand{Id: id, Permission: perm}, "web:ubadminpanel")
		assigned, _ := adapter.GetRolePermissions(r.Context(), id)
		memberSet := make(map[string]bool, len(assigned))
		for _, p := range assigned {
			memberSet[p] = true
		}
		_ = views.RolePermissionRow(id, perm, memberSet[perm]).Render(r.Context(), w)
	}
    return contracts.Route{
        Path:               "POST /admin/roles/{id}/permissions/remove",
        RequiresPermission: PermSystemAdmin,
        Func:               handler,
    }
}

func RoleCreateRoute(mgmt ubmanage.ManagementService) contracts.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		orgsResp, _ := mgmt.OrganizationList(r.Context())
		orgs := []ubdata.Organization{}
		if orgsResp.Status == ubstatus.Success {
			orgs = orgsResp.Data
		}
		selectedOrg := int64(0)
		if v := strings.TrimSpace(r.URL.Query().Get("org")); v != "" {
			if oid, err := strconv.ParseInt(v, 10, 64); err == nil {
				selectedOrg = oid
			}
		}
		if len(orgs) > 0 && selectedOrg == 0 {
			selectedOrg = orgs[0].ID
		}
		switch r.Method {
		case http.MethodGet:
			_ = views.RoleForm(isHTMX(r), false, nil, orgs, selectedOrg, "", nil).Render(r.Context(), w)
			return
		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				_ = views.RoleForm(isHTMX(r), false, nil, orgs, selectedOrg, "Invalid form submission", nil).Render(r.Context(), w)
				return
			}
			name := strings.TrimSpace(r.FormValue("name"))
			sys := strings.TrimSpace(r.FormValue("system_name"))
			orgStr := r.FormValue("organization_id")
			oid, _ := strconv.ParseInt(orgStr, 10, 64)
			cmd := ubmanage.RoleCreateCommand{Name: name, SystemName: sys, OrganizationId: oid}
			resp, err := mgmt.RoleAdd(r.Context(), cmd, "web:ubadminpanel")
			if err != nil || resp.Status != ubstatus.Success {
				if err != nil {
					slog.Error("role add error", "error", err)
				}
				errMap := resp.GetValidationMap()
				msg := resp.Message
				draft := ubdata.RoleRow{Name: name, SystemName: sys}
				_ = views.RoleForm(isHTMX(r), false, &draft, orgs, oid, msg, errMap).Render(r.Context(), w)
				return
			}
			dest := "/admin/roles/" + strconv.FormatInt(resp.Data.Id, 10)
			if isHTMX(r) {
				w.Header().Set("HX-Redirect", dest)
				w.WriteHeader(http.StatusOK)
				return
			}
			http.Redirect(w, r, dest, http.StatusSeeOther)
			return
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	}
    return contracts.Route{
        Path:               "GET /admin/roles/new",
        RequiresPermission: PermSystemAdmin,
        Func:               handler,
    }
}

func RoleEditRoute(mgmt ubmanage.ManagementService) contracts.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		roleId, _ := strconv.ParseInt(idStr, 10, 64)
		if roleId <= 0 {
			http.NotFound(w, r)
			return
		}
		orgsResp, _ := mgmt.OrganizationList(r.Context())
		orgs := []ubdata.Organization{}
		if orgsResp.Status == ubstatus.Success {
			orgs = orgsResp.Data
		}
		switch r.Method {
		case http.MethodGet:
			rresp, err := mgmt.RoleGetById(r.Context(), roleId)
			if err != nil || rresp.Status != ubstatus.Success {
				http.NotFound(w, r)
				return
			}
			st := rresp.Data.State
			draft := ubdata.RoleRow{ID: roleId, Name: st.Name, SystemName: st.SystemName}
			_ = views.RoleForm(isHTMX(r), true, &draft, orgs, st.OrganizationId, "", nil).Render(r.Context(), w)
			return
		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				_ = views.RoleForm(isHTMX(r), true, nil, orgs, 0, "Invalid form submission", nil).Render(r.Context(), w)
				return
			}
			name := strings.TrimSpace(r.FormValue("name"))
			sys := strings.TrimSpace(r.FormValue("system_name"))
			cmd := ubmanage.RoleUpdateCommand{Id: roleId, Name: &name, SystemName: &sys}
			resp, err := mgmt.RoleUpdate(r.Context(), cmd, "web:ubadminpanel")
			if err != nil || resp.Status != ubstatus.Success {
				errMap := resp.GetValidationMap()
				msg := resp.Message
				draft := ubdata.RoleRow{ID: roleId, Name: name, SystemName: sys}
				rresp, _ := mgmt.RoleGetById(r.Context(), roleId)
				selected := int64(0)
				if rresp.Status == ubstatus.Success {
					selected = rresp.Data.State.OrganizationId
				}
				_ = views.RoleForm(isHTMX(r), true, &draft, orgs, selected, msg, errMap).Render(r.Context(), w)
				return
			}
			dest := "/admin/roles/" + strconv.FormatInt(roleId, 10)
			if isHTMX(r) {
				w.Header().Set("HX-Redirect", dest)
				w.WriteHeader(http.StatusOK)
				return
			}
			http.Redirect(w, r, dest, http.StatusSeeOther)
			return
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	}
    return contracts.Route{
        Path:               "GET /admin/roles/{id}/edit",
        RequiresPermission: PermSystemAdmin,
        Func:               handler,
    }
}

func RoleCreatePostRoute(mgmt ubmanage.ManagementService) contracts.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		orgsResp, _ := mgmt.OrganizationList(r.Context())
		orgs := []ubdata.Organization{}
		if orgsResp.Status == ubstatus.Success {
			orgs = orgsResp.Data
		}
		var f roleCreateForm
		if err := forms.ParseFormToStruct(r, &f); err != nil {
			selectedOrg := int64(0)
			if len(orgs) > 0 {
				selectedOrg = orgs[0].ID
			}
			_ = views.RoleForm(isHTMX(r), false, nil, orgs, selectedOrg, "Invalid form submission", nil).Render(r.Context(), w)
			return
		}
		name := strings.TrimSpace(f.Name)
		sys := strings.TrimSpace(f.SystemName)
		oid := f.OrganizationId
		cmd := ubmanage.RoleCreateCommand{Name: name, SystemName: sys, OrganizationId: oid}
		resp, err := mgmt.RoleAdd(r.Context(), cmd, "web:ubadminpanel")
		if err != nil || resp.Status != ubstatus.Success {
			if err != nil {
				slog.Error("role add error", "error", err)
			}
			errMap := resp.GetValidationMap()
			msg := resp.Message
			draft := ubdata.RoleRow{Name: name, SystemName: sys}
			_ = views.RoleForm(isHTMX(r), false, &draft, orgs, oid, msg, errMap).Render(r.Context(), w)
			return
		}
		dest := "/admin/roles/" + strconv.FormatInt(resp.Data.Id, 10)
		if isHTMX(r) {
			w.Header().Set("HX-Redirect", dest)
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, dest, http.StatusSeeOther)
	}
    return contracts.Route{
        Path:               "POST /admin/roles/new",
        RequiresPermission: PermSystemAdmin,
        Func:               handler,
    }
}

func RoleEditPostRoute(mgmt ubmanage.ManagementService) contracts.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		idStr := r.PathValue("id")
		roleId, _ := strconv.ParseInt(idStr, 10, 64)
		if roleId <= 0 {
			http.NotFound(w, r)
			return
		}
		orgsResp, _ := mgmt.OrganizationList(r.Context())
		orgs := []ubdata.Organization{}
		if orgsResp.Status == ubstatus.Success {
			orgs = orgsResp.Data
		}
		var f roleEditForm
		if err := forms.ParseFormToStruct(r, &f); err != nil {
			_ = views.RoleForm(isHTMX(r), true, nil, orgs, 0, "Invalid form submission", nil).Render(r.Context(), w)
			return
		}
		name := f.Name
		sys := f.SystemName
		cmd := ubmanage.RoleUpdateCommand{Id: roleId, Name: &name, SystemName: &sys}
		resp, err := mgmt.RoleUpdate(r.Context(), cmd, "web:ubadminpanel")
		if err != nil || resp.Status != ubstatus.Success {
			errMap := resp.GetValidationMap()
			msg := resp.Message
			draft := ubdata.RoleRow{ID: roleId, Name: name, SystemName: sys}
			rresp, _ := mgmt.RoleGetById(r.Context(), roleId)
			selected := int64(0)
			if rresp.Status == ubstatus.Success {
				selected = rresp.Data.State.OrganizationId
			}
			_ = views.RoleForm(isHTMX(r), true, &draft, orgs, selected, msg, errMap).Render(r.Context(), w)
			return
		}
		dest := "/admin/roles/" + strconv.FormatInt(roleId, 10)
		if isHTMX(r) {
			w.Header().Set("HX-Redirect", dest)
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, dest, http.StatusSeeOther)
	}
    return contracts.Route{
        Path:               "POST /admin/roles/{id}/edit",
        RequiresPermission: PermSystemAdmin,
        Func:               handler,
    }
}
