package ubadminpanel

import (
    "embed"
    "log/slog"
    "net/http"
    "strconv"
    "strings"
    "time"

    "github.com/kernelplex/ubase/lib/ensure"
    "github.com/kernelplex/ubase/lib/ubadminpanel/templ/views"
    "github.com/kernelplex/ubase/lib/ubapp"
    "github.com/kernelplex/ubase/lib/ubmanage"
    "github.com/kernelplex/ubase/lib/ubdata"
    "github.com/kernelplex/ubase/lib/ubstatus"
    "github.com/kernelplex/ubase/lib/ubwww"
)

//go:embed static
var static embed.FS

func isHTMX(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("HX-Request"), "true")
}

// hello route is a simple placeholder home page
func HelloRoute() ubwww.Route {
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

// OrganizationsRoute shows a searchable list of organizations.
func OrganizationsRoute(mgmt ubmanage.ManagementService) ubwww.Route {
    handler := func(w http.ResponseWriter, r *http.Request) {
        q := strings.TrimSpace(r.URL.Query().Get("q"))
        resp, err := mgmt.OrganizationList(r.Context())
        if err != nil {
            slog.Error("organization list error", "error", err)
            http.Error(w, "Failed to load organizations", http.StatusInternalServerError)
            return
        }
        orgs := resp.Data
        // filter by name or system name (case-insensitive)
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
        _ = views.OrganizationsPage(false, orgs, q).Render(r.Context(), w)
    }

    return ubwww.Route{
        Path:               "/admin/organizations",
        RequiresPermission: PermSystemAdmin,
        Func:               handler,
    }
}

// OrganizationOverviewRoute shows a single organization's overview by ID.
func OrganizationOverviewRoute(mgmt ubmanage.ManagementService) ubwww.Route {
    handler := func(w http.ResponseWriter, r *http.Request) {
        // Expect path like /admin/organizations/{id}
        prefix := "/admin/organizations/"
        if !strings.HasPrefix(r.URL.Path, prefix) {
            http.NotFound(w, r)
            return
        }
        idStr := strings.TrimPrefix(r.URL.Path, prefix)
        if idStr == "" {
            http.NotFound(w, r)
            return
        }
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
        _ = views.OrganizationOverview(false, id, name, systemName, roles).Render(r.Context(), w)
    }

    return ubwww.Route{
        Path:               "/admin/organizations/",
        RequiresPermission: PermSystemAdmin,
        Func:               handler,
    }
}

// OrganizationCreateRoute renders and processes the add organization form.
func OrganizationCreateRoute(mgmt ubmanage.ManagementService) ubwww.Route {
    handler := func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodGet {
            _ = views.OrganizationForm(isHTMX(r), false, nil, "", nil).Render(r.Context(), w)
            return
        }
        if r.Method == http.MethodPost {
            if err := r.ParseForm(); err != nil {
                _ = views.OrganizationForm(isHTMX(r), false, nil, "Invalid form submission", nil).Render(r.Context(), w)
                return
            }
            name := strings.TrimSpace(r.FormValue("name"))
            sys := strings.TrimSpace(r.FormValue("system_name"))
            status := strings.TrimSpace(r.FormValue("status"))
            resp, err := mgmt.OrganizationAdd(r.Context(), ubmanage.OrganizationCreateCommand{Name: name, SystemName: sys, Status: status}, "web:ubadminpanel")
            if err != nil || resp.Status != ubstatus.Success {
                if err != nil { slog.Error("org add error", "error", err) }
                errMap := resp.GetValidationMap()
                msg := resp.Message
                // Preserve entered values
                draft := ubdata.Organization{Name: name, SystemName: sys, Status: status}
                _ = views.OrganizationForm(isHTMX(r), false, &draft, msg, errMap).Render(r.Context(), w)
                return
            }
            // redirect to overview
            dest := "/admin/organizations/" + strconv.FormatInt(resp.Data.Id, 10)
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
    return ubwww.Route{
        Path:               "/admin/organizations/new",
        RequiresPermission: PermSystemAdmin,
        Func:               handler,
    }
}

// OrganizationEditRoute renders and processes the edit organization form.
func OrganizationEditRoute(mgmt ubmanage.ManagementService) ubwww.Route {
    handler := func(w http.ResponseWriter, r *http.Request) {
        prefix := "/admin/organizations/edit/"
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
            oresp, err := mgmt.OrganizationGet(r.Context(), id)
            if err != nil || oresp.Status != ubstatus.Success {
                http.NotFound(w, r)
                return
            }
            // map to ubdata.Organization shape
            org := ubdata.Organization{ID: id, Name: oresp.Data.State.Name, SystemName: oresp.Data.State.SystemName, Status: oresp.Data.State.Status}
            _ = views.OrganizationForm(isHTMX(r), true, &org, "", nil).Render(r.Context(), w)
            return
        }

        if r.Method == http.MethodPost {
            if err := r.ParseForm(); err != nil {
                _ = views.OrganizationForm(isHTMX(r), true, nil, "Invalid form submission", nil).Render(r.Context(), w)
                return
            }
            name := strings.TrimSpace(r.FormValue("name"))
            sys := strings.TrimSpace(r.FormValue("system_name"))
            status := strings.TrimSpace(r.FormValue("status"))
            // Set all fields for now
            cmd := ubmanage.OrganizationUpdateCommand{Id: id, Name: &name, SystemName: &sys, Status: &status}
            uresp, err := mgmt.OrganizationUpdate(r.Context(), cmd, "web:ubadminpanel")
            if err != nil || uresp.Status != ubstatus.Success {
                errMap := uresp.GetValidationMap()
                msg := uresp.Message
                draft := ubdata.Organization{ID: id, Name: name, SystemName: sys, Status: status}
                _ = views.OrganizationForm(isHTMX(r), true, &draft, msg, errMap).Render(r.Context(), w)
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
    return ubwww.Route{
        Path:               "/admin/organizations/edit/",
        RequiresPermission: PermSystemAdmin,
        Func:               handler,
    }
}
// RoleOverviewRoute shows a single role's overview by ID.
func RoleOverviewRoute(app *ubapp.UbaseApp, mgmt ubmanage.ManagementService) ubwww.Route {
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
        adapter := app.GetDBAdapter()
        // Users fragment endpoint
        if more == "users" || more == "users/add" || more == "users/remove" {
            // POST add/remove membership
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

    return ubwww.Route{
        Path:               "/admin/roles/",
        RequiresPermission: PermSystemAdmin,
        Func:               handler,
    }
}

// RoleCreateRoute renders and processes the create role form.
func RoleCreateRoute(mgmt ubmanage.ManagementService) ubwww.Route {
    handler := func(w http.ResponseWriter, r *http.Request) {
        // Preload orgs for selector
        orgsResp, _ := mgmt.OrganizationList(r.Context())
        orgs := []ubdata.Organization{}
        if orgsResp.Status == ubstatus.Success { orgs = orgsResp.Data }
        selectedOrg := int64(0)
        if v := strings.TrimSpace(r.URL.Query().Get("org")); v != "" { if oid, err := strconv.ParseInt(v, 10, 64); err == nil { selectedOrg = oid } }
        if len(orgs) > 0 && selectedOrg == 0 { selectedOrg = orgs[0].ID }
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
                if err != nil { slog.Error("role add error", "error", err) }
                errMap := resp.GetValidationMap()
                msg := resp.Message
                draft := ubdata.RoleRow{Name: name, SystemName: sys}
                _ = views.RoleForm(isHTMX(r), false, &draft, orgs, oid, msg, errMap).Render(r.Context(), w)
                return
            }
            dest := "/admin/roles/" + strconv.FormatInt(resp.Data.Id, 10)
            if isHTMX(r) { w.Header().Set("HX-Redirect", dest); w.WriteHeader(http.StatusOK); return }
            http.Redirect(w, r, dest, http.StatusSeeOther)
            return
        default:
            http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
        }
    }
    return ubwww.Route{ Path: "/admin/roles/new", RequiresPermission: PermSystemAdmin, Func: handler }
}

// RoleEditRoute renders and processes the edit role form.
func RoleEditRoute(mgmt ubmanage.ManagementService) ubwww.Route {
    handler := func(w http.ResponseWriter, r *http.Request) {
        prefix := "/admin/roles/edit/"
        if !strings.HasPrefix(r.URL.Path, prefix) { http.NotFound(w, r); return }
        idStr := strings.TrimPrefix(r.URL.Path, prefix)
        roleId, _ := strconv.ParseInt(idStr, 10, 64)
        if roleId <= 0 { http.NotFound(w, r); return }
        // Load orgs
        orgsResp, _ := mgmt.OrganizationList(r.Context())
        orgs := []ubdata.Organization{}
        if orgsResp.Status == ubstatus.Success { orgs = orgsResp.Data }
        switch r.Method {
        case http.MethodGet:
            rresp, err := mgmt.RoleGetById(r.Context(), roleId)
            if err != nil || rresp.Status != ubstatus.Success { http.NotFound(w, r); return }
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
                errMap := resp.GetValidationMap(); msg := resp.Message
                draft := ubdata.RoleRow{ID: roleId, Name: name, SystemName: sys}
                // Try to get existing org id for selection
                rresp, _ := mgmt.RoleGetById(r.Context(), roleId)
                selected := int64(0)
                if rresp.Status == ubstatus.Success { selected = rresp.Data.State.OrganizationId }
                _ = views.RoleForm(isHTMX(r), true, &draft, orgs, selected, msg, errMap).Render(r.Context(), w)
                return
            }
            dest := "/admin/roles/" + strconv.FormatInt(roleId, 10)
            if isHTMX(r) { w.Header().Set("HX-Redirect", dest); w.WriteHeader(http.StatusOK); return }
            http.Redirect(w, r, dest, http.StatusSeeOther)
            return
        default:
            http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
        }
    }
    return ubwww.Route{ Path: "/admin/roles/edit/", RequiresPermission: PermSystemAdmin, Func: handler }
}
// LoginRoute handles GET (render form) and POST (authenticate).
func LoginRoute(
	app *ubapp.UbaseApp,
	mgmt ubmanage.ManagementService,
	cookieManager ubwww.AuthTokenCookieManager[*ubwww.AuthToken],
) ubwww.Route {
	ensure.That(app != nil, "app cannot be nil")
	primaryOrganization := app.GetConfig().PrimaryOrganization

	return ubwww.Route{
		Path: "/admin/login",
		Func: func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				component := views.Login(isHTMX(r), "")
				_ = component.Render(r.Context(), w)
				return
			case http.MethodPost:
				if err := r.ParseForm(); err != nil {
					slog.Error("parse form error", "error", err)
					_ = views.Login(isHTMX(r), "Invalid form submission").Render(r.Context(), w)
					return
				}
				email := strings.TrimSpace(r.FormValue("email"))
				password := r.FormValue("password")

				resp, err := mgmt.UserAuthenticate(r.Context(), ubmanage.UserLoginCommand{Email: email, Password: password}, "web:ubadminpanel")
				if err != nil {
					slog.Error("auth error", "error", err)
					_ = views.Login(isHTMX(r), "Could not verify this account at this time.").Render(r.Context(), w)
					return
				}

				switch resp.Status {
				case ubstatus.Success:

					now := time.Now().Unix()
					token := &ubwww.AuthToken{
						UserId:               resp.Data.UserId,
						OrganizationId:       primaryOrganization,
						Email:                resp.Data.Email,
						TwoFactorRequired:    false,
						RequiresVerification: resp.Data.RequiresVerification,
						SoftExpiry:           now + 3600,
						HardExpiry:           now + 86400,
					}
					if err := cookieManager.WriteAuthTokenCookie(w, token); err != nil {
						slog.Error("write cookie error", "error", err)
						_ = views.Login(isHTMX(r), "Failed to create session. Try again.").Render(r.Context(), w)
						return
					}
					if isHTMX(r) {
						w.Header().Set("HX-Redirect", "/admin")
						w.WriteHeader(http.StatusOK)
						return
					}
					http.Redirect(w, r, "/admin", http.StatusSeeOther)
					return
				case ubstatus.PartialSuccess:
					if resp.Data.RequiresTwoFactor {
						_ = views.TwoFactor(isHTMX(r), resp.Data.UserId, "").Render(r.Context(), w)
						return
					}
					_ = views.Login(isHTMX(r), "Please verify your email before logging in.").Render(r.Context(), w)
					return
				default:
					msg := resp.Message
					if strings.TrimSpace(msg) == "" {
						msg = "Email or password is incorrect"
					}
					_ = views.Login(isHTMX(r), msg).Render(r.Context(), w)
					return
				}
			default:
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}
		},
	}
}

