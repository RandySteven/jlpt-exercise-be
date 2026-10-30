package ai_chat_workflow

import (
	"context"
	"time"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type (
	ai_chatWorkflow struct {
		workflow       temporal_client.WorkflowExecution
		temporalClient temporal_client.Temporal
	}

	Ai_chatWorkflow interface {
		CreateChatRoom(ctx context.Context, request *requests.AIChatRequest)
		Chat(ctx context.Context, request *requests.AIChatRequest)
	}
)

var nonRetryableErrorTypes = []string{
	"temporal:activity_failure",
	"temporal:child_workflow_failure",
	"temporal:query_failure",
	"temporal:workflow_execution_already_started",
	"temporal:workflow_execution_already_completed",
	"temporal:workflow_execution_already_closed",
}

var (
	createChatRoomActivity     = `create_chat_room_activity`
	chatActivity               = `chat_activity`
	publishChatActivity        = `publish_chat_activity`
	chatSuccessPublish         = `chat_success_publish`
	chatFailedPublish          = `chat_failed_publish`
	notifyChatSuccess          = `notify_chat_success`
	notifyChatFailed           = `notify_chat_failed`
	consumeSuccessChatActivity = `consume_success_chat_activity`
	persistChatActivity        = `persist_chat_activity`

	sgNoNeed         = ``
	sgCreateChatRoom = `sg_create_chat_room`
	sgChat           = `sg_chat`
	sgConsumeChat    = `sg_consume_chat`

	chatWorkflowName = `chat_workflow`
)

func (w *ai_chatWorkflow) registerActivitiesAndWorkflow() {
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 25 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:        5,
			InitialInterval:        1 * time.Minute,
			BackoffCoefficient:     2,
			MaximumInterval:        5 * time.Minute,
			NonRetryableErrorTypes: nonRetryableErrorTypes,
		},
	}
	w.workflow.AddTransitionActivityWithOptions(createChatRoomActivity, sgCreateChatRoom, w.createChatRoom, &activityOptions, chatActivity)
	w.workflow.AddTransitionActivityWithOptions(chatActivity, sgChat, w.ai, &activityOptions, publishChatActivity)
	w.workflow.AddTransitionActivityWithOptions(publishChatActivity, sgNoNeed, w.publishChat, &activityOptions, chatSuccessPublish, chatFailedPublish)

	w.workflow.AddTransitionActivityWithOptions(notifyChatSuccess, sgNoNeed, w.notify, &activityOptions, sgNoNeed, notifyChatSuccess)
	w.workflow.AddTransitionActivityWithOptions(notifyChatFailed, sgNoNeed, w.notify, &activityOptions, sgNoNeed, notifyChatFailed)

	w.workflow.AddTransitionActivityWithOptions(consumeSuccessChatActivity, sgNoNeed, w.consume, &activityOptions, sgNoNeed, persistChatActivity)
	w.workflow.AddTransitionActivityWithOptions(persistChatActivity, sgNoNeed, w.persist, &activityOptions, sgNoNeed)

	w.workflow.RegisterWorkflow(chatWorkflowName, w.chat)
}

func (w *ai_chatWorkflow) CreateChatRoom(ctx context.Context, request *requests.AIChatRequest) {
	return
}

func (w *ai_chatWorkflow) Chat(ctx context.Context, request *requests.AIChatRequest) {
	return
}