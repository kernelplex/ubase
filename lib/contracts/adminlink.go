package contracts

import "net/http"

type AdminLink struct {
	Title              string
	Icon               string
	Path               string
	HtmxAware          bool
	RequiredPermission string // Optional permission required to view this link
}

type AdminLinkService interface {
	GetLinks(r *http.Request) []AdminLink
	AddLink(link AdminLink)
}
