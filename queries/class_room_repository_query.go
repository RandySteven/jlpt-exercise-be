package queries

const (
	InsertClassRoom GoQuery = `
		INSERT INTO class_rooms (class_room_name, room_code, created_by)
		VALUES
			($1, $2, $3)
		RETURNING id
	`

	UpdateClassRoom GoQuery = `
		UPDATE class_rooms
		SET
			class_room_name = $1,
			room_code = $2,
			created_by = $3,
			updated_at = NOW()
		WHERE
			id = $4
			AND deleted_at IS NULL
	`

	SelectClassRoomByID GoQuery = `
		SELECT id, class_room_name, room_code, created_by, created_at, updated_at, deleted_at
		FROM class_rooms
		WHERE id = $1
		AND deleted_at IS NULL
	`

	SelectClassRoomByRoomCode GoQuery = `
		SELECT id, class_room_name, room_code, created_by, created_at, updated_at, deleted_at
		FROM class_rooms
		WHERE room_code = $1
		AND deleted_at IS NULL
	`

	SelectClassRoomsByCreatedBy GoQuery = `
		SELECT id, class_room_name, room_code, created_by, created_at, updated_at, deleted_at
		FROM class_rooms
		WHERE created_by = $1
		AND deleted_at IS NULL
		ORDER BY id DESC
	`

	SelectClassRooms GoQuery = `
		SELECT id, class_room_name, room_code, created_by, created_at, updated_at, deleted_at
		FROM class_rooms
		WHERE deleted_at IS NULL
		ORDER BY id DESC
		LIMIT $1 OFFSET $2
	`

	DeleteClassRoom GoQuery = `
		UPDATE class_rooms
		SET
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE
			id = $1
			AND deleted_at IS NULL
	`
)
