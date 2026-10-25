package queries

const (
	InsertSubscriptionPlan GoQuery = `
		INSERT INTO subscription_plans (name, price, currency, interval, interval_count, feature_flags)
		VALUES
			($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	UpdateSubscriptionPlan GoQuery = `
		UPDATE subscription_plans
		SET
			name = $1,
			price = $2,
			currency = $3,
			interval = $4,
			interval_count = $5,
			feature_flags = $6,
			updated_at = NOW()
		WHERE
			id = $7
			AND deleted_at IS NULL
	`

	SelectSubscriptionPlanByID GoQuery = `
		SELECT id, name, price, currency, interval, interval_count, feature_flags, created_at, updated_at, deleted_at
		FROM subscription_plans
		WHERE id = $1
		AND deleted_at IS NULL
	`

	SelectSubscriptionPlans GoQuery = `
		SELECT id, name, price, currency, interval, interval_count, feature_flags, created_at, updated_at, deleted_at
		FROM subscription_plans
		WHERE deleted_at IS NULL
		ORDER BY id DESC
		LIMIT $1 OFFSET $2
	`

	DeleteSubscriptionPlan GoQuery = `
		UPDATE subscription_plans
		SET
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE
			id = $1
			AND deleted_at IS NULL
	`
)
