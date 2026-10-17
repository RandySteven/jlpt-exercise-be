package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/RandySteven/jlpt-exercise-be/topics"
	"github.com/RandySteven/jlpt-exercise-be/utils"
)

type (
	IDummyHandler interface {
		DummyRequest(w http.ResponseWriter, r *http.Request)
	}

	DummyHandler struct {
		dummyTopic topics.DummyTopic
	}
)

// DummyRequest implements [IDummyHandler].
func (d *DummyHandler) DummyRequest(w http.ResponseWriter, r *http.Request) {
	dataKey := `foo`
	ctx := context.Background()
	err := d.dummyTopic.WriteMessage(ctx, `foo bar`)
	if err != nil {
		log.Println("failed to publish err", err)
	}
	utils.ResponseHandler(w, http.StatusOK, `hello`, &dataKey, nil, nil)
}

var _ IDummyHandler = &DummyHandler{}

func NewDummyHandler(dummyTopic topics.DummyTopic) *DummyHandler {
	return &DummyHandler{
		dummyTopic: dummyTopic,
	}
}
