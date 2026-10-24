package onboarding_workflow

import (
	"context"
	"fmt"
	"log"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"go.temporal.io/sdk/workflow"
)

func (w *onboardingWorkflow) register(ctx workflow.Context, workflowID string, request *requests.RegisterRequest) (response *responses.RegisterResponse, err error) {
	if request == nil {
		return nil, fmt.Errorf("register request is nil")
	}

	info := workflow.GetInfo(ctx)
	executionData := &ExecutionData{
		Request:    request,
		WorkflowID: workflowID,
		RunID:      info.WorkflowExecution.RunID,
	}

	executionDataResult, err := wfInit(ctx,
		w.workflow,
		queryRegisterResponse,
		executionData, func(executionData temporal_client.ExecutionData) error {
			current, ok := executionData.(*ExecutionData)
			if !ok || !registerResponseReady(current.Response) {
				return fmt.Errorf("response is not ready")
			}
			err := w.onboardingCache.SetRegisterResponse(context.Background(), current.Request.Email, current.Response)
			if err != nil {
				log.Println("failed to set register response to cache", err)
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	executionData, ok := executionDataResult.(*ExecutionData)
	if !ok {
		return nil, fmt.Errorf("there is wrong with convert result")
	}

	if _, err = wfExecute(ctx, w.workflow, executionData); err != nil {
		return nil, err
	}

	return executionData.Response, nil
}

func (w *onboardingWorkflow) registerUser(ctx context.Context, executionData *ExecutionData) (*ExecutionData, error) {
	if executionData.Request == nil {
		return nil, temporal_client.NewNonRetryableError(onboardingInputErrorType, "register request is missing")
	}
	user := mappingUserInfo(executionData.Request.AdditionalInfo)
	if executionData.Onboarding == nil {
		return nil, temporal_client.NewNonRetryableError(onboardingInputErrorType, "onboarding record is missing")
	}

	user.OnboardingID = executionData.Onboarding.ID
	user.UserName = executionData.Request.Username
	user.Status = `PROCESSING`

	user, err := w.userRepository.Save(ctx, user)
	if err != nil {
		return nil, err
	}

	onboardingData := executionData.Onboarding
	onboardingData.Status = `ACTIVE`
	onboardingData, err = w.onboardingRepository.Update(ctx, onboardingData)
	if err != nil {
		return nil, err
	}
	executionData.Onboarding = onboardingData
	executionData.User = user

	executionData.ActivationToken = activationTokenGenerate(user.ID, executionData.Onboarding.ID, executionData.WorkflowID, executionData.RunID)
	executionData.setPendingActivationResponse()
	executionData.SetActivity(sendActivationEmailActivity)
	return executionData, nil
}

func (e *ExecutionData) setPendingActivationResponse() {
	resp := &responses.RegisterResponse{
		ActivationToken: e.ActivationToken,
	}
	if e.Onboarding != nil {
		resp.OnboardingID = e.Onboarding.ID
		resp.UserName = e.Onboarding.Name
		resp.UserEmail = e.Onboarding.Email
	}
	if e.User != nil {
		resp.UserID = e.User.ID
		if e.User.UserName != "" {
			resp.UserName = e.User.UserName
		}
	}
	e.Response = resp
}
