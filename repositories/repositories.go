package repositories

import (
	"context"
	"database/sql"

	repository_interfaces "github.com/RandySteven/go-cook/db"
)

type Repositories struct {
	UserRepository UserRepository
}

func NewRepositories(db *sql.DB) *Repositories {
	dbx := func(ctx context.Context) repository_interfaces.Trigger {
		return db
	}

	return &Repositories{
		UserRepository: newUserRepository(dbx),
	}
}
