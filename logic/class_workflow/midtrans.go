package class_workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	midtrans_client "github.com/RandySteven/jlpt-exercise-be/pkg/midtrans"
	"github.com/google/uuid"
)

const maxMidtransRetry = 3
const maxPaymentInitiatedRetry = 3

func (c *classWorkflow) initiatedPaymentForStudent(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData.Class == nil {
		return nil, errors.New("class not found")
	}
	if execData.Student == nil {
		return nil, errors.New("student not found")
	}

	classCatalogue, err := c.classCatalogueRepository.FindByID(ctx, execData.Class.ID)
	if err != nil {
		return nil, err
	}

	payment := &models.Payment{
		ReferenceID: fmt.Sprintf("%s-%d", uuid.New().String(), execData.Class.ID),
		UserID:      execData.Student.ID,
		Amount:      classCatalogue.Price,
		Status:      `INITIATED`,
		Currency:    classCatalogue.Currency,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	var saveErr error
	customErr := c.transaction.RunInTx(ctx, func(ctx context.Context) *apperror.CustomError {
		for i := 0; i < maxPaymentInitiatedRetry; i++ {
			if _, err := c.paymentRepository.Save(ctx, payment); err != nil {
				saveErr = err
				continue
			}
			saveErr = nil
			return nil
		}
		return apperror.DatabaseErrorErr()
	})
	if customErr != nil {
		if saveErr != nil {
			return nil, saveErr
		}
		return nil, customErr
	}

	return execData, nil
}

func (c *classWorkflow) initiateMidtransForStudent(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData.Class == nil {
		return nil, errors.New("class not found")
	}
	if execData.Onboarding == nil {
		return nil, errors.New("onboarding not found")
	}

	var lastErr error
	for i := 0; i < maxMidtransRetry; i++ {
		midtransResponse, err := c.midtransClient.CreateTransaction(ctx, &midtrans_client.TransactionRequest{
			ItemDetails: []midtrans_client.ItemDetails{},
		})
		if err != nil {
			lastErr = err
			continue
		}

		execData.NotificationContents = &NotificationContent{
			UserEmail:           execData.Onboarding.Email,
			UserPhoneNumber:     execData.Onboarding.PhoneNumber,
			ClassName:           execData.Class.Name,
			MidtransRedirectURL: midtransResponse.RedirectURL,
		}
		return execData, nil
	}

	if lastErr == nil {
		lastErr = errors.New("failed to create midtrans transaction")
	}
	return nil, lastErr
}

func (c *classWorkflow) confirmClassPayment(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData.Class == nil {
		return nil, errors.New("class not found")
	}

	return execData, nil
}
