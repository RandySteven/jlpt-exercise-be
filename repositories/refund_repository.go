package repositories

import (
	"context"
	"database/sql"

	mysql_client "github.com/RandySteven/go-cook/db"
	repository_interfaces "github.com/RandySteven/go-cook/db"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/queries"
)

type (
	RefundRepository interface {
		repository_interfaces.Saver[models.Refund]
		repository_interfaces.Finder[models.Refund]
		repository_interfaces.Updater[models.Refund]
		repository_interfaces.Deleter[models.Refund]
		FindByReferenceID(ctx context.Context, referenceID string) (result *models.Refund, err error)
		FindByPaymentID(ctx context.Context, paymentID uint64) (result []*models.Refund, err error)
		FindByUserID(ctx context.Context, userID uint64) (result []*models.Refund, err error)
	}

	refundRepository struct {
		dbx repository_interfaces.DBX
	}
)

func (r *refundRepository) Save(ctx context.Context, entity *models.Refund) (result *models.Refund, err error) {
	metadataJSON, metadataPayload, err := jsonColumn(entity.Metadata)
	if err != nil {
		return nil, err
	}

	id, err := mysql_client.Save[models.Refund](
		ctx,
		r.dbx(ctx),
		queries.InsertRefund.ToString(),
		entity.ReferenceID,
		entity.OriginalReferenceID,
		entity.PaymentID,
		entity.UserID,
		entity.Amount,
		entity.Status,
		entity.Reason,
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

func (r *refundRepository) FindByID(ctx context.Context, id uint64) (result *models.Refund, err error) {
	return r.findOne(ctx, queries.SelectRefundByID.ToString(), id)
}

func (r *refundRepository) FindByReferenceID(ctx context.Context, referenceID string) (result *models.Refund, err error) {
	return r.findOne(ctx, queries.SelectRefundByReferenceID.ToString(), referenceID)
}

func (r *refundRepository) FindByPaymentID(ctx context.Context, paymentID uint64) (result []*models.Refund, err error) {
	rows, err := r.dbx(ctx).QueryContext(ctx, queries.SelectRefundsByPaymentID.ToString(), paymentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRefunds(rows)
}

func (r *refundRepository) FindByUserID(ctx context.Context, userID uint64) (result []*models.Refund, err error) {
	rows, err := r.dbx(ctx).QueryContext(ctx, queries.SelectRefundsByUserID.ToString(), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRefunds(rows)
}

func (r *refundRepository) FindAll(ctx context.Context, skip uint64, take uint64) (result []*models.Refund, err error) {
	rows, err := r.dbx(ctx).QueryContext(ctx, queries.SelectRefunds.ToString(), take, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRefunds(rows)
}

func (r *refundRepository) Update(ctx context.Context, entity *models.Refund) (result *models.Refund, err error) {
	metadataJSON, metadataPayload, err := jsonColumn(entity.Metadata)
	if err != nil {
		return nil, err
	}

	err = mysql_client.Update[models.Refund](
		ctx,
		r.dbx(ctx),
		queries.UpdateRefund.ToString(),
		entity.ReferenceID,
		entity.OriginalReferenceID,
		entity.PaymentID,
		entity.UserID,
		entity.Amount,
		entity.Status,
		entity.Reason,
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

func (r *refundRepository) DeleteByID(ctx context.Context, id uint64) (err error) {
	_, err = r.dbx(ctx).ExecContext(ctx, queries.DeleteRefund.ToString(), id)
	return err
}

func (r *refundRepository) findOne(ctx context.Context, query string, arg any) (result *models.Refund, err error) {
	result = &models.Refund{}
	err = scanRefund(r.dbx(ctx).QueryRowContext(ctx, query, arg), result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func scanRefund(row interface{ Scan(...any) error }, result *models.Refund) error {
	return row.Scan(
		&result.ID,
		&result.ReferenceID,
		&result.OriginalReferenceID,
		&result.PaymentID,
		&result.UserID,
		&result.Amount,
		&result.Status,
		&result.Reason,
		&result.Metadata,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DeletedAt,
	)
}

func scanRefunds(rows *sql.Rows) (result []*models.Refund, err error) {
	result = make([]*models.Refund, 0)
	for rows.Next() {
		item := &models.Refund{}
		if err = scanRefund(rows, item); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func newRefundRepository(dbx repository_interfaces.DBX) *refundRepository {
	return &refundRepository{
		dbx: dbx,
	}
}

var _ RefundRepository = &refundRepository{}
