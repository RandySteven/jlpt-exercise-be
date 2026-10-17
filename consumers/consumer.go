package consumers

import (
	"github.com/RandySteven/go-kopi/caches"
	"github.com/RandySteven/go-kopi/repositories"
	"github.com/RandySteven/go-kopi/topics"
)

type (
	Consumers struct {
		DummyConsumer DummyConsumer
	}

	RunConsumer map[string]ConsumerFunc
)

func (r *Runners) RegisterConsumer(topic string, fun ConsumerFunc) *Runners {
	r.RunConsumers[topic] = fun
	return r
}

func NewConsumers(
	repo *repositories.Repositories,
	cache *caches.Caches,
	topics *topics.Topics,
) *Consumers {
	return &Consumers{
		DummyConsumer: NewDummyConsumer(),
	}
}
