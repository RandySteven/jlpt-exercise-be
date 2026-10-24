package handlers

import "net/http"

type (
	ISubscriptionHandler interface {
		AddSubscriptionPlan(w http.ResponseWriter, r *http.Request)
		GetSubscriptionPlans(w http.ResponseWriter, r *http.Request)
		GetSubscription(w http.ResponseWriter, r *http.Request)
	}
)
