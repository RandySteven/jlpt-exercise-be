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
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	onboarding_workflow "github.com/RandySteven/jlpt-exercise-be/logic/onboarding_workflow"
	oauth2_client "github.com/RandySteven/jlpt-exercise-be/pkg/oauth2"
	smtp_client "github.com/RandySteven/jlpt-exercise-be/pkg/smtp"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
	"github.com/RandySteven/jlpt-exercise-be/utils"
	"github.com/redis/go-redis/v9"
)

type (
	OnboardingUsecase interface {
		Register(ctx context.Context, request *requests.RegisterRequest) (response *responses.RegisterResponse, customErr *apperror.CustomError)
		Login(ctx context.Context, request *requests.LoginRequest) (response *responses.LoginResponse, customErr *apperror.CustomError)
		ActivateUser(ctx context.Context, request *requests.ActivateRequest) (response *responses.ActivateResponse, customErr *apperror.CustomError)
		OAuth2URL(ctx context.Context, request *requests.OAuth2URLRequest) (response *responses.OAuth2URLResponse, customErr *apperror.CustomError)
		OAuth2Callback(ctx context.Context, request *requests.OAuth2CallbackRequest) (response *responses.OAuth2CallbackResponse, customErr *apperror.CustomError)
		GetUserProfile(ctx context.Context, userID uint64) (response *responses.GetUserProfileResponse, customErr *apperror.CustomError)
	}

	onboardingUsecase struct {
		oauth2Client         oauth2_client.OAuth2Client
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

	user, err := o.userRepository.FindByOnboardingID(ctx, onboarding.ID)
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to find user by onboarding id", err)
	}
	if user == nil {
		return nil, apperror.NewCustomError(apperror.ErrNotFound, "user not found", err)
	}

	accessToken, refreshToken := jwt_client.GenerateTokens(user.ID, onboarding.Email)
	return &responses.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		LoginAt:      time.Now(),
	}, nil
}

func (o *onboardingUsecase) Register(ctx context.Context, request *requests.RegisterRequest) (response *responses.RegisterResponse, customErr *apperror.CustomError) {
	response, err := o.onboardingCache.GetRegisterResponse(ctx, request.Email)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to get register response from cache", err)
	}
	if response != nil {
		return response, nil
	}
	response, customErr = o.onboardingWorkflow.OnboardingRegisterWorkflow(ctx, request)
	if customErr != nil {
		return nil, customErr
	}
	return response, nil
}

func (o *onboardingUsecase) ActivateUser(ctx context.Context, request *requests.ActivateRequest) (response *responses.ActivateResponse, customErr *apperror.CustomError) {
	return o.onboardingWorkflow.ActivateUserRegisterWorkflow(ctx, request)
}

func (o *onboardingUsecase) OAuth2URL(ctx context.Context, request *requests.OAuth2URLRequest) (response *responses.OAuth2URLResponse, customErr *apperror.CustomError) {
	url, err := o.oauth2Client.AuthCodeURL(ctx, request.IdempotencyKey)
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to get auth url", err)
	}
	return &responses.OAuth2URLResponse{
		URL: url,
	}, nil
}

func (o *onboardingUsecase) OAuth2Callback(ctx context.Context, request *requests.OAuth2CallbackRequest) (response *responses.OAuth2CallbackResponse, customErr *apperror.CustomError) {
	userInfo, err := o.oauth2Client.Callback(ctx, request.Code)
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to callback oauth2", err)
	}

	onboarding, err := o.onboardingRepository.FindByEmail(ctx, userInfo.Email)
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to find user by email", err)
	}
	if onboarding == nil {
		onboarding, err = o.onboardingRepository.Save(ctx, &models.Onboarding{
			Email: userInfo.Email,
			Name:  userInfo.Name,
		})
		if err != nil {
			return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to save user", err)
		}
		_, err := o.userRepository.Save(ctx, &models.User{
			OnboardingID: onboarding.ID,
			UserName:     userInfo.Name,
			Status:       `ACTIVE`,
		})
		if err != nil {
			return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to save user", err)
		}
	}

	accessToken, refreshToken := jwt_client.GenerateTokens(onboarding.ID, onboarding.Email)

	return &responses.OAuth2CallbackResponse{
		Status:       "success",
		Callback:     true,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		LoginAt:      time.Now(),
	}, nil
}

func (o *onboardingUsecase) GetUserProfile(ctx context.Context, userID uint64) (response *responses.GetUserProfileResponse, customErr *apperror.CustomError) {
	user, err := o.userRepository.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewCustomError(apperror.ErrNotFound, "user not found", err)
		}
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to find user by id", err)
	}
	onboarding, err := o.onboardingRepository.FindByID(ctx, user.OnboardingID)
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to find onboarding by id", err)
	}
	return &responses.GetUserProfileResponse{
		ID:        user.ID,
		UserName:  onboarding.Name,
		UserEmail: onboarding.Email,
	}, nil
}

func newOnboardingUsecase(
	onboardingRepository repositories.OnboardingRepository,
	userRepository repositories.UserRepository,
	onboardingCache caches.OnboardingCache,
	temporal temporal_client.Temporal,
	oauth2Client oauth2_client.OAuth2Client,
	smtpClient smtp_client.SmtpClient,
) *onboardingUsecase {
	onboardingWorkflow := onboarding_workflow.NewOnboardingWorkflow(
		temporal_client.NewWorkflowExecution(temporal),
		temporal,
		onboardingRepository,
		userRepository,
		onboardingCache,
		smtpClient,
	)
	return &onboardingUsecase{
		onboardingRepository: onboardingRepository,
		userRepository:       userRepository,
		onboardingWorkflow:   onboardingWorkflow,
		onboardingCache:      onboardingCache,
		oauth2Client:         oauth2Client,
	}
}

var _ OnboardingUsecase = &onboardingUsecase{}
