package consumers

import (
	"context"

	"github.com/RandySteven/jlpt-exercise-be/caches"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
	"github.com/RandySteven/jlpt-exercise-be/topics"
)

type (
	Consumer interface {
		Consumer(ctx context.Context, body string) (err error)
	}

	Consumers struct {
		DummyConsumer    DummyConsumer
		ExerciseConsumer ExerciseConsumer
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
		DummyConsumer:    NewDummyConsumer(),
		ExerciseConsumer: NewExerciseConsumer(),
	}
}
