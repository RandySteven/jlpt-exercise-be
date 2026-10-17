package repositories

import (
	"context"

	mysql_client "github.com/RandySteven/go-cook/db"
	repository_interfaces "github.com/RandySteven/go-cook/db"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/queries"
)

type (
	UserRepository interface {
		repository_interfaces.Saver[models.User]
		repository_interfaces.Finder[models.User]
		repository_interfaces.Updater[models.User]
		repository_interfaces.Deleter[models.User]
	}

	userRepository struct {
		dbx repository_interfaces.DBX
	}
)

func (u *userRepository) Save(ctx context.Context, entity *models.User) (result *models.User, err error) {
	id, err := mysql_client.Save[models.User](ctx, u.dbx(ctx), queries.InsertUser.ToString(), entity.OnboardingID, entity.UserName, entity.Status, entity.AdditionalInfo)
	if err != nil {
		return nil, err
	}
	result = entity
	result.ID = *id
	return result, nil
}

func (u *userRepository) FindByID(ctx context.Context, id uint64) (result *models.User, err error) {
	result = &models.User{}
	err = mysql_client.FindByID[models.User](ctx, u.dbx(ctx), queries.SelectUserByID.ToString(), id, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (u *userRepository) FindAll(ctx context.Context, skip uint64, take uint64) (result []*models.User, err error) {
	return nil, nil
}

func (u *userRepository) Update(ctx context.Context, entity *models.User) (result *models.User, err error) {
	err = mysql_client.Update[models.User](
		ctx,
		u.dbx(ctx),
		queries.UpdateUser.ToString(),
		entity.OnboardingID,
		entity.UserName,
		entity.Status,
		entity.ID,
	)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (u *userRepository) DeleteByID(ctx context.Context, id uint64) (err error) {
	return
}

func newUserRepository(dbx repository_interfaces.DBX) *userRepository {
	return &userRepository{
		dbx: dbx,
	}
}

var _ UserRepository = &userRepository{}
