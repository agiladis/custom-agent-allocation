package model

type WebhookPayload struct {
	RoomID int64 `json:"room_id" binding:"required"`
}