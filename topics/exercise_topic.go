package topics

import (
	"context"

	nsq_client "github.com/RandySteven/go-cook/nsq"
)

type (
	ExerciseTopic interface {
		WriteMessage(ctx context.Context, value string) (err error)
		ReadMessage(ctx context.Context) (value string, err error)
	}

	exerciseTopic struct {
		nsq nsq_client.Nsq
	}
)

func NewExerciseTopic(nsq nsq_client.Nsq) ExerciseTopic {
	return &exerciseTopic{
		nsq: nsq,
	}
}

func (e *exerciseTopic) WriteMessage(ctx context.Context, value string) (err error) {
	return e.nsq.Publish(ctx, `exercise-topic`, []byte(value))
}

func (e *exerciseTopic) ReadMessage(ctx context.Context) (value string, err error) {
	return
}

var _ ExerciseTopic = &exerciseTopic{}
