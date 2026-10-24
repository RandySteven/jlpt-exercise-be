package queries

const (
	InsertSubscription GoQuery = `
		INSERT INTO subscriptions (reference_id, user_id, plan_id, start_date, end_date, status)
		VALUES
			($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	UpdateSubscription GoQuery = `
		UPDATE subscriptions
		SET
			reference_id = $1,
			user_id = $2,
			plan_id = $3,
			start_date = $4,
			end_date = $5,
			status = $6,
			updated_at = NOW()
		WHERE
			id = $7
			AND deleted_at IS NULL
	`

	SelectSubscriptionByID GoQuery = `
		SELECT id, reference_id, user_id, plan_id, start_date, end_date, status, created_at, updated_at, deleted_at
		FROM subscriptions
		WHERE id = $1
		AND deleted_at IS NULL
	`

	SelectSubscriptionByReferenceID GoQuery = `
		SELECT id, reference_id, user_id, plan_id, start_date, end_date, status, created_at, updated_at, deleted_at
		FROM subscriptions
		WHERE reference_id = $1
		AND deleted_at IS NULL
	`

	SelectCurrentSubscriptionByUserID GoQuery = `
		SELECT id, reference_id, user_id, plan_id, start_date, end_date, status, created_at, updated_at, deleted_at
		FROM subscriptions
		WHERE user_id = $1
		AND deleted_at IS NULL
		AND (
			status = 'PENDING'
			OR (status = 'ACTIVE' AND end_date > NOW())
		)
		ORDER BY id DESC
		LIMIT 1
	`

	SelectSubscriptions GoQuery = `
		SELECT id, reference_id, user_id, plan_id, start_date, end_date, status, created_at, updated_at, deleted_at
		FROM subscriptions
		WHERE deleted_at IS NULL
		ORDER BY id DESC
		LIMIT $1 OFFSET $2
	`

	DeleteSubscription GoQuery = `
		UPDATE subscriptions
		SET
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE
			id = $1
			AND deleted_at IS NULL
	`
)
