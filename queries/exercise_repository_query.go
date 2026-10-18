package queries

const (
	InsertExercise GoQuery = `
		INSERT INTO exercises (n_level, exercise, score, status, user_id, note)
		VALUES
			($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	UpdateExercise GoQuery = `
		UPDATE exercises
		SET
			n_level = $1,
			exercise = $2,
			score = $3,
			status = $4,
			user_id = $5,
			note = $6,
			updated_at = NOW()
		WHERE
			id = $7
			AND deleted_at IS NULL
	`

	SelectExerciseByID GoQuery = `
		SELECT id, n_level, exercise, score, status, user_id, note, created_at, updated_at, deleted_at
		FROM exercises
		WHERE id = $1
		AND deleted_at IS NULL
	`

	SelectExercises GoQuery = `
		SELECT id, n_level, exercise, score, status, user_id, note, created_at, updated_at, deleted_at
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
)
