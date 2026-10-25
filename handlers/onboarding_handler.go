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
		OAuth2Callback(w http.ResponseWriter, r *http.Request)
		OAuth2CallbackRedirect(w http.ResponseWriter, r *http.Request)
		OAuth2URL(w http.ResponseWriter, r *http.Request)
		GetUserProfile(w http.ResponseWriter, r *http.Request)
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

func (o *OnboardingHandler) OAuth2Callback(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(context.Background(), enums.RequestID, uuid.NewString())

	var dataKey = `oauth2_callback`
	var request = &requests.OAuth2CallbackRequest{}
	if err := utils.BindJSON(r, request); err != nil {
		utils.ResponseHandler(w, http.StatusBadRequest, `invalid request`, nil, nil, err)
		return
	}

	response, customErr := o.onboardingUsecase.OAuth2Callback(ctx, request)
	if customErr != nil {
		utils.ResponseHandler(w, customErr.ErrCode(), customErr.Error(), nil, nil, customErr)
		return
	}
	utils.ResponseHandler(w, http.StatusCreated, `success oauth2 callback`, &dataKey, response, nil)
}

func (o *OnboardingHandler) OAuth2CallbackRedirect(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(context.Background(), enums.RequestID, uuid.NewString())
	code := r.URL.Query().Get("code")
	if code == "" {
		utils.ResponseHandler(w, http.StatusBadRequest, `invalid request`, nil, nil, nil)
		return
	}

	var dataKey = `oauth2_callback`
	response, customErr := o.onboardingUsecase.OAuth2Callback(ctx, &requests.OAuth2CallbackRequest{Code: code})
	if customErr != nil {
		utils.ResponseHandler(w, customErr.ErrCode(), customErr.Error(), nil, nil, customErr)
		return
	}
	utils.ResponseHandler(w, http.StatusCreated, `success oauth2 callback`, &dataKey, response, nil)
}

func (o *OnboardingHandler) OAuth2URL(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(context.Background(), enums.RequestID, uuid.NewString())

	var dataKey = `oauth2_url`
	var request = &requests.OAuth2URLRequest{}
	if err := utils.BindJSON(r, request); err != nil {
		utils.ResponseHandler(w, http.StatusBadRequest, `invalid request`, nil, nil, err)
		return
	}
	response, customErr := o.onboardingUsecase.OAuth2URL(ctx, request)
	if customErr != nil {
		utils.ResponseHandler(w, customErr.ErrCode(), customErr.Error(), nil, nil, customErr)
		return
	}
	utils.ResponseHandler(w, http.StatusCreated, `success oauth2 url`, &dataKey, response, nil)
}

func (o *OnboardingHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), enums.RequestID, uuid.NewString())

	userID, ok := ctx.Value(enums.UserID).(uint64)
	if !ok {
		utils.ResponseHandler(w, http.StatusUnauthorized, `unauthorized`, nil, nil, nil)
		return
	}

	var dataKey = `user_profile`
	response, customErr := o.onboardingUsecase.GetUserProfile(ctx, userID)
	if customErr != nil {
		utils.ResponseHandler(w, customErr.ErrCode(), customErr.Error(), nil, nil, customErr)
		return
	}
	utils.ResponseHandler(w, http.StatusOK, `success get user profile`, &dataKey, response, nil)
}

var _ IOnboardingHandler = &OnboardingHandler{}

func NewOnboardingHandler(onboardingUsecase usecases.OnboardingUsecase) *OnboardingHandler {
	return &OnboardingHandler{
		onboardingUsecase: onboardingUsecase,
	}
}
