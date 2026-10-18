package responses

import "time"

type (
	RegisterResponse struct {
		UserID          uint64 `json:"user_id"`
		OnboardingID    uint64 `json:"onboarding_id"`
		UserName        string `json:"user_name"`
		UserEmail       string `json:"user_email"`
		ActivationToken string `json:"activation_token"`
	}

	LoginResponse struct {
		AccessToken  string    `json:"access_token"`
		RefreshToken string    `json:"refresh_token"`
		LoginAt      time.Time `json:"login_at"`
	}

	ActivateResponse struct {
		Status   string `json:"status"`
		Activate bool   `json:"activate"`
	}

	OAuth2URLResponse struct {
		URL string `json:"url"`
	}

	OAuth2CallbackResponse struct {
		Status       string    `json:"status"`
		Callback     bool      `json:"callback"`
		AccessToken  string    `json:"access_token"`
		RefreshToken string    `json:"refresh_token"`
		LoginAt      time.Time `json:"login_at"`
	}
)
