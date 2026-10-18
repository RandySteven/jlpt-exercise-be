package consumers

import (
	"context"

	deepseek_client "github.com/RandySteven/jlpt-exercise-be/pkg/deepseek"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
)

type (
	ExerciseConsumer interface {
		Consumer
	}

	exerciseConsumer struct {
		exerciseRepository repositories.ExerciseRepository
		deepSeek           deepseek_client.DeepSeek
	}
)

func NewExerciseConsumer() ExerciseConsumer {
	return &exerciseConsumer{}
}

func (e *exerciseConsumer) Consumer(ctx context.Context, body string) (err error) {
	//read message from nsq

	//parse message to json

	//ask deepseek to generate note for feedback based on exercise and user's answer

	//save note to database

	//send notification to user

	return nil
}

var _ ExerciseConsumer = &exerciseConsumer{}
