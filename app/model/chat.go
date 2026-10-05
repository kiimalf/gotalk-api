package model

import "time"

type ChatRoom struct {
	ID        int64     `json:"id"`
	User1ID   *int      `json:"user1_id"`
	User2ID   *int      `json:"user2_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ChatRoomResponse struct {
	ID        int64               `json:"id"`
	Partner   *UserPublicResponse `json:"partner,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

type CreateChatRequest struct {
	RecipientID int `json:"recipient_id" validate:"required,gt=0"`
}

type Message struct {
	ID        int64      `json:"id"`
	RoomID    int64      `json:"room_id"`
	SenderID  *int       `json:"sender_id"`
	Content   string     `json:"content"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type MessageResponse struct {
	ID        int64     `json:"id"`
	RoomID    int64     `json:"room_id"`
	SenderID  int       `json:"sender_id"`
	Content   string    `json:"content"`
	IsEdited  bool      `json:"is_edited"`
	CreatedAt time.Time `json:"created_at"`
}

type SendMessageRequest struct {
	Content string `json:"content" validate:"required,min=1,max=4000"`
}

type UpdateMessageRequest struct {
	Content string `json:"content" validate:"required,min=1,max=4000"`
}
