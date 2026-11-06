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
	TransactionHistoryRepository interface {
		repository_interfaces.Saver[models.TransactionHistory]
		repository_interfaces.Finder[models.TransactionHistory]
		repository_interfaces.Updater[models.TransactionHistory]
		repository_interfaces.Deleter[models.TransactionHistory]
		FindByReferenceID(ctx context.Context, referenceID string) (result *models.TransactionHistory, err error)
		FindByUserID(ctx context.Context, userID uint64) (result []*models.TransactionHistory, err error)
		FindBySubscriptionID(ctx context.Context, subscriptionID uint64) (result []*models.TransactionHistory, err error)
	}

	transactionHistoryRepository struct {
		dbx repository_interfaces.DBX
	}
)

func (t *transactionHistoryRepository) Save(ctx context.Context, entity *models.TransactionHistory) (result *models.TransactionHistory, err error) {
	metadataJSON, metadataPayload, err := jsonColumn(entity.Metadata)
	if err != nil {
		return nil, err
	}

	id, err := mysql_client.Save[models.TransactionHistory](
		ctx,
		t.dbx(ctx),
		queries.InsertTransactionHistory.ToString(),
		entity.ReferenceID,
		entity.SubscriptionID,
		entity.UserID,
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

func (t *transactionHistoryRepository) FindByID(ctx context.Context, id uint64) (result *models.TransactionHistory, err error) {
	return t.findOne(ctx, queries.SelectTransactionHistoryByID.ToString(), id)
}

func (t *transactionHistoryRepository) FindByReferenceID(ctx context.Context, referenceID string) (result *models.TransactionHistory, err error) {
	return t.findOne(ctx, queries.SelectTransactionHistoryByReferenceID.ToString(), referenceID)
}

func (t *transactionHistoryRepository) FindByUserID(ctx context.Context, userID uint64) (result []*models.TransactionHistory, err error) {
	rows, err := t.dbx(ctx).QueryContext(ctx, queries.SelectTransactionHistoriesByUserID.ToString(), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTransactionHistories(rows)
}

func (t *transactionHistoryRepository) FindBySubscriptionID(ctx context.Context, subscriptionID uint64) (result []*models.TransactionHistory, err error) {
	rows, err := t.dbx(ctx).QueryContext(ctx, queries.SelectTransactionHistoriesBySubscriptionID.ToString(), subscriptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTransactionHistories(rows)
}

func (t *transactionHistoryRepository) FindAll(ctx context.Context, skip uint64, take uint64) (result []*models.TransactionHistory, err error) {
	rows, err := t.dbx(ctx).QueryContext(ctx, queries.SelectTransactionHistories.ToString(), take, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTransactionHistories(rows)
}

func (t *transactionHistoryRepository) Update(ctx context.Context, entity *models.TransactionHistory) (result *models.TransactionHistory, err error) {
	metadataJSON, metadataPayload, err := jsonColumn(entity.Metadata)
	if err != nil {
		return nil, err
	}

	err = mysql_client.Update[models.TransactionHistory](
		ctx,
		t.dbx(ctx),
		queries.UpdateTransactionHistory.ToString(),
		entity.ReferenceID,
		entity.SubscriptionID,
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

func (t *transactionHistoryRepository) DeleteByID(ctx context.Context, id uint64) (err error) {
	_, err = t.dbx(ctx).ExecContext(ctx, queries.DeleteTransactionHistory.ToString(), id)
	return err
}

func (t *transactionHistoryRepository) findOne(ctx context.Context, query string, arg any) (result *models.TransactionHistory, err error) {
	result = &models.TransactionHistory{}
	err = scanTransactionHistory(t.dbx(ctx).QueryRowContext(ctx, query, arg), result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func scanTransactionHistory(row interface{ Scan(...any) error }, result *models.TransactionHistory) error {
	return row.Scan(
		&result.ID,
		&result.ReferenceID,
		&result.SubscriptionID,
		&result.UserID,
		&result.Amount,
		&result.Currency,
		&result.Status,
		&result.Metadata,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DeletedAt,
	)
}

func scanTransactionHistories(rows *sql.Rows) (result []*models.TransactionHistory, err error) {
	result = make([]*models.TransactionHistory, 0)
	for rows.Next() {
		item := &models.TransactionHistory{}
		if err = scanTransactionHistory(rows, item); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func newTransactionHistoryRepository(dbx repository_interfaces.DBX) *transactionHistoryRepository {
	return &transactionHistoryRepository{
		dbx: dbx,
	}
}

var _ TransactionHistoryRepository = &transactionHistoryRepository{}
