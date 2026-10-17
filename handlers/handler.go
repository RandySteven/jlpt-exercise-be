package handlers

import (
	"github.com/RandySteven/go-kopi/topics"
	"github.com/RandySteven/go-kopi/usecases"
)

type Handlers struct {
	UserHandler  IUserHandler
	DummyHandler IDummyHandler
}

func NewHandlers(usecases *usecases.Usecases, topics *topics.Topics) *Handlers {
	return &Handlers{
		UserHandler: NewUserHandler(usecases.UserUsecase),
		DummyHandler: NewDummyHandler(topics.DummyTopic),
	}
}
