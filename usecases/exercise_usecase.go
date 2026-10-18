package usecases

import (
	"context"

	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
)

type (
	ExerciseUsecase interface {
		GenerateExercise(ctx context.Context, request *requests.GenerateExerciseRequest) (response *responses.GenerateExerciseResponse, customErr *apperror.CustomError)
		GetExercise(ctx context.Context, id uint64) (response *responses.GenerateExerciseResponse, customErr *apperror.CustomError)
		SubmitExercise(ctx context.Context, request *requests.SubmitExerciseRequest) (response *responses.SubmitExerciseResponse, customErr *apperror.CustomError)
	}
)
