package repositories

import (
	"context"

	db_client "github.com/RandySteven/go-cook/db"
	mysql_client "github.com/RandySteven/go-cook/db"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/queries"
)

type (
	OnboardingRepository interface {
		Save(ctx context.Context, entity *models.Onboarding) (result *models.Onboarding, err error)
		Update(ctx context.Context, entity *models.Onboarding) (result *models.Onboarding, err error)
		FindByEmail(ctx context.Context, email string) (result *models.Onboarding, err error)
	}

	onboardingRepository struct {
		dbx db_client.DBX
	}
)

func (o *onboardingRepository) Save(ctx context.Context, entity *models.Onboarding) (result *models.Onboarding, err error) {
	jsonRequest, err := entity.AdditionalInfo.MarshalJSON()
	if err != nil {
		return nil, err
	}
	id, err := mysql_client.Save[models.Onboarding](ctx, o.dbx(ctx), queries.InsertOnboarding.ToString(), entity.Name, entity.Email, entity.ProfilePictureURL, entity.Password, entity.PhoneNumber, entity.Address, entity.RegisterAs.ToString(), entity.Status, string(jsonRequest))
	if err != nil {
		return nil, err
	}
	result = entity
	result.ID = *id
	return result, nil
}

func (o *onboardingRepository) FindByEmail(ctx context.Context, email string) (result *models.Onboarding, err error) {
	result = &models.Onboarding{}
	err = o.dbx(ctx).QueryRowContext(ctx, queries.FindOnboardingByEmail.ToString(), email).Scan(
		&result.ID,
		&result.Name,
		&result.Email,
		&result.Password,
		&result.PhoneNumber,
		&result.Address,
		&result.RegisterAs,
		&result.Status,
		&result.ProfilePictureURL,
		&result.AdditionalInfo,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (o *onboardingRepository) Update(ctx context.Context, entity *models.Onboarding) (result *models.Onboarding, err error) {
	err = mysql_client.Update[models.Onboarding](ctx, o.dbx(ctx), queries.UpdateOnboarding.ToString(),
		entity.Email,
		entity.Password,
		entity.PhoneNumber,
		entity.Name,
		entity.ProfilePictureURL,
		entity.Address,
		entity.RegisterAs,
		entity.Status,
		entity.ID,
	)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func newOnboardingRepository(dbx db_client.DBX) *onboardingRepository {
	return &onboardingRepository{
		dbx: dbx,
	}
}

var _ OnboardingRepository = &onboardingRepository{}
