package class_workflow

import (
	"context"
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	midtrans_client "github.com/RandySteven/jlpt-exercise-be/pkg/midtrans"
)

const maxMidtransRetry = 3
const maxPaymentInitiatedRetry = 3

func (c *classWorkflow) initiatedPaymentForStudent(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	errChan := make(chan error)

	c.transaction.RunInTx(ctx, func(ctx context.Context) *apperror.CustomError {
		go func() {
			for userID := range execData.AllowedUsers {
				payment := &models.Payment{
					ReferenceID: fmt.Sprintf("%s-%s", uuid.New().String(), execData.Class.ID),
					UserID:      uint64(userID),
					Amount:      0,
					Status:      `INITIATED`,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				for i := 0; i < maxPaymentInitiatedRetry; i++ {
					if _, err := c.paymentRepository.Save(ctx, payment); err != nil {
						errChan <- err
					}
					break
				}
			}

		}()
		return nil
	})
	return nil, nil
}

func (c *classWorkflow) initiateMidtransForStudent(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	errChan := make(chan error)
	notificationContentsCh := make(chan []*NotificationContent)
	customErr := c.transaction.RunInTx(ctx, func(ctx context.Context) *apperror.CustomError {
		go func() {
			//retry 3 times if midtrans response 500 or 503
			for i := 0; i < maxMidtransRetry; i++ {
				midtransResponse, err := c.midtransClient.CreateTransaction(ctx, &midtrans_client.TransactionRequest{
					ItemDetails: []midtrans_client.ItemDetails{},
				})
				if err != nil {
					if i == maxMidtransRetry-1 {
						errChan <- err
					}
					continue
				}
				notificationContentsCh <- []*NotificationContent{{
					UserEmail:           execData.Onboarding.Email,
					UserPhoneNumber:     execData.Onboarding.PhoneNumber,
					ClassName:           execData.Class.Name,
					MidtransRedirectURL: midtransResponse.RedirectURL,
				}}
				break
			}
			errChan <- errors.New("failed to create midtrans transaction")
		}()

		select {
		case err := <-errChan:
			if err != nil {
				return apperror.DatabaseErrorErr()
			}
		case notificationContents := <-notificationContentsCh:
			execData.NotificationContents = notificationContents
		}
		return nil
	})
	if customErr != nil {
		return nil, customErr
	}
	return execData, nil
}
