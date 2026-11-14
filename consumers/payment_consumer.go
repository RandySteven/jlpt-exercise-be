package consumers

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/entities/schemas"
	"github.com/RandySteven/jlpt-exercise-be/enums"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
	nsq "github.com/nsqio/go-nsq"
)

type (
	PaymentConsumer interface {
		Consumer
	}

	paymentConsumer struct {
		paymentRepository            repositories.PaymentRepository
		transactionHistoryRepository repositories.TransactionHistoryRepository
	}
)

// Consumer implements [PaymentConsumer].
func (p *paymentConsumer) Consumer(ctx context.Context, nsqMessage *nsq.Message) (err error) {
	var schema schemas.PaymentSchema
	err = json.Unmarshal(nsqMessage.Body, &schema)
	if err != nil {
		return err
	}

	userID, err := strconv.ParseUint(schema.UserID, 10, 64)
	if err != nil {
		return err
	}

	transactionHistory := &models.TransactionHistory{
		TransactionID:      schema.TransactionID,
		TransactionType:    enums.TransactionTypePayment,
		TransactionSubType: schema.TransactionType,
		Amount:             int(schema.Amount),
		Currency:           schema.Currency,
		Status:             schema.Status,
		UserID:             userID,
	}

	_, err = p.transactionHistoryRepository.Save(ctx, transactionHistory)
	if err != nil {
		log.Println("failed to save transaction history", err)
		nsqMessage.Requeue(10 * time.Second)
		return err
	}

	log.Println("transaction history saved", transactionHistory.ID)

	nsqMessage.Finish()	
	return nil
}

func NewPaymentConsumer(paymentRepository repositories.PaymentRepository, transactionHistoryRepository repositories.TransactionHistoryRepository) PaymentConsumer {
	return &paymentConsumer{
		paymentRepository:            paymentRepository,
		transactionHistoryRepository: transactionHistoryRepository,
	}
}
