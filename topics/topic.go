package topics

import (
	nsq_client "github.com/RandySteven/go-cook/nsq"
)

type (

	//Register topics here
	Topics struct {
		DummyTopic    DummyTopic
		ExerciseTopic ExerciseTopic
	}
)

func NewTopics(nsq nsq_client.Nsq) *Topics {
	return &Topics{
		DummyTopic:    NewDummyTopic(nsq),
		ExerciseTopic: NewExerciseTopic(nsq),
	}
}
