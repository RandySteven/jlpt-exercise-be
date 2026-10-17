package usecases

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	jwt_client "github.com/RandySteven/go-cook/security"
	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/caches"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	onboarding_workflow "github.com/RandySteven/jlpt-exercise-be/logic/onboarding_workflow"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
	"github.com/RandySteven/jlpt-exercise-be/utils"
	"github.com/redis/go-redis/v9"
)

type (
	OnboardingUsecase interface {
		Register(ctx context.Context, request *requests.RegisterRequest) (response *responses.RegisterResponse, customErr *apperror.CustomError)
		Login(ctx context.Context, request *requests.LoginRequest) (response *responses.LoginResponse, customErr *apperror.CustomError)
		ActivateUser(ctx context.Context, request *requests.ActivateRequest) (response *responses.ActivateResponse, customErr *apperror.CustomError)
	}

	onboardingUsecase struct {
		onboardingRepository repositories.OnboardingRepository
		userRepository       repositories.UserRepository
		onboardingCache      caches.OnboardingCache
		onboardingWorkflow   onboarding_workflow.OnboardingWorkflow
	}
)

func (o *onboardingUsecase) Login(ctx context.Context, request *requests.LoginRequest) (response *responses.LoginResponse, customErr *apperror.CustomError) {
	attempt, err := o.onboardingCache.GetLoginAttempt(ctx, request.Email)
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Println("failed to get login attempt", err)
	}

	if attempt >= 3 {
		return nil, apperror.NewCustomError(apperror.ErrTooManyRequests, "too many login attempts", nil)
	}

	onboarding, err := o.onboardingRepository.FindByEmail(ctx, request.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewCustomError(apperror.ErrNotFound, "email not found", err)
		}
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to find onboarding by email", err)
	}

	if !utils.ComparePassword(request.Password, onboarding.Password) {
		if setErr := o.onboardingCache.SetLoginAttempt(ctx, request.Email, attempt+1); setErr != nil {
			log.Println("failed to set login attempt", setErr)
		}
		return nil, apperror.NewCustomError(apperror.ErrUnauthorized, "invalid password", nil)
	}

	if resetErr := o.onboardingCache.SetLoginAttempt(ctx, request.Email, 0); resetErr != nil {
		log.Println("failed to reset login attempt", resetErr)
	}

	accessToken, refreshToken := jwt_client.GenerateTokens(onboarding.ID, onboarding.Email)
	return &responses.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		LoginAt:      time.Now(),
	}, nil
}

func (o *onboardingUsecase) Register(ctx context.Context, request *requests.RegisterRequest) (response *responses.RegisterResponse, customErr *apperror.CustomError) {
	response, customErr = o.onboardingWorkflow.OnboardingRegisterWorkflow(ctx, request)
	if customErr != nil {
		return nil, customErr
	}
	return response, nil
}

func (o *onboardingUsecase) ActivateUser(ctx context.Context, request *requests.ActivateRequest) (response *responses.ActivateResponse, customErr *apperror.CustomError) {
	return o.onboardingWorkflow.ActivateUserRegisterWorkflow(ctx, request)
}

func newOnboardingUsecase(
	onboardingRepository repositories.OnboardingRepository,
	userRepository repositories.UserRepository,
	temporal temporal_client.Temporal,
	onboardingCache caches.OnboardingCache,
) *onboardingUsecase {
	onboardingWorkflow := onboarding_workflow.NewOnboardingWorkflow(
		temporal_client.NewWorkflowExecution(temporal),
		temporal,
		onboardingRepository,
		userRepository,
	)
	return &onboardingUsecase{
		onboardingRepository: onboardingRepository,
		userRepository:       userRepository,
		onboardingWorkflow:   onboardingWorkflow,
		onboardingCache:      onboardingCache,
	}
}

var _ OnboardingUsecase = &onboardingUsecase{}
