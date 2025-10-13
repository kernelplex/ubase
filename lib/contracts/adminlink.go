package contracts

import (
	"iter"
	"net/http"

	"github.com/a-h/templ"
)

type AdminLink struct {
	Title              string
	Section            string
	Icon               string
	Path               string
	HtmxAware          bool
	RequiredPermission string // Optional permission required to view this link
}

type AdminSectionLinks struct {
	sections []string
	links    map[string][]AdminLink
}

// type AdminSectionLinks map[string][]AdminLink

func (asl *AdminSectionLinks) Add(link AdminLink) {
	if asl.links == nil {
		asl.links = make(map[string][]AdminLink)
	}
	// Is the section already in links?
	if _, ok := asl.links[link.Section]; !ok {
		asl.sections = append(asl.sections, link.Section)
		asl.links[link.Section] = []AdminLink{}
	}
	asl.links[link.Section] = append(asl.links[link.Section], link)
}

// Iterate the sections in the order they were added
func (asl *AdminSectionLinks) Iter() iter.Seq2[string, []AdminLink] {
	return func(yield func(string, []AdminLink) bool) {
		for _, s := range asl.sections {
			if !yield(s, asl.links[s]) {
				return
			}
		}
	}
}

type AdminLinkService interface {
	GetLinks(r *http.Request) *AdminSectionLinks
	AddLink(link AdminLink)
}

type AdminRenderer interface {
	Render(w http.ResponseWriter, r *http.Request, component templ.Component)
	AddStyle(css string)
}
