package topics

import (
	"context"
	"encoding/json"

	nsq_client "github.com/RandySteven/go-cook/nsq"
	"github.com/RandySteven/jlpt-exercise-be/entities/schemas"
)

const transactionHistoryTopicName = `transaction-history-topic`

type (
	TransactionHistoryTopic interface {
		WriteMessage(ctx context.Context, schema *schemas.TransactionHistorySchema) (err error)
	}

	transactionHistoryTopic struct {
		nsq nsq_client.Nsq
	}
)

func NewTransactionHistoryTopic(nsq nsq_client.Nsq) TransactionHistoryTopic {
	return &transactionHistoryTopic{
		nsq: nsq,
	}
}

func (t *transactionHistoryTopic) WriteMessage(ctx context.Context, schema *schemas.TransactionHistorySchema) (err error) {
	jsonSchema, err := json.Marshal(schema)
	if err != nil {
		return err
	}
	return t.nsq.Publish(ctx, transactionHistoryTopicName, jsonSchema)
}

var _ TransactionHistoryTopic = &transactionHistoryTopic{}
