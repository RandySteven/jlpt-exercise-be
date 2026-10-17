package requests

type (
	UserRegisterRequest struct {
		FirstName string `json:"first_name" mapstructure:"first_name" validate:"required"`
		LastName  string `json:"last_name" mapstructure:"last_name" validate:"required"`
		Email     string `json:"email" mapstructure:"email" validate:"required,email"`
		Password  string `json:"password" mapstructure:"password" validate:"required"`
	}

	UserLoginRequest struct {
		Email    string `json:"email" mapstructure:"email" validate:"required,email"`
		Password string `json:"password" mapstructure:"password" validate:"required"`
	}
)
