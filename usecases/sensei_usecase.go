package usecases

import (
	"context"

	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
)

type (
	SenseiUsecase interface {
		AssignSenseiClass(ctx context.Context, request *requests.AssignSenseiClassRequest) (response *responses.AssignSenseiClassResponse, customErr *apperror.CustomError)
		GetSenseiClassSchedule(ctx context.Context, request *requests.GetSenseiClassScheduleRequest) (response *responses.GetSenseiClassScheduleResponse, customErr *apperror.CustomError)
		GetSenseiClassScheduleDetail(ctx context.Context, request *requests.GetSenseiClassScheduleDetailRequest) (response *responses.GetSenseiClassScheduleDetailResponse, customErr *apperror.CustomError)
	}

	senseiUsecase struct {
	}
)

// AssignSenseiClass implements [SenseiUsecase].
func (s *senseiUsecase) AssignSenseiClass(ctx context.Context, request *requests.AssignSenseiClassRequest) (response *responses.AssignSenseiClassResponse, customErr *apperror.CustomError) {
	panic("unimplemented")
}

// GetSenseiClassSchedule implements [SenseiUsecase].
func (s *senseiUsecase) GetSenseiClassSchedule(ctx context.Context, request *requests.GetSenseiClassScheduleRequest) (response *responses.GetSenseiClassScheduleResponse, customErr *apperror.CustomError) {
	panic("unimplemented")
}

// GetSenseiClassScheduleDetail implements [SenseiUsecase].
func (s *senseiUsecase) GetSenseiClassScheduleDetail(ctx context.Context, request *requests.GetSenseiClassScheduleDetailRequest) (response *responses.GetSenseiClassScheduleDetailResponse, customErr *apperror.CustomError) {
	panic("unimplemented")
}

func newSenseiUsecase() SenseiUsecase {
	return &senseiUsecase{}
}
