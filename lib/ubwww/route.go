package ubwww

import "net/http"

type Route struct {
	Path               string
	RequiresPermission string
	Api                bool
	Func               http.HandlerFunc
}
