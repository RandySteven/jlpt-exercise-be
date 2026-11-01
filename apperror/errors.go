package apperror

import (
	"log"
	"net/http"
)

type (
	ErrType    int
	ErrMessage string
)

const (
	ErrBadRequest ErrType = iota + 1
	ErrNotFound
	ErrInternalServer
	ErrTimeout
	ErrUnauthorized
	ErrTooManyRequests
	ErrForbidden
	ErrConflict
)

const (
	ErrMessageBadRequest         ErrMessage = "BAD REQUEST"
	ErrMessageInvalidRequest     ErrMessage = "INVALID REQUEST"
	ErrMessageNotFound           ErrMessage = "DATA NOT FOUND"
	ErrMessageInvalidToken       ErrMessage = "INVALID TOKEN"
	ErrMessageFeaturedNotAllowed ErrMessage = "FEATURED NOT ALLOWED"
	ErrMessageDatabaseError      ErrMessage = "DATABASE ERROR"
	ErrMessageTimeout            ErrMessage = "TIMEOUT"
	ErrMessageUnauthorized       ErrMessage = "UNAUTHORIZED"
	ErrMessageTooManyRequests    ErrMessage = "TOO MANY REQUESTS"
	ErrMessageForbidden          ErrMessage = "FORBIDDEN"
	ErrMessageConflict           ErrMessage = "CONFLICT"
	ErrMessageInternalServer     ErrMessage = "INTERNAL SERVER ERROR"
)

func (e ErrMessage) String() string {
	return string(e)
}

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

func (cu *CustomError) Unwrap() error {
	if cu == nil {
		return nil
	}
	return cu.err
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
	case ErrForbidden:
		return http.StatusForbidden
	case ErrConflict:
		return http.StatusConflict
	}
	return http.StatusInternalServerError
}

func BadRequestErr() *CustomError {
	return NewCustomError(ErrBadRequest, ErrMessageBadRequest.String(), nil)
}

func NotFoundErr() *CustomError {
	return NewCustomError(ErrNotFound, ErrMessageNotFound.String(), nil)
}

func DatabaseErrorErr() *CustomError {
	return NewCustomError(ErrInternalServer, ErrMessageDatabaseError.String(), nil)
}

func TimeoutErr() *CustomError {
	return NewCustomError(ErrTimeout, ErrMessageTimeout.String(), nil)
}

func UnauthorizedErr() *CustomError {
	return NewCustomError(ErrUnauthorized, ErrMessageUnauthorized.String(), nil)
}

func TooManyRequestsErr() *CustomError {
	return NewCustomError(ErrTooManyRequests, ErrMessageTooManyRequests.String(), nil)
}

func ForbiddenErr() *CustomError {
	return NewCustomError(ErrForbidden, ErrMessageForbidden.String(), nil)
}

func FeaturedNotAllowedErr() *CustomError {
	return NewCustomError(ErrForbidden, ErrMessageFeaturedNotAllowed.String(), nil)
}

func ConflictErr() *CustomError {
	return NewCustomError(ErrConflict, ErrMessageConflict.String(), nil)
}

func InternalServerErr() *CustomError {
	return NewCustomError(ErrInternalServer, ErrMessageInternalServer.String(), nil)
}

func InvalidRequestErr() *CustomError {
	return NewCustomError(ErrBadRequest, ErrMessageInvalidRequest.String(), nil)
}

func InvalidTokenErr() *CustomError {
	return NewCustomError(ErrUnauthorized, ErrMessageInvalidToken.String(), nil)
}