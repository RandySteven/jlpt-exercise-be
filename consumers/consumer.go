package consumers

import (
	"context"

	"github.com/RandySteven/jlpt-exercise-be/caches"
	deepseek_client "github.com/RandySteven/jlpt-exercise-be/pkg/deepseek"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
	"github.com/RandySteven/jlpt-exercise-be/topics"
	"github.com/nsqio/go-nsq"
)

type (
	Consumer interface {
		Consumer(ctx context.Context, nsqMessage *nsq.Message) (err error)
	}

	Consumers struct {
		DummyConsumer    DummyConsumer
		ExerciseConsumer ExerciseConsumer
		PaymentConsumer  PaymentConsumer
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
	deepSeek deepseek_client.DeepSeek,
) *Consumers {
	return &Consumers{
		DummyConsumer:    NewDummyConsumer(),
		ExerciseConsumer: NewExerciseConsumer(repo.ExerciseRepository, cache.ExerciseCache, deepSeek),
		PaymentConsumer:  NewPaymentConsumer(repo.PaymentRepository, repo.TransactionHistoryRepository),
	}
}
