package queries

const (
	InsertSensei GoQuery = `
		INSERT INTO senseis (onboarding_id, sensei_name, status, additional_info)
		VALUES
			($1, $2, $3, $4)
		RETURNING id
	`

	UpdateSensei GoQuery = `
		UPDATE senseis
		SET
			onboarding_id = $1,
			sensei_name = $2,
			status = $3,
			updated_at = NOW()
		WHERE
			id = $4
	`

	SelectSenseiByID GoQuery = `
		SELECT id, onboarding_id, sensei_name, status, additional_info, created_at, updated_at, deleted_at
		FROM senseis
		WHERE id = $1
		AND deleted_at IS NULL
	`

	SelectSenseiByOnboardingID GoQuery = `
		SELECT id, onboarding_id, sensei_name, status, additional_info, created_at, updated_at, deleted_at
		FROM senseis
		WHERE onboarding_id = $1 AND status = 'ACTIVE'
		AND deleted_at IS NULL
	`
)
