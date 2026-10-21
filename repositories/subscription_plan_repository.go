package repositories

import (
	"context"

	mysql_client "github.com/RandySteven/go-cook/db"
	repository_interfaces "github.com/RandySteven/go-cook/db"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/queries"
)

type (
	SubscriptionPlanRepository interface {
		repository_interfaces.Saver[models.SubscriptionPlan]
		repository_interfaces.Finder[models.SubscriptionPlan]
		repository_interfaces.Updater[models.SubscriptionPlan]
		repository_interfaces.Deleter[models.SubscriptionPlan]
	}

	subscriptionPlanRepository struct {
		dbx repository_interfaces.DBX
	}
)

func (s *subscriptionPlanRepository) Save(ctx context.Context, entity *models.SubscriptionPlan) (result *models.SubscriptionPlan, err error) {
	id, err := mysql_client.Save[models.SubscriptionPlan](
		ctx,
		s.dbx(ctx),
		queries.InsertSubscriptionPlan.ToString(),
		entity.Name,
		entity.Price,
		entity.Currency,
		entity.Interval,
		entity.IntervalCount,
	)
	if err != nil {
		return nil, err
	}
	result = entity
	result.ID = *id
	return result, nil
}

func (s *subscriptionPlanRepository) FindByID(ctx context.Context, id uint64) (result *models.SubscriptionPlan, err error) {
	result = &models.SubscriptionPlan{}
	err = s.dbx(ctx).QueryRowContext(ctx, queries.SelectSubscriptionPlanByID.ToString(), id).Scan(
		&result.ID,
		&result.Name,
		&result.Price,
		&result.Currency,
		&result.Interval,
		&result.IntervalCount,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *subscriptionPlanRepository) FindAll(ctx context.Context, skip uint64, take uint64) (result []*models.SubscriptionPlan, err error) {
	rows, err := s.dbx(ctx).QueryContext(ctx, queries.SelectSubscriptionPlans.ToString(), take, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result = make([]*models.SubscriptionPlan, 0)
	for rows.Next() {
		item := &models.SubscriptionPlan{}
		if err = rows.Scan(
			&item.ID,
			&item.Name,
			&item.Price,
			&item.Currency,
			&item.Interval,
			&item.IntervalCount,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *subscriptionPlanRepository) Update(ctx context.Context, entity *models.SubscriptionPlan) (result *models.SubscriptionPlan, err error) {
	err = mysql_client.Update[models.SubscriptionPlan](
		ctx,
		s.dbx(ctx),
		queries.UpdateSubscriptionPlan.ToString(),
		entity.Name,
		entity.Price,
		entity.Currency,
		entity.Interval,
		entity.IntervalCount,
		entity.ID,
	)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *subscriptionPlanRepository) DeleteByID(ctx context.Context, id uint64) (err error) {
	_, err = s.dbx(ctx).ExecContext(ctx, queries.DeleteSubscriptionPlan.ToString(), id)
	return err
}

func newSubscriptionPlanRepository(dbx repository_interfaces.DBX) *subscriptionPlanRepository {
	return &subscriptionPlanRepository{
		dbx: dbx,
	}
}

var _ SubscriptionPlanRepository = &subscriptionPlanRepository{}
