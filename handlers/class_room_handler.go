package handlers

import "net/http"

type (
	IClassRoomHandler interface {
		CreateClassRoom(w http.ResponseWriter, r *http.Request)
		JoinClassRoom(w http.ResponseWriter, r *http.Request)
		LeaveClassRoom(w http.ResponseWriter, r *http.Request)
	}

	ClassRoomHandler struct{}
)

// CreateClassRoom implements [IClassRoomHandler].
func (c *ClassRoomHandler) CreateClassRoom(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// JoinClassRoom implements [IClassRoomHandler].
func (c *ClassRoomHandler) JoinClassRoom(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// LeaveClassRoom implements [IClassRoomHandler].
func (c *ClassRoomHandler) LeaveClassRoom(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

var _ IClassRoomHandler = &ClassRoomHandler{}
