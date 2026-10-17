package queries

const (
	InsertUser GoQuery = `
		INSERT INTO users (onboarding_id, user_name, status, additional_info)
		VALUES
			($1, $2, $3, $4)
		RETURNING id
	`

	UpdateUser GoQuery = `
		UPDATE users
		SET
			onboarding_id = $1,
			user_name = $2,
			status = $3,
			updated_at = NOW()
		WHERE
			id = $4
	`

	SelectUserByID GoQuery = `
		SELECT id, onboarding_id, user_name, status, additional_info, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1
		AND deleted_at IS NULL
	`
)
