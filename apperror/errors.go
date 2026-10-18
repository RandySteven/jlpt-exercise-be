package apperror

import (
	"log"
	"net/http"
)

type ErrType int

const (
	ErrBadRequest ErrType = iota + 1
	ErrNotFound
	ErrInternalServer
	ErrTimeout
	ErrUnauthorized
	ErrTooManyRequests
	ErrConflict
)

type CustomError struct {
	ErrType    ErrType
	LogMessage string
	err        error
}

func (cu *CustomError) Error() string {
	if cu.err != nil {
		return cu.err.Error()
	}
	return cu.LogMessage
}

func NewCustomError(errType ErrType, logMessage string, err error) *CustomError {
	log.Println(logMessage)
	return &CustomError{
		ErrType:    errType,
		LogMessage: logMessage,
		err:        err,
	}
}

func (cu *CustomError) ErrCode() int {
	switch cu.ErrType {
	case ErrBadRequest:
		return http.StatusBadRequest
	case ErrNotFound:
		return http.StatusNotFound
	case ErrTimeout:
		return http.StatusGatewayTimeout
	case ErrUnauthorized:
		return http.StatusUnauthorized
	case ErrTooManyRequests:
		return http.StatusTooManyRequests
	case ErrConflict:
		return http.StatusConflict
	}
	return http.StatusInternalServerError
}
