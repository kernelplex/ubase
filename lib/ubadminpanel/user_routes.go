package ubadminpanel

import (
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

func UsersRoute(adapter ubdata.DataAdapter) ubwww.Route {

	handler := func(w http.ResponseWriter, r *http.Request) {
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		const limit = 25
		users, err := adapter.SearchUsers(r.Context(), q, limit, 0)
		if err != nil {
			slog.Error("user search error", "error", err)
			http.Error(w, "Failed to load users", http.StatusInternalServerError)
			return
		}
		if isHTMX(r) {
			_ = views.UsersTable(users).Render(r.Context(), w)
			return
		}
		_ = views.UsersPage(false, users, q).Render(r.Context(), w)
	}
	return ubwww.Route{Path: "/admin/users", RequiresPermission: PermSystemAdmin, Func: handler}
}

func UserOverviewRoute(mgmt ubmanage.ManagementService) ubwww.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		prefix := "/admin/users/"
		if !strings.HasPrefix(r.URL.Path, prefix) {
			http.NotFound(w, r)
			return
		}
		rest := strings.TrimPrefix(r.URL.Path, prefix)
		if rest == "" {
			http.NotFound(w, r)
			return
		}
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
		if more == "roles" || more == "roles/add" || more == "roles/remove" {
			orgStr := r.URL.Query().Get("org")
			if orgStr == "" && r.Method == http.MethodPost {
				_ = r.ParseForm()
				orgStr = r.FormValue("org")
			}
			orgId, _ := strconv.ParseInt(orgStr, 10, 64)
			if orgId <= 0 {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			if r.Method == http.MethodPost {
				_ = r.ParseForm()
				roleStr := r.FormValue("role_id")
				roleId, _ := strconv.ParseInt(roleStr, 10, 64)
				if roleId <= 0 {
					http.Error(w, "Bad Request", http.StatusBadRequest)
					return
				}
				if more == "roles/add" {
					_, _ = mgmt.UserAddToRole(r.Context(), ubmanage.UserAddToRoleCommand{UserId: id, RoleId: roleId}, "web:ubadminpanel")
				} else if more == "roles/remove" {
					_, _ = mgmt.UserRemoveFromRole(r.Context(), ubmanage.UserRemoveFromRoleCommand{UserId: id, RoleId: roleId}, "web:ubadminpanel")
				}
				hasResp, _ := mgmt.UserGetOrganizationRoles(r.Context(), id, orgId)
				memberSet := map[int64]bool{}
				if hasResp.Status == ubstatus.Success {
					for _, rr := range hasResp.Data {
						memberSet[rr.ID] = true
					}
				}
				rolesResp, _ := mgmt.RoleList(r.Context(), orgId)
				var role ubdata.RoleRow
				if rolesResp.Status == ubstatus.Success {
					for _, rr := range rolesResp.Data {
						if rr.ID == roleId {
							role = rr
							break
						}
					}
				}
				_ = views.UserRoleRow(id, role, memberSet[roleId], orgId).Render(r.Context(), w)
				return
			}
			rolesResp, err := mgmt.RoleList(r.Context(), orgId)
			if err != nil || rolesResp.Status != ubstatus.Success {
				http.Error(w, "Failed to load roles", http.StatusInternalServerError)
				return
			}
			hasResp, _ := mgmt.UserGetOrganizationRoles(r.Context(), id, orgId)
			memberSet := map[int64]bool{}
			if hasResp.Status == ubstatus.Success {
				for _, rr := range hasResp.Data {
					memberSet[rr.ID] = true
				}
			}
			_ = views.UserRolesTable(id, rolesResp.Data, memberSet, orgId).Render(r.Context(), w)
			return
		}
		resp, err := mgmt.UserGetById(r.Context(), id)
		if err != nil || resp.Status != ubstatus.Success {
			slog.Error("user get error", "error", err, "id", id, "status", resp.Status)
			http.NotFound(w, r)
			return
		}
		st := resp.Data.State
		orgsResp, oerr := mgmt.OrganizationList(r.Context())
		orgs := []ubdata.Organization{}
		if oerr == nil && orgsResp.Status == ubstatus.Success {
			orgs = orgsResp.Data
		}
		var selectedOrg int64
		if v := strings.TrimSpace(r.URL.Query().Get("org")); v != "" {
			if oid, err := strconv.ParseInt(v, 10, 64); err == nil {
				selectedOrg = oid
			}
		}
		if len(orgs) > 0 && selectedOrg == 0 {
			selectedOrg = orgs[0].ID
		}
		_ = views.UserOverview(
			false,
			id,
			st.DisplayName,
			st.Email,
			st.FirstName,
			st.LastName,
			st.Verified,
			st.Disabled,
			st.LastLogin,
			st.LoginCount,
			st.LastLoginAttempt,
			st.FailedLoginAttempts,
			orgs,
			selectedOrg,
		).Render(r.Context(), w)
	}
	return ubwww.Route{Path: "/admin/users/", RequiresPermission: PermSystemAdmin, Func: handler}
}

