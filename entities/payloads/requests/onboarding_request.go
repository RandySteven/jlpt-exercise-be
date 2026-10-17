package requests

import "encoding/json"

type (
	RegisterRequest struct {
		FirstName      string          `json:"first_name"`
		LastName       string          `json:"last_name"`
		Email          string          `json:"email"`
		Username       string          `json:"username"`
		Password       string          `json:"password"`
		PhoneNumber    string          `json:"phone_number"`
		Address        string          `json:"address"`
		DateOfBirth    string          `json:"date_of_birth"`
		AdditionalInfo json.RawMessage `json:"additional_info"`
	}

	LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	ActivateRequest struct {
		Token string `json:"token"`
	}
)
