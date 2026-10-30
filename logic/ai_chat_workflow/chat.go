package ai_chat_workflow

import (
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"go.temporal.io/sdk/workflow"
)

func (w *ai_chatWorkflow) chat(ctx workflow.Context, request *requests.AIChatRequest) (result *responses.AIChatResponse, err error) {
	return
}
