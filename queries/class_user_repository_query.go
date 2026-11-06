package queries

const (
	InsertClassUser GoQuery = `
		INSERT INTO class_users (class_id, user_id)
		VALUES
			($1, $2)
		RETURNING id
	`

	UpdateClassUser GoQuery = `
		UPDATE class_users
		SET
			class_id = $1,
			user_id = $2,
			updated_at = NOW()
		WHERE
			id = $3
			AND deleted_at IS NULL
	`

	SelectClassUserByID GoQuery = `
		SELECT id, class_id, user_id, created_at, updated_at, deleted_at
		FROM class_users
		WHERE id = $1
		AND deleted_at IS NULL
	`

	SelectClassUserByClassIDAndUserID GoQuery = `
		SELECT id, class_id, user_id, created_at, updated_at, deleted_at
		FROM class_users
		WHERE class_id = $1
		AND user_id = $2
		AND deleted_at IS NULL
	`

	SelectClassUsersByClassID GoQuery = `
		SELECT id, class_id, user_id, created_at, updated_at, deleted_at
		FROM class_users
		WHERE class_id = $1
		AND deleted_at IS NULL
		ORDER BY id DESC
	`

	SelectClassUsersByUserID GoQuery = `
		SELECT id, class_id, user_id, created_at, updated_at, deleted_at
		FROM class_users
		WHERE user_id = $1
		AND deleted_at IS NULL
		ORDER BY id DESC
	`

	SelectClassUsers GoQuery = `
		SELECT id, class_id, user_id, created_at, updated_at, deleted_at
		FROM class_users
		WHERE deleted_at IS NULL
		ORDER BY id DESC
		LIMIT $1 OFFSET $2
	`

	DeleteClassUser GoQuery = `
		UPDATE class_users
		SET
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE
			id = $1
			AND deleted_at IS NULL
	`
)
