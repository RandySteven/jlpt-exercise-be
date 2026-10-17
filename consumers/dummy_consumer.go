package consumers

import (
	"context"
	"log"
)

type (
	DummyConsumer interface {
		DummyFunc(ctx context.Context, body string) (err error)
	}

	dummyConsumer struct {
	}
)

// DummyFunc implements [DummyConsumer].
func (d *dummyConsumer) DummyFunc(ctx context.Context, body string) (err error) {
	log.Println(body)
	return nil
}

var _ DummyConsumer = &dummyConsumer{}

func NewDummyConsumer() DummyConsumer {
	return &dummyConsumer{}
}
