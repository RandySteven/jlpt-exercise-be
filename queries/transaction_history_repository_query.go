package queries

const (
	InsertTransactionHistory GoQuery = `
		INSERT INTO transaction_histories (reference_id, transaction_id, transaction_type, transaction_sub_type, user_id, amount, currency, status, metadata)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	UpdateTransactionHistory GoQuery = `
		UPDATE transaction_histories
		SET
			reference_id = $1,
			transaction_id = $2,
			transaction_type = $3,
			transaction_sub_type = $4,
			user_id = $5,
			amount = $6,
			currency = $7,
			status = $8,
			metadata = $9,
			updated_at = NOW()
		WHERE
			id = $10
			AND deleted_at IS NULL
	`

	SelectTransactionHistoryByID GoQuery = `
		SELECT id, reference_id, transaction_id, transaction_type, transaction_sub_type, user_id, amount, currency, status, metadata, created_at, updated_at, deleted_at
		FROM transaction_histories
		WHERE id = $1
		AND deleted_at IS NULL
	`

	SelectTransactionHistoryByReferenceID GoQuery = `
		SELECT id, reference_id, transaction_id, transaction_type, transaction_sub_type, user_id, amount, currency, status, metadata, created_at, updated_at, deleted_at
		FROM transaction_histories
		WHERE reference_id = $1
		AND deleted_at IS NULL
	`

	SelectTransactionHistoriesByUserID GoQuery = `
		SELECT id, reference_id, transaction_id, transaction_type, transaction_sub_type, user_id, amount, currency, status, metadata, created_at, updated_at, deleted_at
		FROM transaction_histories
		WHERE user_id = $1
		AND deleted_at IS NULL
		ORDER BY id DESC
	`

	SelectTransactionHistoriesByTransactionID GoQuery = `
		SELECT id, reference_id, transaction_id, transaction_type, transaction_sub_type, user_id, amount, currency, status, metadata, created_at, updated_at, deleted_at
		FROM transaction_histories
		WHERE transaction_id = $1
		AND deleted_at IS NULL
		ORDER BY id DESC
	`

	SelectTransactionHistories GoQuery = `
		SELECT id, reference_id, transaction_id, transaction_type, transaction_sub_type, user_id, amount, currency, status, metadata, created_at, updated_at, deleted_at
		FROM transaction_histories
		WHERE deleted_at IS NULL
		ORDER BY id DESC
		LIMIT $1 OFFSET $2
	`

	DeleteTransactionHistory GoQuery = `
		UPDATE transaction_histories
		SET
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE
			id = $1
			AND deleted_at IS NULL
	`
)
