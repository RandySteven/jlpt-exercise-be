package topics

import (
	"context"
	"log"

	nsq_client "github.com/RandySteven/go-cook/nsq"
)

type (
	DummyTopic interface {
		Topic
	}

	dummyTopic struct {
		nsq nsq_client.Nsq
	}
)

// ReadMessage implements [DummyTopic].
func (d *dummyTopic) ReadMessage(ctx context.Context) (value string, err error) {
	return
}

// WriteMessage implements [DummyTopic].
func (d *dummyTopic) WriteMessage(ctx context.Context, value string) (err error) {
	err = d.nsq.Publish(ctx, `dummy-topic`, []byte(value))
	if err != nil {
		log.Println(`failed to publish `, err)
		return err
	}
	return nil
}

var _ DummyTopic = &dummyTopic{}

func NewDummyTopic(nsq nsq_client.Nsq) *dummyTopic {
	return &dummyTopic{
		nsq: nsq,
	}
}
