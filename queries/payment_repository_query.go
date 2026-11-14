package queries

const (
	InsertPayment GoQuery = `
		INSERT INTO payments (reference_id, transaction_id, transaction_type, user_id, amount, currency, status, metadata)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	UpdatePayment GoQuery = `
		UPDATE payments
		SET
			reference_id = $1,
			transaction_id = $2,
			transaction_type = $3,
			user_id = $4,
			amount = $5,
			currency = $6,
			status = $7,
			metadata = $8,
			updated_at = NOW()
		WHERE
			id = $9
			AND deleted_at IS NULL
	`

	SelectPaymentByID GoQuery = `
		SELECT id, reference_id, transaction_id, transaction_type, user_id, amount, currency, status, metadata, created_at, updated_at, deleted_at
		FROM payments
		WHERE id = $1
		AND deleted_at IS NULL
	`

	SelectPaymentByTransactionID GoQuery = `
		SELECT id, reference_id, transaction_id, transaction_type, user_id, amount, currency, status, metadata, created_at, updated_at, deleted_at
		FROM payments
		WHERE transaction_id = $1
		AND deleted_at IS NULL
	`

	SelectPaymentByReferenceID GoQuery = `
		SELECT id, reference_id, transaction_id, transaction_type, user_id, amount, currency, status, metadata, created_at, updated_at, deleted_at
		FROM payments
		WHERE reference_id = $1
		AND deleted_at IS NULL
	`

	SelectPaymentsByUserID GoQuery = `
		SELECT id, reference_id, transaction_id, transaction_type, user_id, amount, currency, status, metadata, created_at, updated_at, deleted_at
		FROM payments
		WHERE user_id = $1
		AND deleted_at IS NULL
		ORDER BY id DESC
	`

	SelectPayments GoQuery = `
		SELECT id, reference_id, transaction_id, transaction_type, user_id, amount, currency, status, metadata, created_at, updated_at, deleted_at
		FROM payments
		WHERE deleted_at IS NULL
		ORDER BY id DESC
		LIMIT $1 OFFSET $2
	`

	DeletePayment GoQuery = `
		UPDATE payments
		SET
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE
			id = $1
			AND deleted_at IS NULL
	`
)
