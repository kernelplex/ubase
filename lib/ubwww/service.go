package ubwww

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/kernelplex/ubase/lib/contracts"
)

type WebService interface {
	AddRoute(route contracts.Route) WebService
	AddRouteHandler(path string, handler http.Handler) WebService
	Start() error
	Stop() error
}

type WebServiceImpl struct {
	routes         []contracts.Route
	port           uint
	mux            *http.ServeMux
	server         *http.Server
	cookieManager  AuthTokenCookieManager
	permMiddleware *PermissionMiddleware
}

func NewWebService(port uint, cookieManager AuthTokenCookieManager, permMiddleware *PermissionMiddleware) WebService {
	mux := http.NewServeMux()
	return &WebServiceImpl{
		routes:         make([]contracts.Route, 0),
		port:           port,
		mux:            mux,
		cookieManager:  cookieManager,
		permMiddleware: permMiddleware,
	}
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
