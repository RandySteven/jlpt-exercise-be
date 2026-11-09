package consumers

import (
	"context"
	"log"

	"github.com/nsqio/go-nsq"
)

type (
	DummyConsumer interface {
		Consumer
	}

	dummyConsumer struct {
	}
)

// Consumer implements [DummyConsumer].
func (d *dummyConsumer) Consumer(ctx context.Context, nsqMessage *nsq.Message) (err error) {
	log.Println(nsqMessage.Body)
	return nil
}

var _ DummyConsumer = &dummyConsumer{}

func NewDummyConsumer() DummyConsumer {
	return &dummyConsumer{}
}
