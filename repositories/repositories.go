package repositories

import (
	"context"
	"database/sql"

	repository_interfaces "github.com/RandySteven/go-cook/db"
)

type Repositories struct {
	UserRepository               UserRepository
	SenseiRepository             SenseiRepository
	OnboardingRepository         OnboardingRepository
	ExerciseRepository           ExerciseRepository
	PaymentRepository            PaymentRepository
	SubscriptionRepository       SubscriptionRepository
	SubscriptionPlanRepository   SubscriptionPlanRepository
	ClassRepository              ClassRepository
	ClassUserRepository          ClassUserRepository
	ClassRoomRepository          ClassRoomRepository
	ClassCategoryRepository      ClassCategoryRepository
	ClassCatalogueRepository     ClassCatalogueRepository
	TransactionHistoryRepository TransactionHistoryRepository
	RefundRepository             RefundRepository
}

func NewRepositories(db *sql.DB) *Repositories {
	dbx := func(ctx context.Context) repository_interfaces.Trigger {
		return db
	}

	return &Repositories{
		UserRepository:               newUserRepository(dbx),
		SenseiRepository:             newSenseiRepository(dbx),
		OnboardingRepository:         newOnboardingRepository(dbx),
		ExerciseRepository:           newExerciseRepository(dbx),
		PaymentRepository:            newPaymentRepository(dbx),
		SubscriptionRepository:       newSubscriptionRepository(dbx),
		SubscriptionPlanRepository:   newSubscriptionPlanRepository(dbx),
		ClassRepository:              newClassRepository(dbx),
		ClassUserRepository:          newClassUserRepository(dbx),
		ClassRoomRepository:          newClassRoomRepository(dbx),
		ClassCategoryRepository:      newClassCategoryRepository(dbx),
		ClassCatalogueRepository:     newClassCatalogueRepository(dbx),
		TransactionHistoryRepository: newTransactionHistoryRepository(dbx),
		RefundRepository:             newRefundRepository(dbx),
	}
}