// VerifyTwoFactorRoute handles POST verification of 2FA code.
func VerifyTwoFactorRoute(
	mgmt ubmanage.ManagementService,
	cookieManager ubwww.AuthTokenCookieManager[*ubwww.AuthToken],
) ubwww.Route {
	return ubwww.Route{
		Path: "/admin/verify-2fa",
		Func: func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}
			if err := r.ParseForm(); err != nil {
				_ = views.TwoFactor(isHTMX(r), 0, "Invalid form submission").Render(r.Context(), w)
				return
			}
			idStr := r.FormValue("user_id")
			code := strings.TrimSpace(r.FormValue("code"))
			userId, _ := strconv.ParseInt(idStr, 10, 64)

			verifyResp, err := mgmt.UserVerifyTwoFactorCode(r.Context(), ubmanage.UserVerifyTwoFactorLoginCommand{UserId: userId, Code: code}, "web:ubadminpanel")
			if err != nil || verifyResp.Status != ubstatus.Success {
				msg := "Two factor code does not match"
				if err != nil {
					slog.Error("2fa error", "error", err)
				}
				_ = views.TwoFactor(isHTMX(r), userId, msg).Render(r.Context(), w)
				return
			}

			now := time.Now().Unix()
			token := &ubwww.AuthToken{
				UserId:               userId,
				OrganizationId:       0,
				Email:                "",
				TwoFactorRequired:    false,
				RequiresVerification: false,
				SoftExpiry:           now + 3600,
				HardExpiry:           now + 86400,
			}
			if err := cookieManager.WriteAuthTokenCookie(w, token); err != nil {
				slog.Error("write cookie error", "error", err)
				_ = views.TwoFactor(isHTMX(r), userId, "Failed to create session. Try again.").Render(r.Context(), w)
				return
			}
			if isHTMX(r) {
				w.Header().Set("HX-Redirect", "/admin")
				w.WriteHeader(http.StatusOK)
				return
			}
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
		},
	}
}

