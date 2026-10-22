package repositories

import (
	"context"

	mysql_client "github.com/RandySteven/go-cook/db"
	repository_interfaces "github.com/RandySteven/go-cook/db"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/queries"
)

type (
	SubscriptionRepository interface {
		repository_interfaces.Saver[models.Subscription]
		repository_interfaces.Finder[models.Subscription]
		repository_interfaces.Updater[models.Subscription]
		repository_interfaces.Deleter[models.Subscription]
		FindByPaymentID(ctx context.Context, paymentID uint64) (result *models.Subscription, err error)
		FindCurrentByUserID(ctx context.Context, userID uint64) (result *models.Subscription, err error)
	}

	subscriptionRepository struct {
		dbx repository_interfaces.DBX
	}
)

func (s *subscriptionRepository) Save(ctx context.Context, entity *models.Subscription) (result *models.Subscription, err error) {
	id, err := mysql_client.Save[models.Subscription](
		ctx,
		s.dbx(ctx),
		queries.InsertSubscription.ToString(),
		entity.UserID,
		entity.PaymentID,
		entity.PlanID,
		entity.StartDate,
		entity.EndDate,
		entity.Status,
	)
	if err != nil {
		return nil, err
	}
	result = entity
	result.ID = *id
	return result, nil
}

func (s *subscriptionRepository) FindByID(ctx context.Context, id uint64) (result *models.Subscription, err error) {
	result = &models.Subscription{}
	err = s.dbx(ctx).QueryRowContext(ctx, queries.SelectSubscriptionByID.ToString(), id).Scan(
		&result.ID,
		&result.UserID,
		&result.PaymentID,
		&result.PlanID,
		&result.StartDate,
		&result.EndDate,
		&result.Status,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *subscriptionRepository) FindByPaymentID(ctx context.Context, paymentID uint64) (result *models.Subscription, err error) {
	return s.findOne(ctx, queries.SelectSubscriptionByPaymentID.ToString(), paymentID)
}

func (s *subscriptionRepository) FindCurrentByUserID(ctx context.Context, userID uint64) (result *models.Subscription, err error) {
	return s.findOne(ctx, queries.SelectCurrentSubscriptionByUserID.ToString(), userID)
}

func (s *subscriptionRepository) findOne(ctx context.Context, query string, id uint64) (result *models.Subscription, err error) {
	result = &models.Subscription{}
	err = s.dbx(ctx).QueryRowContext(ctx, query, id).Scan(
		&result.ID,
		&result.UserID,
		&result.PaymentID,
		&result.PlanID,
		&result.StartDate,
		&result.EndDate,
		&result.Status,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *subscriptionRepository) FindAll(ctx context.Context, skip uint64, take uint64) (result []*models.Subscription, err error) {
	rows, err := s.dbx(ctx).QueryContext(ctx, queries.SelectSubscriptions.ToString(), take, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result = make([]*models.Subscription, 0)
	for rows.Next() {
		item := &models.Subscription{}
		if err = rows.Scan(
			&item.ID,
			&item.UserID,
			&item.PaymentID,
			&item.PlanID,
			&item.StartDate,
			&item.EndDate,
			&item.Status,
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

func (s *subscriptionRepository) Update(ctx context.Context, entity *models.Subscription) (result *models.Subscription, err error) {
	err = mysql_client.Update[models.Subscription](
		ctx,
		s.dbx(ctx),
		queries.UpdateSubscription.ToString(),
		entity.UserID,
		entity.PaymentID,
		entity.PlanID,
		entity.StartDate,
		entity.EndDate,
		entity.Status,
		entity.ID,
	)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *subscriptionRepository) DeleteByID(ctx context.Context, id uint64) (err error) {
	_, err = s.dbx(ctx).ExecContext(ctx, queries.DeleteSubscription.ToString(), id)
	return err
}

func newSubscriptionRepository(dbx repository_interfaces.DBX) *subscriptionRepository {
	return &subscriptionRepository{
		dbx: dbx,
	}
}

var _ SubscriptionRepository = &subscriptionRepository{}
