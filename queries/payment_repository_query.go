package queries

const (
	InsertPayment GoQuery = `
		INSERT INTO payments (reference_id, user_id, subscription_id, amount, currency, status, metadata)
		VALUES
			($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	UpdatePayment GoQuery = `
		UPDATE payments
		SET
			reference_id = $1,
			user_id = $2,
			amount = $3,
			currency = $4,
			status = $5,
			metadata = $6,
			updated_at = NOW()
		WHERE
			id = $7
			AND deleted_at IS NULL
	`

	SelectPaymentByID GoQuery = `
		SELECT id, reference_id, user_id, amount, currency, status, metadata, created_at, updated_at, deleted_at
		FROM payments
		WHERE id = $1
		AND deleted_at IS NULL
	`

	SelectPaymentByReferenceID GoQuery = `
		SELECT id, reference_id, user_id, amount, currency, status, metadata, created_at, updated_at, deleted_at
		FROM payments
		WHERE reference_id = $1
		AND deleted_at IS NULL
	`

	SelectPaymentsByUserID GoQuery = `
		SELECT id, reference_id, user_id, amount, currency, status, metadata, created_at, updated_at, deleted_at
		FROM payments
		WHERE user_id = $1
		AND deleted_at IS NULL
		ORDER BY id DESC
		LIMIT $2 OFFSET $3
	`

	SelectPaymentsBySubscriptionID GoQuery = `
		SELECT id, reference_id, user_id, amount, currency, status, metadata, created_at, updated_at, deleted_at
		FROM payments
		WHERE subscription_id = $1
		AND deleted_at IS NULL
	`

	SelectPayments GoQuery = `
		SELECT id, reference_id, user_id, amount, currency, status, metadata, created_at, updated_at, deleted_at
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
