package requests

type (
	CreateClassRoomRequest struct {
		ClassRoomName string `json:"class_room_name" validate:"required"`
		RoomCode      string `json:"room_code" validate:"required"`
	}

	JoinClassRoomRequest struct {
		RoomCode string `json:"room_code" validate:"required"`
	}

	LeaveClassRoomRequest struct{}
)
