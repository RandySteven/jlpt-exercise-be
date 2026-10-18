package consumers

import (
	"context"
	"log"
)

type (
	DummyConsumer interface {
		Consumer
	}

	dummyConsumer struct {
	}
)

// Consumer implements [DummyConsumer].
func (d *dummyConsumer) Consumer(ctx context.Context, body string) (err error) {
	log.Println(body)
	return nil
}

var _ DummyConsumer = &dummyConsumer{}

func NewDummyConsumer() DummyConsumer {
	return &dummyConsumer{}
}
