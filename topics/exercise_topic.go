package topics

import (
	"context"
	"encoding/json"

	nsq_client "github.com/RandySteven/go-cook/nsq"
	"github.com/RandySteven/jlpt-exercise-be/entities/schemas"
)

type (
	ExerciseTopic interface {
		WriteMessage(ctx context.Context, schema *schemas.ExerciseSchema) (err error)
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

func (e *exerciseTopic) WriteMessage(ctx context.Context, schema *schemas.ExerciseSchema) (err error) {
	jsonSchema, err := json.Marshal(schema)
	if err != nil {
		return err
	}
	return e.nsq.Publish(ctx, `exercise-topic`, jsonSchema)
}

func (e *exerciseTopic) ReadMessage(ctx context.Context) (value string, err error) {
	return
}

var _ ExerciseTopic = &exerciseTopic{}
