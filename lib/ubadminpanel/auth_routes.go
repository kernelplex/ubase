package ubadminpanel

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kernelplex/ubase/lib/ensure"
	"github.com/kernelplex/ubase/lib/ubadminpanel/templ/views"
	"github.com/kernelplex/ubase/lib/ubmanage"
	"github.com/kernelplex/ubase/lib/ubstatus"
	"github.com/kernelplex/ubase/lib/ubwww"
)

// LoginRoute handles GET (render form) and POST (authenticate).
func LoginRoute(
	primaryOrganization int64,
	mgmt ubmanage.ManagementService,
	cookieManager ubwww.AuthTokenCookieManager[*ubwww.AuthToken],
) ubwww.Route {
	ensure.That(primaryOrganization > 0, "primary organization must be set and greater than zero")

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
