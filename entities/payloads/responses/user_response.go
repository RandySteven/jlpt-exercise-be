package responses

import "time"

type (
	UserLoginResponse struct {
		Token string `json:"token"`
	}

	UserRegisterResponse struct {
		ID        string     `json:"id"`
		Name      string     `json:"name"`
		Email     string     `json:"email"`
		CreatedAt time.Time  `json:"created_at"`
		UpdatedAt time.Time  `json:"updated_at"`
		DeletedAt *time.Time `json:"deleted_at"`
	}
)
