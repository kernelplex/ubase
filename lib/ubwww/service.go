package ubwww

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/kernelplex/ubase/lib/contracts"
	"github.com/kernelplex/ubase/lib/ensure"
	"github.com/kernelplex/ubase/lib/ubadminpanel"
	"github.com/kernelplex/ubase/lib/ubdata"
	"github.com/kernelplex/ubase/lib/ubmanage"
)

type WebService interface {
	AddRoute(route contracts.Route) WebService
	AddRouteHandler(path string, handler http.Handler) WebService
	AddAdminRoutes() WebService
	Start() error
	Stop() error
}

type WebServiceImpl struct {
	routes              []contracts.Route
	primaryOrganization int64
	adapter             ubdata.DataAdapter
	port                uint
	mux                 *http.ServeMux
	server              *http.Server
	cookieManager       contracts.AuthTokenCookieManager
	managementService   ubmanage.ManagementService
	permMiddleware      *PermissionMiddleware
	permissions         []string
}

func NewWebService(
	port uint,
	primaryOrganization int64,
	dataAdapter ubdata.DataAdapter,
	cookieManager contracts.AuthTokenCookieManager,
	managementService ubmanage.ManagementService,
	permMiddleware *PermissionMiddleware,
	permissions []string) WebService {
	mux := http.NewServeMux()
	return &WebServiceImpl{
		routes:              make([]contracts.Route, 0, 20),
		port:                port,
		mux:                 mux,
		adapter:             dataAdapter,
		cookieManager:       cookieManager,
		permMiddleware:      permMiddleware,
		managementService:   managementService,
		primaryOrganization: primaryOrganization,
		permissions:         permissions,
	}
}

func (ws *WebServiceImpl) AddAdminRoutes() WebService {
	ensure.That(ws.primaryOrganization > 0, "primary organization must be set and greater than zero")
	ensure.That(ws.adapter != nil, "data adapter is required")
	ensure.That(ws.managementService != nil, "management service is required")

	// Serve static files
	fs := http.FileServer(http.FS(ubadminpanel.Static))
	ws.AddRouteHandler("/admin/static/", http.StripPrefix("/admin", fs))

	// Home placeholder (can be permission-protected later)
	ws.AddRoute(ubadminpanel.AdminRoute(ws.adapter, ws.managementService))
	ws.AddRoute(ubadminpanel.OrganizationsRoute(ws.managementService))
	ws.AddRoute(ubadminpanel.OrganizationOverviewRoute(ws.managementService))
	ws.AddRoute(ubadminpanel.OrganizationCreateRoute(ws.managementService))
	ws.AddRoute(ubadminpanel.OrganizationCreatePostRoute(ws.managementService))
	ws.AddRoute(ubadminpanel.OrganizationEditRoute(ws.managementService))
	ws.AddRoute(ubadminpanel.RoleOverviewRoute(ws.adapter, ws.managementService, ws.permissions))
	ws.AddRoute(ubadminpanel.RoleUsersListRoute(ws.adapter))
	ws.AddRoute(ubadminpanel.RoleUsersAddRoute(ws.adapter, ws.managementService))
	ws.AddRoute(ubadminpanel.RoleUsersRemoveRoute(ws.adapter, ws.managementService))
	ws.AddRoute(ubadminpanel.RolePermissionsListRoute(ws.adapter, ws.permissions))
	ws.AddRoute(ubadminpanel.RolePermissionsAddRoute(ws.adapter, ws.managementService))
	ws.AddRoute(ubadminpanel.RolePermissionsRemoveRoute(ws.adapter, ws.managementService))
	ws.AddRoute(ubadminpanel.RoleCreateRoute(ws.managementService))
	ws.AddRoute(ubadminpanel.RoleCreatePostRoute(ws.managementService))
	ws.AddRoute(ubadminpanel.RoleEditRoute(ws.managementService))
	ws.AddRoute(ubadminpanel.RoleEditPostRoute(ws.managementService))
	ws.AddRoute(ubadminpanel.UsersListRoute(ws.adapter))
	ws.AddRoute(ubadminpanel.UserOverviewRoute(ws.managementService))
	ws.AddRoute(ubadminpanel.UserRolesListRoute(ws.managementService))
	ws.AddRoute(ubadminpanel.UserRolesAddRoute(ws.managementService))
	ws.AddRoute(ubadminpanel.UserRolesRemoveRoute(ws.managementService))
	ws.AddRoute(ubadminpanel.UserCreateRoute(ws.managementService))
	ws.AddRoute(ubadminpanel.UserCreatePostRoute(ws.managementService))
	ws.AddRoute(ubadminpanel.UserEditRoute(ws.managementService))
	ws.AddRoute(ubadminpanel.LoginRoute(ws.primaryOrganization, ws.managementService, ws.cookieManager))
	ws.AddRoute(ubadminpanel.VerifyTwoFactorRoute(ws.managementService, ws.cookieManager))
	ws.AddRoute(ubadminpanel.LogoutRoute(ws.cookieManager))

	return ws
}

func (ws *WebServiceImpl) AddRoute(route contracts.Route) WebService {
	ws.routes = append(ws.routes, route)
	return ws
}

func (ws *WebServiceImpl) AddRouteHandler(path string, handler http.Handler) WebService {
	ws.mux.Handle(path, handler)
	slog.Info("Registered route handler", "path", path)
	return ws
}

func (ws *WebServiceImpl) AddStatic(filepath string, route string) WebService {
	fs := http.FileServer(http.Dir(filepath))
	ws.mux.Handle(route, http.StripPrefix(route, fs))
	slog.Info("Registered static route", "path", route, "filepath", filepath)
	return ws
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Received request", "method", r.Method, "url", r.URL.Path, "remote_addr", r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func (ws *WebServiceImpl) setupRoutes() {
	for _, route := range ws.routes {
		handler := route.Func
		handler = LoggerMiddleware(http.HandlerFunc(handler)).ServeHTTP
		if route.RequiresPermission != "" {
			handler = ws.permMiddleware.RequirePermission(route.RequiresPermission, handler)
		}
		handler = ws.cookieManager.MiddlewareFunc(handler)

		ws.mux.HandleFunc(route.Path, handler)
		slog.Info("Registered route", "path", route.Path, "requires_permission", route.RequiresPermission, "api", route.Api)
	}
}

func (ws *WebServiceImpl) Start() error {
	ws.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", ws.port),
		Handler: ws.mux,
	}
	ws.setupRoutes()
	go func() {
		slog.Info("Starting web service", "port", ws.port)
		if err := ws.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Web service failed", "error", err)
		}
	}()
	return nil
}

func (ws *WebServiceImpl) Stop() error {
	slog.Info("Stopping web service")
	if err := ws.server.Close(); err != nil {
		slog.Error("Error stopping web service", "error", err)
	}
	return nil
}
