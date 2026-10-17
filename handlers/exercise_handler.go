package handlers

import "net/http"

type (
	ExerciseHandler interface {
		GenerateExercise(w http.ResponseWriter, r *http.Request)
		GetExerciseDetail(w http.ResponseWriter, r *http.Request)
	}
)
