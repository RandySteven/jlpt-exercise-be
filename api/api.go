package api

import (
	"context"

	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/interfaces/commons"
)

type (
	API interface {
		RegisterUser(ctx context.Context, user *requests.RegisterRequest) (*responses.RegisterResponse, error)
		LoginUser(ctx context.Context, user *requests.LoginRequest) (*responses.LoginResponse, error)
		ActivateUser(ctx context.Context, user *requests.ActivateRequest) (*responses.ActivateResponse, error)
		CreateSubscription(ctx context.Context, subscription *requests.CreateSubscriptionRequest) (*responses.CreateSubscriptionResponse, error)
	}

	APIImpl struct {
		klient commons.Klient
	}
)

func (a *APIImpl) RegisterUser(ctx context.Context, user *requests.RegisterRequest) (*responses.RegisterResponse, error) {
	response, err := a.klient.Call(ctx, "POST", "/auth/register", user, nil)
	if err != nil {
		return nil, err
	}
	return response.(*responses.RegisterResponse), nil
}

func (a *APIImpl) LoginUser(ctx context.Context, user *requests.LoginRequest) (*responses.LoginResponse, error) {
	response, err := a.klient.Call(ctx, "POST", "/auth/login", user, nil)
	if err != nil {
		return nil, err
	}
	return response.(*responses.LoginResponse), nil
}

func (a *APIImpl) ActivateUser(ctx context.Context, user *requests.ActivateRequest) (*responses.ActivateResponse, error) {
	response, err := a.klient.Call(ctx, "POST", "/auth/activated", user, nil)
	if err != nil {
		return nil, err
	}
	return response.(*responses.ActivateResponse), nil
}

func (a *APIImpl) CreateSubscription(ctx context.Context, subscription *requests.CreateSubscriptionRequest) (*responses.CreateSubscriptionResponse, error) {
	response, err := a.klient.Call(ctx, "POST", "/subscriptions", subscription, nil)
	if err != nil {
		return nil, err
	}
	return response.(*responses.CreateSubscriptionResponse), nil
}

func NewAPI(klient commons.Klient) API {
	return &APIImpl{klient: klient}
}
