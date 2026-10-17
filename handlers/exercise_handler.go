package handlers

import (
	"errors"
	"net/http"

	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/utils"
)

type (
	IExerciseHandler interface {
		GenerateExercise(w http.ResponseWriter, r *http.Request)
		GetExerciseDetail(w http.ResponseWriter, r *http.Request)
		SubmitExercise(w http.ResponseWriter, r *http.Request)
	}

	ExerciseHandler struct {
	}
)

// GenerateExercise implements [IExerciseHandler].
func (e *ExerciseHandler) GenerateExercise(w http.ResponseWriter, r *http.Request) {
	var req requests.GenerateExerciseRequest
	if err := utils.BindJSON(r, &req); err != nil {
		utils.ResponseHandler(w, http.StatusBadRequest, "Invalid request body", nil, nil, err)
		return
	}

	if err := validateJLPTNLevel(req.JLPTNLevel); err != nil {
		utils.ResponseHandler(w, http.StatusBadRequest, "Invalid request body", nil, nil, err)
		return
	}

	// execData, err := e.exerciseWorkflow.GenerateExercise(r.Context(), req)
	// if err != nil {
	// 	utils.ResponseHandler(w, http.StatusInternalServerError, "Failed to generate exercise", nil, nil, err)
	// 	return
	// }

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

func NewExerciseHandler() IExerciseHandler {
	return &ExerciseHandler{}
}
