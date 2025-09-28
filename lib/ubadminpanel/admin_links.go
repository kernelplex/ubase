package ubadminpanel

import (
	"net/http"

	"github.com/kernelplex/ubase/lib/contracts"
	"github.com/kernelplex/ubase/lib/ubmanage"
)

type AdminLinkServiceImpl struct {
	links          []contracts.AdminLink
	prefectService ubmanage.PrefectService
	cookieManager  contracts.AuthTokenCookieManager
}

func NewAdminLinkService(
	prefectService ubmanage.PrefectService,
	cookieManager contracts.AuthTokenCookieManager) contracts.AdminLinkService {
	return &AdminLinkServiceImpl{
		links:          getAdminLinks(),
		prefectService: prefectService,
		cookieManager:  cookieManager,
	}
}

func (als *AdminLinkServiceImpl) GetLinks(r *http.Request) []contracts.AdminLink {
	identity, found := als.cookieManager.IdentityFromContext(r.Context())

	linksToShow := make([]contracts.AdminLink, 0, len(als.links))
	for _, link := range als.links {
		if link.RequiredPermission == "" {
			linksToShow = append(linksToShow, link)
			continue
		}
		if !found || identity.UserID == 0 || identity.OrganizationID == 0 {
			continue
		}
		hasPermission, err := als.prefectService.UserHasPermission(r.Context(), identity.UserID, identity.OrganizationID, link.RequiredPermission)
		if err == nil && hasPermission {
			linksToShow = append(linksToShow, link)
		}
	}
	return linksToShow
}

func (als *AdminLinkServiceImpl) AddLink(link contracts.AdminLink) {
	als.links = append(als.links, link)
}

// GetAdminLinks returns the standard admin navigation links
func getAdminLinks() []contracts.AdminLink {
	return []contracts.AdminLink{
		{Title: "Dashboard", Icon: "home", Path: "/admin/", HtmxAware: true},
		{Title: "Organizations", Icon: "building", Path: "/admin/organizations", HtmxAware: true},
		{Title: "Users", Icon: "users", Path: "/admin/users", HtmxAware: true},
		{Title: "Roles", Icon: "key", Path: "/admin/roles", HtmxAware: true},
	}
}
