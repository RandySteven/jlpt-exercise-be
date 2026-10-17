package routes

import (
	"net/http"

	"github.com/RandySteven/go-kopi/enums"
)

// registerEndpointRouter creates a new Router with the given configuration.
// This is an internal helper function used by the HTTP method functions.
//
// Parameters:
//   - methodName: Identifier for the route (used in logs)
//   - path: URL path pattern
//   - method: HTTP method string
//   - handler: Handler function to execute
//   - middlewares: Optional middlewares to apply
func registerEndpointRouter(methodName string, path, method string, handler HandlerFunc, middlewares ...enums.Middleware) *Router {
	return &Router{methodName: methodName, path: path, handler: handler, method: method, middlewares: middlewares}
}

// Post creates a POST route.
//
// Usage:
//
//	Post("CreateUser", "/users", handler)
//	Post("CreateUser", "/users", handler, enums.AuthenticationMiddleware)
func Post(methodName string, path string, handler HandlerFunc, middlewares ...enums.Middleware) *Router {
	return registerEndpointRouter(methodName, path, http.MethodPost, handler, middlewares...)
}

// Get creates a GET route.
//
// Usage:
//
//	Get("GetUsers", "/users", handler)
//	Get("GetUserByID", "/users/{id}", handler)
func Get(methodName string, path string, handler HandlerFunc, middlewares ...enums.Middleware) *Router {
	return registerEndpointRouter(methodName, path, http.MethodGet, handler, middlewares...)
}

// Put creates a PUT route.
//
// Usage:
//
//	Put("UpdateUser", "/users/{id}", handler)
//	Put("UpdateUser", "/users/{id}", handler, enums.AuthenticationMiddleware)
func Put(methodName string, path string, handler HandlerFunc, middlewares ...enums.Middleware) *Router {
	return registerEndpointRouter(methodName, path, http.MethodPut, handler, middlewares...)
}

// Delete creates a DELETE route.
//
// Usage:
//
//	Delete("DeleteUser", "/users/{id}", handler)
//	Delete("DeleteUser", "/users/{id}", handler, enums.AuthenticationMiddleware)
func Delete(methodName string, path string, handler HandlerFunc, middlewares ...enums.Middleware) *Router {
	return registerEndpointRouter(methodName, path, http.MethodDelete, handler, middlewares...)
}

// Patch creates a PATCH route.
//
// Usage:
//
//	Patch("UpdateUserStatus", "/users/{id}/status", handler)
func Patch(methodName string, path string, handler HandlerFunc, middlewares ...enums.Middleware) *Router {
	return registerEndpointRouter(methodName, path, http.MethodPatch, handler, middlewares...)
}
