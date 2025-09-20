package ubadminpanel

import (
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/kernelplex/ubase/lib/ubadminpanel/templ/views"
	"github.com/kernelplex/ubase/lib/ubdata"
	"github.com/kernelplex/ubase/lib/ubmanage"
	"github.com/kernelplex/ubase/lib/ubstatus"
	"github.com/kernelplex/ubase/lib/ubwww"
)

func RoleOverviewRoute(adapter ubdata.DataAdapter, mgmt ubmanage.ManagementService, permissions []string) ubwww.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		prefix := "/admin/roles/"
		if !strings.HasPrefix(r.URL.Path, prefix) {
			http.NotFound(w, r)
			return
		}
		rest := strings.TrimPrefix(r.URL.Path, prefix)
		if rest == "" {
			http.NotFound(w, r)
			return
		}
		// Accept both /{id} and /{id}/users
		idPart := rest
		more := ""
		if i := strings.IndexByte(rest, '/'); i >= 0 {
			idPart = rest[:i]
			more = rest[i+1:]
		}
		id, err := strconv.ParseInt(idPart, 10, 64)
		if err != nil || id <= 0 {
			http.NotFound(w, r)
			return
		}
		if more == "users" || more == "users/add" || more == "users/remove" {
			if r.Method == http.MethodPost {
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
				if more == "users/add" {
					// add user to role
					_, _ = mgmt.UserAddToRole(r.Context(), ubmanage.UserAddToRoleCommand{UserId: uid, RoleId: id}, "web:ubadminpanel")
				} else if more == "users/remove" {
					_, _ = mgmt.UserRemoveFromRole(r.Context(), ubmanage.UserRemoveFromRoleCommand{UserId: uid, RoleId: id}, "web:ubadminpanel")
				}
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
			if r.Method == http.MethodPost {
				uidStr := r.FormValue("user_id")
				uid, _ := strconv.ParseInt(uidStr, 10, 64)
				user, gerr := adapter.GetUser(r.Context(), uid)
				if gerr != nil {
					slog.Error("get user error", "error", gerr, "user", uid)
					http.Error(w, "Not Found", http.StatusNotFound)
					return
				}
				inRole := memberSet[uid]
				_ = views.RoleUserRow(id, user, inRole).Render(r.Context(), w)
				return
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
			return
		}

		// Permissions endpoints
		if more == "permissions" || more == "permissions/add" || more == "permissions/remove" {
			// helper to render a single permission row
			writePermRow := func(w http.ResponseWriter, roleId int64, perm string, inRole bool) {
				// row class if in role
				rowClass := ""
				if inRole {
					rowClass = " class=\"row-in-role\""
				}
				// Start row
				_, _ = w.Write([]byte("<tr" + rowClass + ">"))
				// Permission cell
				_, _ = w.Write([]byte("<td>" + html.EscapeString(perm) + "</td>"))
				// Toggle cell
				_, _ = w.Write([]byte("<td>"))
				if inRole {
					// remove form
					_, _ = w.Write([]byte(
						fmt.Sprintf("<form hx-post=\"/admin/roles/%d/permissions/remove\" hx-include='[name=\"q\"]' hx-target=\"closest tr\" hx-swap=\"outerHTML\" style=\"display:inline\">", roleId),
					))
					_, _ = w.Write([]byte("<input type=\"hidden\" name=\"permission\" value=\"" + html.EscapeString(perm) + "\"/>"))
					_, _ = w.Write([]byte("<button type=\"submit\" class=\"role-toggle minus\" title=\"Remove permission\">-</button>"))
					_, _ = w.Write([]byte("</form>"))
				} else {
					// add form
					_, _ = w.Write([]byte(
						fmt.Sprintf("<form hx-post=\"/admin/roles/%d/permissions/add\" hx-include='[name=\"q\"]' hx-target=\"closest tr\" hx-swap=\"outerHTML\" style=\"display:inline\">", roleId),
					))
					_, _ = w.Write([]byte("<input type=\"hidden\" name=\"permission\" value=\"" + html.EscapeString(perm) + "\"/>"))
					_, _ = w.Write([]byte("<button type=\"submit\" class=\"role-toggle plus\" title=\"Add permission\">+</button>"))
					_, _ = w.Write([]byte("</form>"))
				}
				_, _ = w.Write([]byte("</td>"))
				// End row
				_, _ = w.Write([]byte("</tr>"))
			}

			// Process add/remove actions
			if r.Method == http.MethodPost {
				if err := r.ParseForm(); err != nil {
					http.Error(w, "Bad Request", http.StatusBadRequest)
					return
				}
				perm := strings.TrimSpace(r.FormValue("permission"))
				if perm == "" {
					http.Error(w, "Bad Request", http.StatusBadRequest)
					return
				}
				if more == "permissions/add" {
					_, _ = mgmt.RolePermissionAdd(r.Context(), ubmanage.RolePermissionAddCommand{Id: id, Permission: perm}, "web:ubadminpanel")
				} else if more == "permissions/remove" {
					_, _ = mgmt.RolePermissionRemove(r.Context(), ubmanage.RolePermissionRemoveCommand{Id: id, Permission: perm}, "web:ubadminpanel")
				}

				// compute membership set after change to render the single row
				assigned, aerr := adapter.GetRolePermissions(r.Context(), id)
				if aerr != nil {
					slog.Error("get role permissions error", "error", aerr, "role", id)
				}
				memberSet := make(map[string]bool, len(assigned))
				for _, p := range assigned {
					memberSet[p] = true
				}
				writePermRow(w, id, perm, memberSet[perm])
				return
			}

			// GET: render permissions table (with optional search)
			assigned, aerr := adapter.GetRolePermissions(r.Context(), id)
			if aerr != nil {
				slog.Error("get role permissions error", "error", aerr, "role", id)
				assigned = []string{}
			}
			memberSet := make(map[string]bool, len(assigned))
			for _, p := range assigned {
				memberSet[p] = true
			}

			// filter by query
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

			// render table
			_, _ = w.Write([]byte("<div id=\"role-permissions\">"))
			_, _ = w.Write([]byte("<table class=\"data-table\">"))
			_, _ = w.Write([]byte("<thead><tr><th style=\"text-align: left;\">Permission</th><th style=\"width: 80px; text-align: left;\">In Role</th></tr></thead>"))
			_, _ = w.Write([]byte("<tbody>"))
			if len(filtered) == 0 {
				_, _ = w.Write([]byte("<tr><td colspan=\"2\" style=\"color: var(--text-muted); padding: 0.75rem 0;\">No permissions found.</td></tr>"))
			} else {
				for _, p := range filtered {
					writePermRow(w, id, p, memberSet[p])
				}
			}
			_, _ = w.Write([]byte("</tbody></table></div>"))
			return
		}
		// Page endpoint
		resp, err := mgmt.RoleGetById(r.Context(), id)
		if err != nil || resp.Status != ubstatus.Success {
			slog.Error("role get error", "error", err, "id", id, "status", resp.Status)
			http.NotFound(w, r)
			return
		}
		state := resp.Data.State
		_ = views.RoleOverview(false, id, state.Name, state.SystemName, state.OrganizationId).Render(r.Context(), w)
	}

	return ubwww.Route{Path: "/admin/roles/", RequiresPermission: PermSystemAdmin, Func: handler}
}

func RoleCreateRoute(mgmt ubmanage.ManagementService) ubwww.Route {
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
	return ubwww.Route{Path: "/admin/roles/new", RequiresPermission: PermSystemAdmin, Func: handler}
}

func RoleEditRoute(mgmt ubmanage.ManagementService) ubwww.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		prefix := "/admin/roles/edit/"
		if !strings.HasPrefix(r.URL.Path, prefix) {
			http.NotFound(w, r)
			return
		}
		idStr := strings.TrimPrefix(r.URL.Path, prefix)
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
	return ubwww.Route{Path: "/admin/roles/edit/", RequiresPermission: PermSystemAdmin, Func: handler}
}
