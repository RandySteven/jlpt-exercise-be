package onboarding_workflow

import (
	"context"

	temporal_client "github.com/RandySteven/go-cook/temporal"
)

const (
	activationEmailSubject  = "Activate your JLPT account"
	activationEmailTemplate = "activation_onboarding.template"
)

func (o *onboardingWorkflow) sendActivationEmail(ctx context.Context, executionData *ExecutionData) (*ExecutionData, error) {
	if executionData == nil || executionData.Onboarding == nil || executionData.Onboarding.Email == "" {
		return nil, temporal_client.NewNonRetryableError(onboardingInputErrorType, "onboarding email is missing")
	}
	if executionData.ActivationToken == "" {
		return nil, temporal_client.NewNonRetryableError(onboardingInputErrorType, "activation token is missing")
	}
	if o.smtpClient == nil {
		return nil, temporal_client.NewNonRetryableError(onboardingInputErrorType, "smtp client is missing")
	}

	err := o.smtpClient.SendHTMLTemplate(executionData.Onboarding.Email, activationEmailSubject, activationEmailTemplate, map[string]string{
		"onboarding_email": executionData.Onboarding.Email,
		"activation_token": executionData.ActivationToken,
	})
	if err != nil {
		return nil, err
	}

	executionData.SetActivity(updateOnboardingStatusActivity)
	return executionData, nil
}
