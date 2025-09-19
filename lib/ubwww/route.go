package ubwww

import "net/http"

type Route struct {
	Path                  string
	RequiresPermission    string
	RequiresAuthenticated bool
	Api                   bool
	Func                  http.HandlerFunc
}
