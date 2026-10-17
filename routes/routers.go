// Package routes provides HTTP routing configuration and endpoint registration.
//
// This package defines the routing structure for the Go-Kopi framework,
// including route registration, middleware assignment, and router initialization.
//
// Usage:
//
//	api := api_http.NewHTTPs(usecases)
//	routers := routes.NewEndpointRouters(api)
//	routes.InitRouter(routers, muxRouter)
package routes

import (
	"log"
	"net/http"

	"github.com/RandySteven/go-kopi/enums"
	"github.com/RandySteven/go-kopi/handlers"
	"github.com/RandySteven/go-kopi/middlewares"
	"github.com/gorilla/mux"
)

type (
	// HandlerFunc is the standard HTTP handler function signature.
	HandlerFunc func(w http.ResponseWriter, r *http.Request)

	// Router represents a single route configuration including its path,
	// HTTP method, handler function, and associated middlewares.
	//
	// Fields:
	//   - methodName: Identifier name for the route (used for logging)
	//   - path: URL path pattern (e.g., "/users", "/users/{id}")
	//   - handler: The handler function to execute
	//   - method: HTTP method (GET, POST, PUT, DELETE, PATCH)
	//   - middlewares: List of middleware enums to apply to this route
	Router struct {
		methodName  string
		path        string
		handler     HandlerFunc
		method      string
		middlewares []enums.Middleware
	}

	// RouterPrefix maps route prefixes to their associated routes.
	// This allows grouping routes under common URL prefixes.
	//
	// Example:
	//
	//	RouterPrefix{
	//	    enums.AuthPrefix: []*Router{...},    // /auth/*
	//	    enums.UsersPrefix: []*Router{...},   // /users/*
	//	}
	RouterPrefix map[enums.RouterPrefix][]*Router
)

// NewEndpointRouters creates and returns all API endpoint configurations.
// This is where all routes are defined and grouped by their prefix.
//
// Parameters:
//   - api: The HTTPs struct containing all HTTP handlers
//
// Returns:
//   - RouterPrefix: Map of all configured routes grouped by prefix
//
// Example - Adding new routes:
//
//	// Authentication routes (prefix: /auth)
//	endpointRouters[enums.AuthPrefix] = []*Router{
//	    Post("RegisterUser", "/register", api.UserHTTP.RegisterUser),
//	    Post("LoginUser", "/login", api.UserHTTP.LoginUser),
//	}
//
//	// User routes (prefix: /users)
//	endpointRouters[enums.UsersPrefix] = []*Router{
//	    Get("GetUsers", "/", api.UserHTTP.GetUsers, enums.AuthenticationMiddleware),
//	    Get("GetUserByID", "/{id}", api.UserHTTP.GetUserByID),
//	    Put("UpdateUser", "/{id}", api.UserHTTP.UpdateUser, enums.AuthenticationMiddleware),
//	    Delete("DeleteUser", "/{id}", api.UserHTTP.DeleteUser, enums.AuthenticationMiddleware),
//	}
func NewEndpointRouters(api *handlers.Handlers) RouterPrefix {
	endpointRouters := make(RouterPrefix)

	// ============================================================
	// AUTH ROUTES - Prefix: /auth
	// ============================================================
	// POST /auth/register - Register a new user account
	// POST /auth/login    - Authenticate user and get token
	// ============================================================
	endpointRouters[enums.AuthPrefix] = []*Router{
		Post("RegisterUser", "/register", api.UserHandler.RegisterUser),
		Post("LoginUser", "/login", api.UserHandler.LoginUser),
	}

	endpointRouters[enums.DummyPrefix] = []*Router{
		Get("DummyRequest", ``, api.DummyHandler.DummyRequest),
	}

	return endpointRouters
}

// InitRouter initializes the mux router with all configured routes and middlewares.
// It sets up global middlewares and registers all endpoint routes.
//
// Global middlewares applied (in order):
//   - LoggingMiddleware: Logs all incoming requests
//   - CorsMiddleware: Handles CORS headers
//   - TimeoutMiddleware: Sets request timeout
//   - CheckHealthMiddleware: Health check endpoint
//   - AuthenticationMiddleware: JWT authentication
//   - RateLimiterMiddleware: Rate limiting
//
// Parameters:
//   - routers: RouterPrefix containing all route configurations
//   - r: The gorilla/mux router instance
func InitRouter(routers RouterPrefix, r *mux.Router) {
	middleware := middlewares.NewMiddlewares()
	clientMiddleware := middlewares.RegisterClientMiddleware(middleware)
	serverMiddleware := middlewares.RegisterServerMiddleware(middleware)

	r.Use(
		serverMiddleware.LoggingMiddleware,
		serverMiddleware.CorsMiddleware,
		serverMiddleware.TimeoutMiddleware,
		serverMiddleware.CheckHealthMiddleware,
		clientMiddleware.AuthenticationMiddleware,
		clientMiddleware.RateLimiterMiddleware,
	)

	onboardingRouter := r.PathPrefix(enums.AuthPrefix.ToString()).Subrouter()
	for _, routers := range routers[enums.AuthPrefix] {
		onboardingRouter.HandleFunc(routers.path, routers.handler).Methods(routers.method)
	}

	dummyRouter := r.PathPrefix(enums.DummyPrefix.ToString()).Subrouter()
	for _, routers := range routers[enums.DummyPrefix] {
		dummyRouter.HandleFunc(routers.path, routers.handler).Methods(routers.method)
	}
	
	routerLog(r)
}

// log checking registered router
func routerLog(r *mux.Router) {
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()
	
		if len(methods) == 0 {
			log.Printf("         %-30s (subrouter)", path)
		} else {
			log.Printf("%-8s %-30s", methods[0], path)
		}
		return nil
	})
}
