package queries

const (
	InsertClass GoQuery = `
		INSERT INTO classes (sensei_id, schedules)
		VALUES
			($1, $2)
		RETURNING id
	`

	UpdateClass GoQuery = `
		UPDATE classes
		SET
			sensei_id = $1,
			schedules = $2,
			updated_at = NOW()
		WHERE
			id = $3
			AND deleted_at IS NULL
	`

	SelectClassByID GoQuery = `
		SELECT id, sensei_id, schedules, created_at, updated_at, deleted_at
		FROM classes
		WHERE id = $1
		AND deleted_at IS NULL
	`

	SelectClassesBySenseiID GoQuery = `
		SELECT id, sensei_id, schedules, created_at, updated_at, deleted_at
		FROM classes
		WHERE sensei_id = $1
		AND deleted_at IS NULL
		ORDER BY id DESC
	`

	SelectClasses GoQuery = `
		SELECT id, sensei_id, schedules, created_at, updated_at, deleted_at
		FROM classes
		WHERE deleted_at IS NULL
		ORDER BY id DESC
		LIMIT $1 OFFSET $2
	`

	DeleteClass GoQuery = `
		UPDATE classes
		SET
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE
			id = $1
			AND deleted_at IS NULL
	`
)
