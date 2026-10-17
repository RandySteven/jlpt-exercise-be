package onboarding_workflow

import (
	"context"

	temporal_client "github.com/RandySteven/go-cook/temporal"
)

const onboardingInputErrorType = "OnboardingInputError"

func (o *onboardingWorkflow) persistRequest(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	if execData.Request == nil {
		return nil, temporal_client.NewNonRetryableError(onboardingInputErrorType, "register request is missing")
	}

	execData.Onboarding = onboardingRegisterRequestMapping(execData.Request)
	if execData.Onboarding == nil {
		return nil, temporal_client.NewNonRetryableError(onboardingInputErrorType, "failed to map onboarding request")
	}

	onboardingResult, err := o.onboardingRepository.Save(ctx, execData.Onboarding)
	if err != nil {
		return nil, err
	}

	execData.Onboarding = onboardingResult
	execData.SetActivity(registerUserActivity)
	return execData, nil
}
