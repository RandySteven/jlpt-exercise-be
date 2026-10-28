package topics

import (
	"context"
	"encoding/json"

	nsq_client "github.com/RandySteven/go-cook/nsq"
	"github.com/RandySteven/jlpt-exercise-be/entities/schemas"
)

const paymentTopicName = `payment-topic`

type (
	PaymentTopic interface {
		WriteMessage(ctx context.Context, schema *schemas.PaymentSchema) (err error)
	}

	paymentTopic struct {
		nsq nsq_client.Nsq
	}
)

func NewPaymentTopic(nsq nsq_client.Nsq) PaymentTopic {
	return &paymentTopic{
		nsq: nsq,
	}
}

func (p *paymentTopic) WriteMessage(ctx context.Context, schema *schemas.PaymentSchema) (err error) {
	jsonSchema, err := json.Marshal(schema)
	if err != nil {
		return err
	}
	return p.nsq.Publish(ctx, paymentTopicName, jsonSchema)
}

var _ PaymentTopic = &paymentTopic{}
