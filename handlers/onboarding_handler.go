package handlers

import (
	"context"
	"net/http"

	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/enums"
	"github.com/RandySteven/jlpt-exercise-be/usecases"
	"github.com/RandySteven/jlpt-exercise-be/utils"
	"github.com/google/uuid"
)

type (
	IOnboardingHandler interface {
		RegisterUser(w http.ResponseWriter, r *http.Request)
		LoginUser(w http.ResponseWriter, r *http.Request)
		ActivateUser(w http.ResponseWriter, r *http.Request)
	}

	OnboardingHandler struct {
		onboardingUsecase usecases.OnboardingUsecase
	}
)

func (o *OnboardingHandler) ActivateUser(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(context.Background(), enums.RequestID, uuid.NewString())

	var dataKey = `register`
	var request = &requests.ActivateRequest{}
	if err := utils.BindJSON(r, request); err != nil {
		utils.ResponseHandler(w, http.StatusBadRequest, `invalid request`, nil, nil, err)
		return
	}

	response, customErr := o.onboardingUsecase.ActivateUser(ctx, request)
	if customErr != nil {
		utils.ResponseHandler(w, customErr.ErrCode(), customErr.Error(), nil, nil, customErr)
		return
	}
	utils.ResponseHandler(w, http.StatusCreated, `success register user`, &dataKey, response, nil)
}

func (o *OnboardingHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(context.Background(), enums.RequestID, uuid.NewString())

	var dataKey = `login`
	var request = &requests.LoginRequest{}
	if err := utils.BindJSON(r, request); err != nil {
		utils.ResponseHandler(w, http.StatusBadRequest, `invalid request`, nil, nil, err)
		return
	}

	response, customErr := o.onboardingUsecase.Login(ctx, request)
	if customErr != nil {
		utils.ResponseHandler(w, customErr.ErrCode(), customErr.Error(), nil, nil, customErr)
		return
	}
	utils.ResponseHandler(w, http.StatusCreated, `success login user`, &dataKey, response, nil)
}

func (o *OnboardingHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(context.Background(), enums.RequestID, uuid.NewString())

	var dataKey = `register`
	var request = &requests.RegisterRequest{}
	if err := utils.BindJSON(r, request); err != nil {
		utils.ResponseHandler(w, http.StatusBadRequest, `invalid request`, nil, nil, err)
		return
	}

	response, customErr := o.onboardingUsecase.Register(ctx, request)
	if customErr != nil {
		utils.ResponseHandler(w, customErr.ErrCode(), customErr.Error(), nil, nil, customErr)
		return
	}
	utils.ResponseHandler(w, http.StatusCreated, `success register user`, &dataKey, response, nil)
}

var _ IOnboardingHandler = &OnboardingHandler{}

func NewOnboardingHandler(onboardingUsecase usecases.OnboardingUsecase) *OnboardingHandler {
	return &OnboardingHandler{
		onboardingUsecase: onboardingUsecase,
	}
}
