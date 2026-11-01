package repositories

import (
	"context"

	mysql_client "github.com/RandySteven/go-cook/db"
	repository_interfaces "github.com/RandySteven/go-cook/db"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/queries"
)

type (
	SenseiRepository interface {
		repository_interfaces.Saver[models.Sensei]
		repository_interfaces.Finder[models.Sensei]
		repository_interfaces.Updater[models.Sensei]
		repository_interfaces.Deleter[models.Sensei]
		FindByOnboardingID(ctx context.Context, onboardingID uint64) (result *models.Sensei, err error)
	}

	senseiRepository struct {
		dbx repository_interfaces.DBX
	}
)

func (s *senseiRepository) Save(ctx context.Context, entity *models.Sensei) (result *models.Sensei, err error) {
	id, err := mysql_client.Save[models.Sensei](ctx, s.dbx(ctx), queries.InsertSensei.ToString(), entity.OnboardingID, entity.SenseiName, entity.Status, entity.AdditionalInfo)
	if err != nil {
		return nil, err
	}
	result = entity
	result.ID = *id
	return result, nil
}

func (s *senseiRepository) FindByID(ctx context.Context, id uint64) (result *models.Sensei, err error) {
	result = &models.Sensei{}
	err = mysql_client.FindByID[models.Sensei](ctx, s.dbx(ctx), queries.SelectSenseiByID.ToString(), id, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *senseiRepository) FindAll(ctx context.Context, skip uint64, take uint64) (result []*models.Sensei, err error) {
	return nil, nil
}

func (s *senseiRepository) Update(ctx context.Context, entity *models.Sensei) (result *models.Sensei, err error) {
	err = mysql_client.Update[models.Sensei](
		ctx,
		s.dbx(ctx),
		queries.UpdateSensei.ToString(),
		entity.OnboardingID,
		entity.SenseiName,
		entity.Status,
		entity.ID,
	)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *senseiRepository) DeleteByID(ctx context.Context, id uint64) (err error) {
	return
}

func (s *senseiRepository) FindByOnboardingID(ctx context.Context, onboardingID uint64) (result *models.Sensei, err error) {
	result = &models.Sensei{}
	err = mysql_client.FindByID[models.Sensei](ctx, s.dbx(ctx), queries.SelectSenseiByOnboardingID.ToString(), onboardingID, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func newSenseiRepository(dbx repository_interfaces.DBX) *senseiRepository {
	return &senseiRepository{
		dbx: dbx,
	}
}

var _ SenseiRepository = &senseiRepository{}
