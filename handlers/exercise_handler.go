package handlers

import (
	"errors"
	"net/http"

	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/usecases"
	"github.com/RandySteven/jlpt-exercise-be/utils"
)

type (
	IExerciseHandler interface {
		GenerateExercise(w http.ResponseWriter, r *http.Request)
		GetExerciseDetail(w http.ResponseWriter, r *http.Request)
		SubmitExercise(w http.ResponseWriter, r *http.Request)
	}

	ExerciseHandler struct {
		exerciseUsecase usecases.ExerciseUsecase
	}
)

// GenerateExercise implements [IExerciseHandler].
func (e *ExerciseHandler) GenerateExercise(w http.ResponseWriter, r *http.Request) {
	var req requests.GenerateExerciseRequest
	dataKey := "generate_exercise_response"
	if err := utils.BindJSON(r, &req); err != nil {
		utils.ResponseHandler(w, http.StatusBadRequest, "Invalid request body", nil, nil, err)
		return
	}

	if err := validateJLPTNLevel(req.JLPTNLevel); err != nil {
		utils.ResponseHandler(w, http.StatusBadRequest, "Invalid request body", nil, nil, err)
		return
	}

	response, customErr := e.exerciseUsecase.GenerateExercise(r.Context(), &req)
	if customErr != nil {
		message := "Failed to generate exercise"
		if customErr.ErrType == apperror.ErrConflict {
			message = "exercise is still in progress"
		}
		utils.ResponseHandler(w, customErr.ErrCode(), message, nil, nil, customErr)
		return
	}

	utils.ResponseHandler(w, http.StatusOK, "Exercise generated successfully", &dataKey, response, nil)
}

// GetExerciseDetail implements [IExerciseHandler].
func (e *ExerciseHandler) GetExerciseDetail(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// SubmitExercise implements [IExerciseHandler].
func (e *ExerciseHandler) SubmitExercise(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

func validateJLPTNLevel(jlptNLevel int) error {
	if jlptNLevel < 1 || jlptNLevel > 5 {
		return errors.New("invalid jlpt n level")
	}
	return nil
}

var _ IExerciseHandler = &ExerciseHandler{}

func NewExerciseHandler(exerciseUsecase usecases.ExerciseUsecase) IExerciseHandler {
	return &ExerciseHandler{
		exerciseUsecase: exerciseUsecase,
	}
}
