package onboarding_workflow

import (
	"context"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/enums"
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

	switch executionData.Onboarding.RegisterAs {
	case enums.RegisterAsUser:
		user := executionData.User
		user.Status = `ACTIVE`
		user, err = o.userRepository.Update(ctx, user)
		if err != nil {
			return nil, err
		}
		executionData.User = user
	case enums.RegisterAsSensei:
		sensei := executionData.Sensei
		sensei.Status = `ACTIVE`
		sensei, err = o.senseiRepository.Update(ctx, sensei)
		if err != nil {
			return nil, err
		}
		executionData.Sensei = sensei
	}

	executionData.setPendingActivationResponse()
	return executionData, nil
}
