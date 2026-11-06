package queries

const (
	InsertRefund GoQuery = `
		INSERT INTO refunds (reference_id, original_reference_id, payment_id, user_id, amount, status, reason, metadata)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	UpdateRefund GoQuery = `
		UPDATE refunds
		SET
			reference_id = $1,
			original_reference_id = $2,
			payment_id = $3,
			user_id = $4,
			amount = $5,
			status = $6,
			reason = $7,
			metadata = $8,
			updated_at = NOW()
		WHERE
			id = $9
			AND deleted_at IS NULL
	`

	SelectRefundByID GoQuery = `
		SELECT id, reference_id, original_reference_id, payment_id, user_id, amount, status, reason, metadata, created_at, updated_at, deleted_at
		FROM refunds
		WHERE id = $1
		AND deleted_at IS NULL
	`

	SelectRefundByReferenceID GoQuery = `
		SELECT id, reference_id, original_reference_id, payment_id, user_id, amount, status, reason, metadata, created_at, updated_at, deleted_at
		FROM refunds
		WHERE reference_id = $1
		AND deleted_at IS NULL
	`

	SelectRefundsByPaymentID GoQuery = `
		SELECT id, reference_id, original_reference_id, payment_id, user_id, amount, status, reason, metadata, created_at, updated_at, deleted_at
		FROM refunds
		WHERE payment_id = $1
		AND deleted_at IS NULL
		ORDER BY id DESC
	`

	SelectRefundsByUserID GoQuery = `
		SELECT id, reference_id, original_reference_id, payment_id, user_id, amount, status, reason, metadata, created_at, updated_at, deleted_at
		FROM refunds
		WHERE user_id = $1
		AND deleted_at IS NULL
		ORDER BY id DESC
	`

	SelectRefunds GoQuery = `
		SELECT id, reference_id, original_reference_id, payment_id, user_id, amount, status, reason, metadata, created_at, updated_at, deleted_at
		FROM refunds
		WHERE deleted_at IS NULL
		ORDER BY id DESC
		LIMIT $1 OFFSET $2
	`

	DeleteRefund GoQuery = `
		UPDATE refunds
		SET
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE
			id = $1
			AND deleted_at IS NULL
	`
)
