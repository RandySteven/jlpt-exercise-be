package onboarding_workflow

import (
	"context"

	temporal_client "github.com/RandySteven/go-cook/temporal"
)

func (o *onboardingWorkflow) updateOnboardingStatus(ctx context.Context, executionData *ExecutionData) (*ExecutionData, error) {
	if executionData.Onboarding == nil {
		return nil, temporal_client.NewNonRetryableError(onboardingInputErrorType, "onboarding record is missing")
	}

	onboarding := executionData.Onboarding
	onboarding.Status = `ONBOARDED`
	onboarding, err := o.onboardingRepository.Update(ctx, onboarding)
	if err != nil {
		return nil, err
	}
	executionData.Onboarding = onboarding

	if executionData.User != nil {
		user := executionData.User
		user.Status = `ACTIVE`
		user, err = o.userRepository.Update(ctx, user)
		if err != nil {
			return nil, err
		}
		executionData.User = user
	}

	executionData.setPendingActivationResponse()
	return executionData, nil
}
