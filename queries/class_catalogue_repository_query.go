package queries

const (
	InsertClassCatalogue GoQuery = `
		INSERT INTO class_catalogues (name, category_id, description, additional_info)
		VALUES
			($1, $2, $3, $4)
		RETURNING id
	`

	UpdateClassCatalogue GoQuery = `
		UPDATE class_catalogues
		SET
			name = $1,
			category_id = $2,
			description = $3,
			additional_info = $4,
			updated_at = NOW()
		WHERE
			id = $5
			AND deleted_at IS NULL
	`

	SelectClassCatalogueByID GoQuery = `
		SELECT id, name, category_id, description, additional_info, created_at, updated_at, deleted_at
		FROM class_catalogues
		WHERE id = $1
		AND deleted_at IS NULL
	`

	SelectClassCataloguesByCategoryID GoQuery = `
		SELECT id, name, category_id, description, additional_info, created_at, updated_at, deleted_at
		FROM class_catalogues
		WHERE category_id = $1
		AND deleted_at IS NULL
		ORDER BY id DESC
	`

	SelectClassCatalogues GoQuery = `
		SELECT id, name, category_id, description, additional_info, created_at, updated_at, deleted_at
		FROM class_catalogues
		WHERE deleted_at IS NULL
		ORDER BY id DESC
		LIMIT $1 OFFSET $2
	`

	DeleteClassCatalogue GoQuery = `
		UPDATE class_catalogues
		SET
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE
			id = $1
			AND deleted_at IS NULL
	`
)
