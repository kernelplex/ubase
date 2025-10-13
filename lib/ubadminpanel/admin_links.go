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

func (als *AdminLinkServiceImpl) GetLinks(r *http.Request) *contracts.AdminSectionLinks {
	identity, found := als.cookieManager.IdentityFromContext(r.Context())

	linksToShow := contracts.AdminSectionLinks{}
	for _, link := range als.links {
		if link.RequiredPermission == "" {
			linksToShow.Add(link)
			continue
		}
		if !found || identity.UserID == 0 || identity.OrganizationID == 0 {
			continue
		}
		hasPermission, err := als.prefectService.UserHasPermission(r.Context(), identity.UserID, identity.OrganizationID, link.RequiredPermission)
		if err == nil && hasPermission {
			linksToShow.Add(link)
		}
	}
	return &linksToShow
}

func (als *AdminLinkServiceImpl) AddLink(link contracts.AdminLink) {
	als.links = append(als.links, link)
}

// GetAdminLinks returns the standard admin navigation links
func getAdminLinks() []contracts.AdminLink {
	return []contracts.AdminLink{
		{Title: "Dashboard", Icon: "home", Path: "/admin/", HtmxAware: true, RequiredPermission: PermSystemAdmin, Section: "General"},
		{Title: "Organizations", Icon: "building", Path: "/admin/organizations", HtmxAware: true, RequiredPermission: PermSystemAdmin, Section: "System"},
		{Title: "Users", Icon: "users", Path: "/admin/users", HtmxAware: true, RequiredPermission: PermSystemAdmin, Section: "System"},
	}
}
