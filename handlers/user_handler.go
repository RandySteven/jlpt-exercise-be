package handlers

import (
	"net/http"

	"github.com/RandySteven/go-kopi/usecases"
)

type IUserHandler interface {
	RegisterUser(w http.ResponseWriter, r *http.Request)
	LoginUser(w http.ResponseWriter, r *http.Request)
}

type UserHandler struct {
	userUsecase usecases.UserUsecase
}

func (u *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
}

func (u *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
}

func NewUserHandler(userUsecase usecases.UserUsecase) *UserHandler {
	return &UserHandler{
		userUsecase: userUsecase,
	}
}

var _ IUserHandler = &UserHandler{}
