package usecases

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/enums"
	"github.com/RandySteven/jlpt-exercise-be/logic/payment_workflow"
	midtrans_client "github.com/RandySteven/jlpt-exercise-be/pkg/midtrans"
	smtp_client "github.com/RandySteven/jlpt-exercise-be/pkg/smtp"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
	"github.com/RandySteven/jlpt-exercise-be/topics"
)

type (
	PaymentUsecase interface {
		CreateSubscription(ctx context.Context, request *requests.CreateSubscriptionRequest) (subscription *responses.CreateSubscriptionResponse, customErr *apperror.CustomError)
		ConfirmPayment(ctx context.Context, request *requests.PaymentConfirmationRequest) (payment *responses.PaymentConfirmationResponse, customErr *apperror.CustomError)
		CancelSubscription(ctx context.Context, request *requests.CancelSubscriptionRequest) (subscription *responses.CancelSubscriptionResponse, customErr *apperror.CustomError)
		GetPaymentHistory(ctx context.Context, request *requests.GetPaymentHistoryRequest) (paymentHistory *responses.GetPaymentHistoryResponse, customErr *apperror.CustomError)
	}

	paymentUsecase struct {
		paymentWorkflow        payment_workflow.PaymentWorkflow
		subscriptionRepository repositories.SubscriptionRepository
		paymentRepository      repositories.PaymentRepository
		userRepository         repositories.UserRepository
		onboardingRepository   repositories.OnboardingRepository
		midtransClient         midtrans_client.MidtransClient
		planRepository         repositories.SubscriptionPlanRepository
		paymentTopic           topics.PaymentTopic
		smtpClient             smtp_client.SmtpClient
		temporal               temporal_client.Temporal
	}
)

func (p *paymentUsecase) CreateSubscription(ctx context.Context, request *requests.CreateSubscriptionRequest) (subscription *responses.CreateSubscriptionResponse, customErr *apperror.CustomError) {
	if request == nil {
		return nil, apperror.NewCustomError(apperror.ErrBadRequest, "request is nil", fmt.Errorf("request is nil"))
	}
	//check if subscription already exists
	userID, ok := ctx.Value(enums.UserID).(uint64)
	if !ok || userID == 0 {
		return nil, apperror.NewCustomError(apperror.ErrUnauthorized, "user is missing", fmt.Errorf("user id is missing"))
	}
	log.Println("userID", userID)
	subscriptionModel, err := p.subscriptionRepository.FindCurrentByUserID(ctx, userID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to find subscription", err)
	}
	if subscriptionModel != nil {
		if subscriptionModel.Status == `ACTIVE` {
			planModel, err := p.planRepository.FindByID(ctx, subscriptionModel.PlanID)
			if err != nil {
				return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to find plan", err)
			}
			if planModel == nil {
				return nil, apperror.NewCustomError(apperror.ErrNotFound, "plan not found", fmt.Errorf("plan not found"))
			}
			return &responses.CreateSubscriptionResponse{
				SubscriptionID: subscriptionModel.ID,
				Status:         subscriptionModel.Status,
				CreatedAt:      subscriptionModel.CreatedAt,
				UpdatedAt:      subscriptionModel.UpdatedAt,
				StartDate:      subscriptionModel.StartDate,
				EndDate:        subscriptionModel.EndDate,
				PlanID:         subscriptionModel.PlanID,
				PlanName:       planModel.Name,
			}, nil
		}
		if subscriptionModel.Status == `PENDING` {
			return p.paymentWorkflow.ContinueSubscription(ctx, request)
		}
		return nil, apperror.NewCustomError(apperror.ErrBadRequest, "subscription already exists", fmt.Errorf("subscription already exists"))
	}
	subscription, customErr = p.paymentWorkflow.CreateSubscription(ctx, request)
	if customErr != nil {
		return nil, customErr
	}
	return subscription, customErr
}

func (p *paymentUsecase) ConfirmPayment(ctx context.Context, request *requests.PaymentConfirmationRequest) (payment *responses.PaymentConfirmationResponse, customErr *apperror.CustomError) {
	if request == nil {
		return nil, apperror.NewCustomError(apperror.ErrBadRequest, "request is nil", fmt.Errorf("request is nil"))
	}
	if request.PlanID == 0 {
		//fallback
		subscriptionModel, err := p.subscriptionRepository.FindByReferenceID(ctx, request.IdempotencyKey)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to find subscription", err)
		}
		if subscriptionModel == nil {
			return nil, apperror.NewCustomError(apperror.ErrNotFound, "subscription not found", fmt.Errorf("subscription not found"))
		}
		request.PlanID = subscriptionModel.PlanID
	}
	payment, customErr = p.paymentWorkflow.PaymentConfirmation(ctx, request)
	if customErr != nil {
		return nil, customErr
	}
	return payment, customErr
}

func (p *paymentUsecase) CancelSubscription(ctx context.Context, request *requests.CancelSubscriptionRequest) (subscription *responses.CancelSubscriptionResponse, customErr *apperror.CustomError) {
	if request == nil {
		return nil, apperror.NewCustomError(apperror.ErrBadRequest, "request is nil", fmt.Errorf("request is nil"))
	}

	subscriptionID := request.SubscriptionID

	subscriptionModel, err := p.subscriptionRepository.FindByID(ctx, subscriptionID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to find subscription", err)
	}
	if subscriptionModel == nil {
		return nil, apperror.NewCustomError(apperror.ErrNotFound, "subscription not found", fmt.Errorf("subscription not found"))
	}

	switch subscriptionModel.Status {
	case `ACTIVE`, `PENDING`:
		subscriptionModel.Status = `CANCELLED`
		_, err = p.subscriptionRepository.Update(ctx, subscriptionModel)
		if err != nil {
			return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to update subscription", err)
		}
	case `CANCELLED`:
		return nil, apperror.NewCustomError(apperror.ErrBadRequest, "subscription is cancelled", fmt.Errorf("subscription is cancelled"))
	case `FAILED`:
		return nil, apperror.NewCustomError(apperror.ErrBadRequest, "subscription is failed", fmt.Errorf("subscription is failed"))
	}
	return nil, nil
}

func (p *paymentUsecase) GetPaymentHistory(ctx context.Context, request *requests.GetPaymentHistoryRequest) (paymentHistory *responses.GetPaymentHistoryResponse, customErr *apperror.CustomError) {
	if request == nil {
		return nil, apperror.NewCustomError(apperror.ErrBadRequest, "request is nil", fmt.Errorf("request is nil"))
	}

	return
}

func newPaymentUsecase(
	paymentWorkflow payment_workflow.PaymentWorkflow,
	subscriptionRepository repositories.SubscriptionRepository,
	paymentRepository repositories.PaymentRepository,
	userRepository repositories.UserRepository,
	onboardingRepository repositories.OnboardingRepository,
	planRepository repositories.SubscriptionPlanRepository,
	midtransClient midtrans_client.MidtransClient,
	paymentTopic topics.PaymentTopic,
	smtpClient smtp_client.SmtpClient,
	temporal temporal_client.Temporal,
) PaymentUsecase {
	pw := &paymentUsecase{
		paymentWorkflow:        paymentWorkflow,
		subscriptionRepository: subscriptionRepository,
		paymentRepository:      paymentRepository,
		userRepository:         userRepository,
		onboardingRepository:   onboardingRepository,
		planRepository:         planRepository,
		midtransClient:         midtransClient,
		paymentTopic:           paymentTopic,
		smtpClient:             smtpClient,
		temporal:               temporal,
	}
	return pw
}
