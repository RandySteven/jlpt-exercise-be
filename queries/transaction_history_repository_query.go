package queries

const (
	InsertTransactionHistory GoQuery = `
		INSERT INTO transaction_histories (reference_id, subscription_id, user_id, amount, currency, status, metadata)
		VALUES
			($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	UpdateTransactionHistory GoQuery = `
		UPDATE transaction_histories
		SET
			reference_id = $1,
			subscription_id = $2,
			user_id = $3,
			amount = $4,
			currency = $5,
			status = $6,
			metadata = $7,
			updated_at = NOW()
		WHERE
			id = $8
			AND deleted_at IS NULL
	`

	SelectTransactionHistoryByID GoQuery = `
		SELECT id, reference_id, subscription_id, user_id, amount, currency, status, metadata, created_at, updated_at, deleted_at
		FROM transaction_histories
		WHERE id = $1
		AND deleted_at IS NULL
	`

	SelectTransactionHistoryByReferenceID GoQuery = `
		SELECT id, reference_id, subscription_id, user_id, amount, currency, status, metadata, created_at, updated_at, deleted_at
		FROM transaction_histories
		WHERE reference_id = $1
		AND deleted_at IS NULL
	`

	SelectTransactionHistoriesByUserID GoQuery = `
		SELECT id, reference_id, subscription_id, user_id, amount, currency, status, metadata, created_at, updated_at, deleted_at
		FROM transaction_histories
		WHERE user_id = $1
		AND deleted_at IS NULL
		ORDER BY id DESC
	`

	SelectTransactionHistoriesBySubscriptionID GoQuery = `
		SELECT id, reference_id, subscription_id, user_id, amount, currency, status, metadata, created_at, updated_at, deleted_at
		FROM transaction_histories
		WHERE subscription_id = $1
		AND deleted_at IS NULL
		ORDER BY id DESC
	`

	SelectTransactionHistories GoQuery = `
		SELECT id, reference_id, subscription_id, user_id, amount, currency, status, metadata, created_at, updated_at, deleted_at
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
