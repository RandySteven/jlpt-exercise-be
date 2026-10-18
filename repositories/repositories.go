package repositories

import (
	"context"
	"database/sql"

	repository_interfaces "github.com/RandySteven/go-cook/db"
)

type Repositories struct {
	UserRepository       UserRepository
	OnboardingRepository OnboardingRepository
	ExerciseRepository   ExerciseRepository
}

func NewRepositories(db *sql.DB) *Repositories {
	dbx := func(ctx context.Context) repository_interfaces.Trigger {
		return db
	}

	return &Repositories{
		UserRepository:       newUserRepository(dbx),
		OnboardingRepository: newOnboardingRepository(dbx),
		ExerciseRepository:   newExerciseRepository(dbx),
	}
}
