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

	"github.com/RandySteven/jlpt-exercise-be/enums"
	"github.com/RandySteven/jlpt-exercise-be/handlers"
	"github.com/RandySteven/jlpt-exercise-be/middlewares"
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
		Post("RegisterUser", "/register", api.OnboardingHandler.RegisterUser),
		Post("LoginUser", "/login", api.OnboardingHandler.LoginUser),
		Post("ActivationUser", "/activated", api.OnboardingHandler.ActivateUser),
		Post("OAuth2URL", "/oauth2/url", api.OnboardingHandler.OAuth2URL),
		Post("OAuth2Callback", "/oauth2/callback", api.OnboardingHandler.OAuth2Callback),
		Get("OAuth2CallbackRedirect", "/oauth2/callback", api.OnboardingHandler.OAuth2CallbackRedirect),
		Get("GetUserProfile", "/profile", api.OnboardingHandler.GetUserProfile, enums.AuthenticationMiddleware),
	}

	endpointRouters[enums.DummyPrefix] = []*Router{
		Get("DummyRequest", ``, api.DummyHandler.DummyRequest),
	}

	endpointRouters[enums.WebhookPrefix] = []*Router{
		Post("Callback", "/callback", api.WebhookHandler.Callback),
	}

	endpointRouters[enums.ExercisePrefix] = []*Router{
		Post("GenerateExercise", "/generate", api.ExerciseHandler.GenerateExercise, enums.AuthenticationMiddleware),
		Post("ListExercise", "/list", api.ExerciseHandler.ListExercise, enums.AuthenticationMiddleware),
		Post("SubmitExercise", "/submit", api.ExerciseHandler.SubmitExercise, enums.AuthenticationMiddleware),
		Get("GetExerciseDetail", "/{id}", api.ExerciseHandler.GetExerciseDetail, enums.AuthenticationMiddleware),
	}

	endpointRouters[enums.PaymentPrefix] = []*Router{
		Post("CreateSubscription", "/subscription", api.PaymentHandler.CreateSubscription, enums.AuthenticationMiddleware),
		Post("ConfirmPayment", "/confirmation", api.PaymentHandler.ConfirmPayment, enums.AuthenticationMiddleware),
		Get("GetSubscription", "/subscription", api.PaymentHandler.GetSubscription, enums.AuthenticationMiddleware),
		Post("PaymentHistory", "/payment-history", api.PaymentHandler.GetPaymentHistory, enums.AuthenticationMiddleware),
	}

	endpointRouters[enums.SubscriptionPlanPrefix] = []*Router{
		Post("AddSubscriptionPlan", "", api.SubscriptionPlanHandler.AddSubscriptionPlan, enums.AuthenticationMiddleware),
		Get("GetSubscriptionPlans", "/list", api.SubscriptionPlanHandler.GetSubscriptionPlans, enums.AuthenticationMiddleware),
		Get("GetSubscriptionPlanDetail", "/{id}", api.SubscriptionPlanHandler.GetSubscriptionPlanDetail, enums.AuthenticationMiddleware),
		Put("FeatureFlagEditSubscriptionPlan", "/feature-flag/{id}", api.SubscriptionPlanHandler.FeatureFlagEditSubscriptionPlan, enums.AuthenticationMiddleware),
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
	for _, route := range routers[enums.AuthPrefix] {
		middleware.RegisterMiddleware(enums.AuthPrefix, route.method, route.path, route.middlewares)
		onboardingRouter.HandleFunc(route.path, route.handler).Methods(route.method, http.MethodOptions)
	}

	dummyRouter := r.PathPrefix(enums.DummyPrefix.ToString()).Subrouter()
	for _, route := range routers[enums.DummyPrefix] {
		middleware.RegisterMiddleware(enums.DummyPrefix, route.method, route.path, route.middlewares)
		dummyRouter.HandleFunc(route.path, route.handler).Methods(route.method, http.MethodOptions)
	}

	exerciseRouter := r.PathPrefix(enums.ExercisePrefix.ToString()).Subrouter()
	for _, route := range routers[enums.ExercisePrefix] {
		middleware.RegisterMiddleware(enums.ExercisePrefix, route.method, route.path, route.middlewares)
		exerciseRouter.HandleFunc(route.path, route.handler).Methods(route.method, http.MethodOptions)
	}

	paymentRouter := r.PathPrefix(enums.PaymentPrefix.ToString()).Subrouter()
	for _, route := range routers[enums.PaymentPrefix] {
		middleware.RegisterMiddleware(enums.PaymentPrefix, route.method, route.path, route.middlewares)
		paymentRouter.HandleFunc(route.path, route.handler).Methods(route.method, http.MethodOptions)
	}

	webhookRouter := r.PathPrefix(enums.WebhookPrefix.ToString()).Subrouter()
	for _, route := range routers[enums.WebhookPrefix] {
		middleware.RegisterMiddleware(enums.WebhookPrefix, route.method, route.path, route.middlewares)
		webhookRouter.HandleFunc(route.path, route.handler).Methods(route.method, http.MethodOptions)
	}

	subscriptionPlanRouter := r.PathPrefix(enums.SubscriptionPlanPrefix.ToString()).Subrouter()
	for _, route := range routers[enums.SubscriptionPlanPrefix] {
		middleware.RegisterMiddleware(enums.SubscriptionPlanPrefix, route.method, route.path, route.middlewares)
		subscriptionPlanRouter.HandleFunc(route.path, route.handler).Methods(route.method, http.MethodOptions)
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
