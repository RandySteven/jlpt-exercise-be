package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/RandySteven/jlpt-exercise-be/enums"
	"github.com/RandySteven/jlpt-exercise-be/usecases"
	"github.com/RandySteven/jlpt-exercise-be/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type (
	ISubscriptionPlanHandler interface {
		AddSubscriptionPlan(w http.ResponseWriter, r *http.Request)
		GetSubscriptionPlans(w http.ResponseWriter, r *http.Request)
		GetSubscriptionPlanDetail(w http.ResponseWriter, r *http.Request)
	}

	SubscriptionPlanHandler struct {
		subscriptionPlanUsecase usecases.SubscriptionPlanUsecase
	}
)

func NewSubscriptionPlanHandler(
	subscriptionPlanUsecase usecases.SubscriptionPlanUsecase,
) ISubscriptionPlanHandler {
	return &SubscriptionPlanHandler{
		subscriptionPlanUsecase: subscriptionPlanUsecase,
	}
}

func (h *SubscriptionPlanHandler) AddSubscriptionPlan(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

func (h *SubscriptionPlanHandler) GetSubscriptionPlans(w http.ResponseWriter, r *http.Request) {
	_, err := h.subscriptionPlanUsecase.ListSubscriptionPlans(r.Context())
	if err != nil {
		return
	}

}

func (h *SubscriptionPlanHandler) GetSubscriptionPlanDetail(w http.ResponseWriter, r *http.Request) {
	dataKey := "subscription_plan_detail"
	id, ok := mux.Vars(r)["id"]
	if !ok {
		utils.ResponseHandler(w, http.StatusBadRequest, "id is required", nil, nil, errors.New("id is required"))
	}
	ctx := context.WithValue(r.Context(), enums.RequestID, uuid.New().String())

	idUint64, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		utils.ResponseHandler(w, http.StatusBadRequest, "id is invalid", nil, nil, errors.New("id is invalid"))
	}

	response, customErr := h.subscriptionPlanUsecase.GetSubscriptionPlanDetail(ctx, idUint64)
	if customErr != nil {
		utils.ResponseHandler(w, customErr.ErrCode(), customErr.Error(), nil, nil, customErr)
		return
	}

	utils.ResponseHandler(w, http.StatusOK, "success", &dataKey, response, nil)
}
