package handlers

import "net/http"

type (
	ISubscriptionHandler interface {
		AddSubscriptionPlan(w http.ResponseWriter, r *http.Request)
		GetSubscriptionPlans(w http.ResponseWriter, r *http.Request)
		GetSubscriptionPlanDetail(w http.ResponseWriter, r *http.Request)
	}
)
