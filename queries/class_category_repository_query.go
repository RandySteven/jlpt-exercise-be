package queries

const (
	InsertClassCategory GoQuery = `
		INSERT INTO class_categories (category, description)
		VALUES
			($1, $2)
		RETURNING id
	`

	UpdateClassCategory GoQuery = `
		UPDATE class_categories
		SET
			category = $1,
			description = $2,
			updated_at = NOW()
		WHERE
			id = $3
			AND deleted_at IS NULL
	`

	SelectClassCategoryByID GoQuery = `
		SELECT id, category, description, created_at, updated_at, deleted_at
		FROM class_categories
		WHERE id = $1
		AND deleted_at IS NULL
	`

	SelectClassCategoryByCategory GoQuery = `
		SELECT id, category, description, created_at, updated_at, deleted_at
		FROM class_categories
		WHERE category = $1
		AND deleted_at IS NULL
	`

	SelectClassCategories GoQuery = `
		SELECT id, category, description, created_at, updated_at, deleted_at
		FROM class_categories
		WHERE deleted_at IS NULL
		ORDER BY id DESC
		LIMIT $1 OFFSET $2
	`

	DeleteClassCategory GoQuery = `
		UPDATE class_categories
		SET
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE
			id = $1
			AND deleted_at IS NULL
	`
)