// LogoutRoute clears the auth cookie.
func LogoutRoute(cookieManager ubwww.AuthTokenCookieManager[*ubwww.AuthToken]) ubwww.Route {
	return ubwww.Route{
		Path: "/admin/logout",
		Func: func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}
			cookieManager.ClearAuthTokenCookie(w)
			if isHTMX(r) {
				w.Header().Set("HX-Redirect", "/admin/login")
				w.WriteHeader(http.StatusOK)
				return
			}
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		},
	}
}

// RegisterAdminPanelRoutes adds admin routes and static files.
// cookieManager is used to read/write the auth cookie; mgmt performs authentication.
func RegisterAdminPanelRoutes(
    app *ubapp.UbaseApp,
    web ubwww.WebService,
    mgmt ubmanage.ManagementService,
    cookieManager ubwww.AuthTokenCookieManager[*ubwww.AuthToken],
) {

	ensure.That(app != nil, "app cannot be nil")

	// Serve static files
	fs := http.FileServer(http.FS(static))
	web.AddRouteHandler("/admin/static/", http.StripPrefix("/admin", fs))

	// Home placeholder (can be permission-protected later)
    web.AddRoute(HelloRoute())
    web.AddRoute(OrganizationsRoute(mgmt))
    web.AddRoute(OrganizationOverviewRoute(mgmt))
    web.AddRoute(OrganizationCreateRoute(mgmt))
    web.AddRoute(OrganizationEditRoute(mgmt))
    web.AddRoute(RoleOverviewRoute(app, mgmt))
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

// UsersRoute shows a searchable list of users.
func UsersRoute(app *ubapp.UbaseApp) ubwww.Route {
    handler := func(w http.ResponseWriter, r *http.Request) {
        adapter := app.GetDBAdapter()
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
    return ubwww.Route{
        Path:               "/admin/users",
        RequiresPermission: PermSystemAdmin,
        Func:               handler,
    }
}

// UserCreateRoute renders and processes the create user form.
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
                if err != nil { slog.Error("user add error", "error", err) }
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
    return ubwww.Route{
        Path:               "/admin/users/new",
        RequiresPermission: PermSystemAdmin,
        Func:               handler,
    }
}

// UserEditRoute renders and processes the edit user form.
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
            if password != "" { cmd.Password = &password }
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
    return ubwww.Route{
        Path:               "/admin/users/edit/",
        RequiresPermission: PermSystemAdmin,
        Func:               handler,
    }
}
// UserOverviewRoute shows a single user's details by ID.
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
        // Roles fragment and toggle endpoints
        if more == "roles" || more == "roles/add" || more == "roles/remove" {
            // Parse org id from query or form
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
            // Add/remove membership
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
                // Determine membership and return the updated row
                hasResp, _ := mgmt.UserGetOrganizationRoles(r.Context(), id, orgId)
                memberSet := map[int64]bool{}
                if hasResp.Status == ubstatus.Success {
                    for _, rr := range hasResp.Data {
                        memberSet[rr.ID] = true
                    }
                }
                // Build role row for re-render
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
            // Full table for selected org
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
        // Organizations for picker
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
        if len(orgs) > 0 {
            if selectedOrg == 0 {
                selectedOrg = orgs[0].ID
            }
        }
        _ = views.UserOverview(false, id, st.DisplayName, st.Email, st.FirstName, st.LastName, st.Verified, st.Disabled, orgs, selectedOrg).Render(r.Context(), w)
    }
    return ubwww.Route{
        Path:               "/admin/users/",
        RequiresPermission: PermSystemAdmin,
        Func:               handler,
    }
}
