package queries

const (
	InsertOnboarding GoQuery = `
		INSERT INTO onboardings
		(name, email, profile_picture_url, password, phone_number, address, register_as, status, additional_info)
		VALUES
		($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	UpdateOnboarding GoQuery = `
		UPDATE onboardings
		SET
			email = $1,
			password = $2,
			phone_number = $3,
			name = $4,
			profile_picture_url = $5,
			address = $6,
			register_as = $7,
			status = $8,
			updated_at = NOW()
		WHERE
			id = $9
	`

	FindOnboardingByEmail GoQuery = `
		SELECT id,
			name,
			email,
			password,
			phone_number,
			address,
			register_as,
			status,
			profile_picture_url,
			additional_info,
			created_at,
			updated_at,
			deleted_at
		FROM onboardings
		WHERE email = $1
		AND deleted_at IS NULL
	`
)
