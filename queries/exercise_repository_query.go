package queries

const (
	InsertExercise GoQuery = `
		INSERT INTO exercises (reference_id, n_level, exercise, score, status, user_id, note)
		VALUES
			($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	UpdateExercise GoQuery = `
		UPDATE exercises
		SET
			reference_id = $1,
			n_level = $2,
			exercise = $3,
			score = $4,
			status = $5,
			user_id = $6,
			note = $7,
			updated_at = NOW()
		WHERE
			id = $8
			AND deleted_at IS NULL
	`

	SelectExerciseByID GoQuery = `
		SELECT id, reference_id, n_level, exercise, score, status, user_id, note, created_at, updated_at, deleted_at
		FROM exercises
		WHERE id = $1
		AND deleted_at IS NULL
	`

	SelectExercises GoQuery = `
		SELECT id, reference_id, n_level, exercise, score, status, user_id, note, created_at, updated_at, deleted_at
		FROM exercises
		WHERE deleted_at IS NULL
		ORDER BY id DESC
		LIMIT $1 OFFSET $2
	`

	DeleteExercise GoQuery = `
		UPDATE exercises
		SET
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE
			id = $1
			AND deleted_at IS NULL
	`

	SelectUserLatestExercise GoQuery = `
		SELECT id, reference_id, n_level, exercise, score, status, user_id, note, created_at, updated_at, deleted_at
		FROM exercises
		WHERE user_id = $1
		AND deleted_at IS NULL
		ORDER BY id DESC
		LIMIT 1
	`

	UpdateExerciseStatusIfCurrent GoQuery = `
		UPDATE exercises
		SET
			status = $1,
			note = $2,
			updated_at = NOW()
		WHERE
			id = $3
			AND status = $4
			AND deleted_at IS NULL
	`
)
