package topics

import (
	"context"

	nsq_client "github.com/RandySteven/go-cook/nsq"
)

type (
	//List of interface method to register topic
	Topic interface {
		WriteMessage(ctx context.Context, value string) (err error)
		ReadMessage(ctx context.Context) (value string, err error)
	}

	//Register topics here
	Topics struct {
		DummyTopic DummyTopic
	}
)

func NewTopics(nsq nsq_client.Nsq) *Topics {
	return &Topics{
		DummyTopic: NewDummyTopic(nsq),
	}
}
