package class_workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/enums"
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
	if classCatalogue == nil || classCatalogue.Price <= 0 {
		return nil, errors.New("class price is missing")
	}

	referenceID := fmt.Sprintf("%s-%d-%d", uuid.New().String(), execData.Class.ID, execData.Student.ID)
	payment := &models.Payment{
		ReferenceID:     referenceID,
		TransactionID:   referenceID,
		TransactionType: enums.TransactionTypeClassPurchase,
		UserID:          execData.Student.ID,
		Amount:          classCatalogue.Price,
		Status:          `INITIATED`,
		Currency:        classCatalogue.Currency,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	var saveErr error
	customErr := c.transaction.RunInTx(ctx, func(ctx context.Context) *apperror.CustomError {
		for i := 0; i < maxPaymentInitiatedRetry; i++ {
			saved, err := c.paymentRepository.Save(ctx, payment)
			if err != nil {
				saveErr = err
				continue
			}
			payment = saved
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

	execData.Payment = payment
	execData.SetActivity(initiateMidtransForStudentActivityName)
	return execData, nil
}

func (c *classWorkflow) initiateMidtransForStudent(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData.Class == nil {
		return nil, errors.New("class not found")
	}
	if execData.Student == nil {
		return nil, errors.New("student not found")
	}
	if execData.Onboarding == nil {
		return nil, errors.New("onboarding not found")
	}
	if execData.Payment == nil || execData.Payment.Amount <= 0 || execData.Payment.ReferenceID == "" {
		return nil, errors.New("payment is missing")
	}

	orderID := execData.Payment.TransactionID
	if orderID == "" {
		orderID = execData.Payment.ReferenceID
	}
	className := execData.Class.Name
	if className == "" {
		className = "Class"
	}

	var lastErr error
	for i := 0; i < maxMidtransRetry; i++ {
		midtransResponse, err := c.midtransClient.CreateTransaction(ctx, &midtrans_client.TransactionRequest{
			TransactionDetails: midtrans_client.TransactionDetails{
				OrderID:     orderID,
				GrossAmount: int64(execData.Payment.Amount),
			},
			ItemDetails: []midtrans_client.ItemDetails{{
				ID:       fmt.Sprintf("%d", execData.Class.ID),
				Price:    int64(execData.Payment.Amount),
				Quantity: 1,
				Name:     className,
			}},
			CustomerDetails: &midtrans_client.CustomerDetails{
				FirstName: execData.Onboarding.Name,
				Email:     execData.Onboarding.Email,
				Phone:     execData.Onboarding.PhoneNumber,
			},
		})
		if err != nil {
			lastErr = err
			continue
		}

		if midtransResponse == nil || midtransResponse.RedirectURL == "" {
			lastErr = errors.New("midtrans redirect url is missing")
			continue
		}

		execData.NotificationContents = &NotificationContent{
			UserEmail:           execData.Onboarding.Email,
			UserPhoneNumber:     execData.Onboarding.PhoneNumber,
			ClassName:           execData.Class.Name,
			MidtransRedirectURL: midtransResponse.RedirectURL,
		}
		execData.SetActivity(publishClassActivityName)
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