func UserCreateRoute(mgmt ubmanage.ManagementService) ubwww.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_ = views.UserForm(isHTMX(r), false, nil, "", nil).Render(r.Context(), w)
			return
		}
		if r.Method == http.MethodPost {
			if err := r.ParseForm(); err != nil {
				_ = views.UserForm(isHTMX(r), false, nil, "Invalid form submission", nil).Render(r.Context(), w)
				return
			}
			email := strings.TrimSpace(r.FormValue("email"))
			password := r.FormValue("password")
			first := strings.TrimSpace(r.FormValue("first_name"))
			last := strings.TrimSpace(r.FormValue("last_name"))
			display := strings.TrimSpace(r.FormValue("display_name"))
			verified := r.FormValue("verified") == "on"
			cmd := ubmanage.UserCreateCommand{Email: email, Password: password, FirstName: first, LastName: last, DisplayName: display, Verified: verified}
			resp, err := mgmt.UserAdd(r.Context(), cmd, "web:ubadminpanel")
			if err != nil || resp.Status != ubstatus.Success {
				if err != nil {
					slog.Error("user add error", "error", err)
				}
				errMap := resp.GetValidationMap()
				msg := resp.Message
				draft := ubdata.User{Email: email, FirstName: first, LastName: last, DisplayName: display, Verified: verified}
				_ = views.UserForm(isHTMX(r), false, &draft, msg, errMap).Render(r.Context(), w)
				return
			}
			dest := "/admin/users/" + strconv.FormatInt(resp.Data.Id, 10)
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
	return ubwww.Route{Path: "/admin/users/new", RequiresPermission: PermSystemAdmin, Func: handler}
}

func UserEditRoute(mgmt ubmanage.ManagementService) ubwww.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		prefix := "/admin/users/edit/"
		if !strings.HasPrefix(r.URL.Path, prefix) {
			http.NotFound(w, r)
			return
		}
		idStr := strings.TrimPrefix(r.URL.Path, prefix)
		id, _ := strconv.ParseInt(idStr, 10, 64)
		if id <= 0 {
			http.NotFound(w, r)
			return
		}
		if r.Method == http.MethodGet {
			uresp, err := mgmt.UserGetById(r.Context(), id)
			if err != nil || uresp.Status != ubstatus.Success {
				http.NotFound(w, r)
				return
			}
			st := uresp.Data.State
			draft := ubdata.User{UserID: id, Email: st.Email, FirstName: st.FirstName, LastName: st.LastName, DisplayName: st.DisplayName, Verified: st.Verified}
			_ = views.UserForm(isHTMX(r), true, &draft, "", nil).Render(r.Context(), w)
			return
		}
		if r.Method == http.MethodPost {
			if err := r.ParseForm(); err != nil {
				_ = views.UserForm(isHTMX(r), true, nil, "Invalid form submission", nil).Render(r.Context(), w)
				return
			}
			email := strings.TrimSpace(r.FormValue("email"))
			password := strings.TrimSpace(r.FormValue("password"))
			first := strings.TrimSpace(r.FormValue("first_name"))
			last := strings.TrimSpace(r.FormValue("last_name"))
			display := strings.TrimSpace(r.FormValue("display_name"))
			verified := r.FormValue("verified") == "on"
			cmd := ubmanage.UserUpdateCommand{Id: id}
			cmd.Email = &email
			if password != "" {
				cmd.Password = &password
			}
			cmd.FirstName = &first
			cmd.LastName = &last
			cmd.DisplayName = &display
			cmd.Verified = &verified
			resp, err := mgmt.UserUpdate(r.Context(), cmd, "web:ubadminpanel")
			if err != nil || resp.Status != ubstatus.Success {
				errMap := resp.GetValidationMap()
				msg := resp.Message
				draft := ubdata.User{UserID: id, Email: email, FirstName: first, LastName: last, DisplayName: display, Verified: verified}
				_ = views.UserForm(isHTMX(r), true, &draft, msg, errMap).Render(r.Context(), w)
				return
			}
			dest := "/admin/users/" + strconv.FormatInt(id, 10)
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
	return ubwww.Route{Path: "/admin/users/edit/", RequiresPermission: PermSystemAdmin, Func: handler}
}
