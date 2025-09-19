package ubwww

import (
	"net/http"
	"strings"

	"github.com/kernelplex/ubase/lib/ubmanage"
)

type PermissionMiddleware struct {
	prefectService ubmanage.PrefectService
	cookieManager  AuthTokenCookieManager[*AuthToken]
}

func NewPermissionMiddleware(
	prefectService ubmanage.PrefectService,
	cookieManager AuthTokenCookieManager[*AuthToken],

) *PermissionMiddleware {
	return &PermissionMiddleware{
		prefectService: prefectService,
		cookieManager:  cookieManager,
	}
}

func (pm *PermissionMiddleware) RequirePermission(permission string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, found := pm.cookieManager.IdentityFromContext(r.Context())
		if !found {
			// If unauthenticated, redirect to login. Support HTMX fragments.
			if strings.EqualFold(r.Header.Get("HX-Request"), "true") {
				w.Header().Set("HX-Redirect", "/admin/login")
				w.WriteHeader(http.StatusOK)
			} else {
				http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			}
			return
		}

		// Check if the user has the required permission
		hasPermission, err := pm.prefectService.UserHasPermission(r.Context(), identity.UserID, identity.OrganizationID, permission)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if !hasPermission {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Call the next handler if permission is granted
		next.ServeHTTP(w, r)
	}
}
