package repositories

import (
	"context"

	mysql_client "github.com/RandySteven/go-cook/db"
	repository_interfaces "github.com/RandySteven/go-cook/db"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/queries"
)

type (
	PaymentRepository interface {
		repository_interfaces.Saver[models.Payment]
		repository_interfaces.Finder[models.Payment]
		repository_interfaces.Updater[models.Payment]
		repository_interfaces.Deleter[models.Payment]
		FindByReferenceID(ctx context.Context, referenceID string) (result *models.Payment, err error)
		FindByUserID(ctx context.Context, userID uint64) (result []*models.Payment, err error)
		FindBySubscriptionID(ctx context.Context, subscriptionID uint64) (result *models.Payment, err error)
	}

	paymentRepository struct {
		dbx repository_interfaces.DBX
	}
)

func (p *paymentRepository) Save(ctx context.Context, entity *models.Payment) (result *models.Payment, err error) {
	metadataJSON, metadataPayload, err := jsonColumn(entity.Metadata)
	if err != nil {
		return nil, err
	}

	id, err := mysql_client.Save[models.Payment](
		ctx,
		p.dbx(ctx),
		queries.InsertPayment.ToString(),
		entity.ReferenceID,
		entity.UserID,
		entity.SubscriptionID,
		entity.Amount,
		entity.Currency,
		entity.Status,
		metadataPayload,
	)
	if err != nil {
		return nil, err
	}
	result = entity
	result.ID = *id
	result.Metadata = metadataJSON
	return result, nil
}

func (p *paymentRepository) FindByID(ctx context.Context, id uint64) (result *models.Payment, err error) {
	result = &models.Payment{}
	err = p.dbx(ctx).QueryRowContext(ctx, queries.SelectPaymentByID.ToString(), id).Scan(
		&result.ID,
		&result.ReferenceID,
		&result.UserID,
		&result.Amount,
		&result.Currency,
		&result.Status,
		&result.Metadata,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (p *paymentRepository) FindByReferenceID(ctx context.Context, referenceID string) (result *models.Payment, err error) {
	result = &models.Payment{}
	err = p.dbx(ctx).QueryRowContext(ctx, queries.SelectPaymentByReferenceID.ToString(), referenceID).Scan(
		&result.ID,
		&result.ReferenceID,
		&result.UserID,
		&result.Amount,
		&result.Currency,
		&result.Status,
		&result.Metadata,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (p *paymentRepository) FindAll(ctx context.Context, skip uint64, take uint64) (result []*models.Payment, err error) {
	rows, err := p.dbx(ctx).QueryContext(ctx, queries.SelectPayments.ToString(), take, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result = make([]*models.Payment, 0)
	for rows.Next() {
		item := &models.Payment{}
		if err = rows.Scan(
			&item.ID,
			&item.ReferenceID,
			&item.UserID,
			&item.Amount,
			&item.Currency,
			&item.Status,
			&item.Metadata,
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

func (p *paymentRepository) Update(ctx context.Context, entity *models.Payment) (result *models.Payment, err error) {
	metadataJSON, metadataPayload, err := jsonColumn(entity.Metadata)
	if err != nil {
		return nil, err
	}

	err = mysql_client.Update[models.Payment](
		ctx,
		p.dbx(ctx),
		queries.UpdatePayment.ToString(),
		entity.ReferenceID,
		entity.UserID,
		entity.Amount,
		entity.Currency,
		entity.Status,
		metadataPayload,
		entity.ID,
	)
	if err != nil {
		return nil, err
	}
	result = entity
	result.Metadata = metadataJSON
	return result, nil
}

func (p *paymentRepository) DeleteByID(ctx context.Context, id uint64) (err error) {
	_, err = p.dbx(ctx).ExecContext(ctx, queries.DeletePayment.ToString(), id)
	return err
}

func (p *paymentRepository) FindByUserID(ctx context.Context, userID uint64) (result []*models.Payment, err error) {
	rows, err := p.dbx(ctx).QueryContext(ctx, queries.SelectPaymentsByUserID.ToString(), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result = make([]*models.Payment, 0)
	for rows.Next() {
		item := &models.Payment{}
		if err = rows.Scan(
			&item.ID,
			&item.ReferenceID,
			&item.UserID,
			&item.Amount,
			&item.Currency,
			&item.Status,
			&item.Metadata,
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

func (p *paymentRepository) FindBySubscriptionID(ctx context.Context, subscriptionID uint64) (result *models.Payment, err error) {
	result = &models.Payment{}
	err = p.dbx(ctx).QueryRowContext(ctx, queries.SelectPaymentBySubscriptionID.ToString(), subscriptionID).Scan(
		&result.ID,
		&result.ReferenceID,
		&result.UserID,
		&result.Amount,
		&result.Currency,
		&result.Status,
		&result.Metadata,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func newPaymentRepository(dbx repository_interfaces.DBX) *paymentRepository {
	return &paymentRepository{
		dbx: dbx,
	}
}

var _ PaymentRepository = &paymentRepository{}
