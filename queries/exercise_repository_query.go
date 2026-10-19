package queries

const (
	InsertExercise GoQuery = `
		INSERT INTO exercises (reference_id, n_level, exercise, answer, score, status, user_id, note)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	UpdateExercise GoQuery = `
		UPDATE exercises
		SET
			reference_id = $1,
			n_level = $2,
			exercise = $3,
			answer = $4,
			score = $5,
			status = $6,
			user_id = $7,
			note = $8,
			updated_at = NOW()
		WHERE
			id = $9
			AND deleted_at IS NULL
	`

	SelectExerciseByID GoQuery = `
		SELECT id, reference_id, n_level, exercise, answer, score, status, user_id, note, created_at, updated_at, deleted_at
		FROM exercises
		WHERE id = $1
		AND deleted_at IS NULL
	`

	SelectExercises GoQuery = `
		SELECT id, reference_id, n_level, exercise, answer, score, status, user_id, note, created_at, updated_at, deleted_at
		FROM exercises
		WHERE deleted_at IS NULL
		AND user_id = $1
		ORDER BY id DESC
		LIMIT $2 OFFSET $3
	`

	CountExercisesByUserID GoQuery = `
		SELECT COUNT(*)
		FROM exercises
		WHERE user_id = $1
		AND deleted_at IS NULL
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
		SELECT id, reference_id, n_level, exercise, answer, score, status, user_id, note, created_at, updated_at, deleted_at
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
