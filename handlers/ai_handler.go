package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	jwt_client "github.com/RandySteven/go-cook/security/v2"
	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/enums"
	"github.com/RandySteven/jlpt-exercise-be/usecases"
	"github.com/RandySteven/jlpt-exercise-be/utils"
	"github.com/coder/websocket"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type (
	IAIHandler interface {
		Chat(w http.ResponseWriter, r *http.Request)
		AvailableModels(w http.ResponseWriter, r *http.Request)
		GetChat(w http.ResponseWriter, r *http.Request)
	}

	AIHandler struct {
		aiUsecase usecases.AIUsecase
	}

	wsClientFrame struct {
		Type            string `json:"type"`
		ClientMessageID string `json:"client_message_id"`
		Content         string `json:"content"`
		RoomID          uint64 `json:"room_id"`
	}

	wsServerFrame struct {
		Type            string `json:"type"`
		RoomID          uint64 `json:"room_id,omitempty"`
		ClientMessageID string `json:"client_message_id,omitempty"`
		Text            string `json:"text,omitempty"`
		Code            string `json:"code,omitempty"`
		Message         string `json:"message,omitempty"`
	}
)

func NewAIHandler(aiUsecase usecases.AIUsecase) IAIHandler {
	return &AIHandler{aiUsecase: aiUsecase}
}

func (h *AIHandler) AvailableModels(w http.ResponseWriter, r *http.Request) {
	dataKey := "models"
	ctx := context.WithValue(r.Context(), enums.RequestID, uuid.New().String())

	response, err := h.aiUsecase.ListModels(ctx)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}
	utils.ResponseHandler(w, http.StatusOK, "success", &dataKey, response, nil)
}

func (h *AIHandler) GetChat(w http.ResponseWriter, r *http.Request) {
	dataKey := "chat"
	ctx := context.WithValue(r.Context(), enums.RequestID, uuid.New().String())

	request := &requests.GetChatRequest{}
	if err := utils.BindJSON(r, &request); err != nil {
		utils.ResponseHandler(w, http.StatusBadRequest, "failed to bind request", nil, nil, err)
		return
	}

	response, err := h.aiUsecase.GetChat(ctx, request)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}
	utils.ResponseHandler(w, http.StatusOK, "success", &dataKey, response, nil)
}

func (h *AIHandler) Chat(w http.ResponseWriter, r *http.Request) {
	ctx, err := authenticateChatRequest(r)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}
	defer conn.CloseNow()

	session, cancel := context.WithCancel(context.WithoutCancel(ctx))
	defer cancel()

	if err := writeWS(session, conn, wsServerFrame{Type: "ready"}); err != nil {
		return
	}

	for {
		_, data, err := conn.Read(session)
		if err != nil {
			return
		}

		var frame wsClientFrame
		if err := json.Unmarshal(data, &frame); err != nil {
			_ = writeWS(session, conn, wsServerFrame{Type: "error", Code: "invalid_frame", Message: "invalid frame"})
			continue
		}

		switch frame.Type {
		case "ping":
			if err := writeWS(session, conn, wsServerFrame{Type: "pong"}); err != nil {
				return
			}
		case "user_message":
			if err := h.streamTurn(session, conn, frame); err != nil {
				return
			}
		default:
			if err := writeWS(session, conn, wsServerFrame{
				Type:            "error",
				Code:            "unknown_type",
				ClientMessageID: frame.ClientMessageID,
				Message:         "unknown frame type",
			}); err != nil {
				return
			}
		}
	}
}

func (h *AIHandler) streamTurn(ctx context.Context, conn *websocket.Conn, frame wsClientFrame) error {
	if err := writeWS(ctx, conn, wsServerFrame{
		Type:            "ack",
		RoomID:          frame.RoomID,
		ClientMessageID: frame.ClientMessageID,
	}); err != nil {
		return err
	}

	stream, err := h.aiUsecase.StreamChat(ctx, &requests.AIChatRequest{
		Message:    frame.Content,
		ChatRoomID: frame.RoomID,
	})
	if err != nil {
		return writeWS(ctx, conn, wsErrorFrame(frame.ClientMessageID, err))
	}

	for chunk := range stream {
		if chunk.Err != nil {
			return writeWS(ctx, conn, wsServerFrame{
				Type:            "error",
				Code:            "stream_failed",
				ClientMessageID: frame.ClientMessageID,
				Message:         chunk.Err.Error(),
			})
		}
		if chunk.Text == "" {
			continue
		}
		if err := writeWS(ctx, conn, wsServerFrame{
			Type:            "delta",
			ClientMessageID: frame.ClientMessageID,
			Text:            chunk.Text,
		}); err != nil {
			return err
		}
	}

	return writeWS(ctx, conn, wsServerFrame{
		Type:            "done",
		ClientMessageID: frame.ClientMessageID,
	})
}

func authenticateChatRequest(r *http.Request) (context.Context, error) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		var err error
		tokenStr, err = utils.GetAuthTokenFromHeader(r)
		if err != nil {
			return nil, apperror.InvalidTokenErr()
		}
	}

	claims := &jwt_client.JWTAccessClaim{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(j *jwt.Token) (interface{}, error) {
		return jwt_client.JwtKey, nil
	})
	if err != nil || !token.Valid {
		return nil, apperror.InvalidTokenErr()
	}

	ctx := context.WithValue(r.Context(), enums.UserID, claims.ID)
	ctx = context.WithValue(ctx, enums.RoleID, claims.Role)
	return ctx, nil
}

func writeWS(ctx context.Context, conn *websocket.Conn, frame wsServerFrame) error {
	data, err := json.Marshal(frame)
	if err != nil {
		return err
	}
	return conn.Write(ctx, websocket.MessageText, data)
}

func wsErrorFrame(clientMessageID string, err error) wsServerFrame {
	frame := wsServerFrame{
		Type:            "error",
		ClientMessageID: clientMessageID,
		Code:            "internal",
		Message:         "the tutor could not start this reply",
	}
	var customErr *apperror.CustomError
	if !errors.As(err, &customErr) {
		return frame
	}
	switch customErr.ErrCode() {
	case http.StatusBadRequest:
		frame.Code = "invalid_message"
		frame.Message = "message is required"
	case http.StatusUnauthorized:
		frame.Code = "unauthorized"
		frame.Message = "invalid token"
	case http.StatusForbidden:
		frame.Code = "forbidden"
		frame.Message = "ai chat is not available on this plan"
	case http.StatusNotFound:
		frame.Code = "subscription_not_found"
		frame.Message = "no active subscription"
	}
	return frame
}

func writeUsecaseError(w http.ResponseWriter, err error) {
	var customErr *apperror.CustomError
	if errors.As(err, &customErr) {
		utils.ResponseHandler(w, customErr.ErrCode(), customErr.Error(), nil, nil, customErr)
		return
	}
	utils.ResponseHandler(w, http.StatusInternalServerError, "internal server error", nil, nil, err)
}
