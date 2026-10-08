package service

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"

	"gotalk-api/app/model"
	"gotalk-api/app/repository"
	"gotalk-api/helper"
)

type ChatService struct {
	chats    repository.ChatRepository
	messages repository.MessageRepository
	users    repository.UserRepository
}

func NewChatService(
	chats repository.ChatRepository,
	messages repository.MessageRepository,
	users repository.UserRepository,
) *ChatService {
	return &ChatService{chats: chats, messages: messages, users: users}
}

func (s *ChatService) CreateOrGetRoom(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("Belum terautentikasi")
	}

	var req model.CreateChatRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("Body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	if current.UserID == req.RecipientID {
		return helper.BadRequest("Tidak dapat membuat obrolan dengan diri sendiri")
	}

	recipient, err := s.users.FindByID(ctx, req.RecipientID)
	if err != nil || !recipient.IsActive {
		return helper.NotFound("Pengguna tujuan tidak ditemukan atau tidak aktif")
	}

	room, created, err := s.chats.FindOrCreateRoom(ctx, current.UserID, req.RecipientID)
	if err != nil {
		return helper.Internal(err)
	}

	res := model.ChatRoomResponse{
		ID: room.ID,
		Partner: &model.UserPublicResponse{
			ID:        recipient.ID,
			Username:  recipient.Username,
			CreatedAt: recipient.CreatedAt,
		},
		CreatedAt: room.CreatedAt,
		UpdatedAt: room.UpdatedAt,
	}

	if created {
		return helper.Created(c, "Ruang obrolan berhasil dibuat", res, fmt.Sprintf("/api/v1/chats/%d", room.ID))
	}
	return helper.Success(c, fiber.StatusOK, "Ruang obrolan ditemukan", res)
}

func (s *ChatService) ListRooms(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("Belum terautentikasi")
	}

	rooms, err := s.chats.FindUserRooms(ctx, current.UserID)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "Daftar obrolan berhasil diambil", rooms)
}

func (s *ChatService) SendMessage(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("Belum terautentikasi")
	}

	roomID, ok := helper.ParamId64(c)
	if !ok {
		return helper.BadRequest("ID room tidak valid")
	}

	isMember, err := s.chats.IsParticipant(ctx, roomID, current.UserID)
	if err != nil || !isMember {
		return helper.Forbidden("Anda bukan partisipan dalam ruang obrolan ini")
	}

	var req model.SendMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("Body harus berupa JSON yang valid")
	}

	req.Content = strings.TrimSpace(req.Content)
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	senderID := current.UserID
	msg, err := s.messages.Create(ctx, model.Message{
		RoomID:   roomID,
		SenderID: &senderID,
		Content:  req.Content,
	})
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Created(c, "Pesan berhasil dikirim", model.MessageResponse{
		ID:        msg.ID,
		RoomID:    msg.RoomID,
		SenderID:  *msg.SenderID,
		Content:   msg.Content,
		IsEdited:  false,
		CreatedAt: msg.CreatedAt,
	}, fmt.Sprintf("/api/v1/messages/%d", msg.ID))
}

func (s *ChatService) ListMessages(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("Belum terautentikasi")
	}

	roomID, ok := helper.ParamId64(c)
	if !ok {
		return helper.BadRequest("ID room tidak valid")
	}

	isMember, err := s.chats.IsParticipant(ctx, roomID, current.UserID)
	if err != nil || !isMember {
		return helper.Forbidden("Anda bukan partisipan dalam ruang obrolan ini")
	}

	q, err := helper.ParseCursorQuery(c)
	if err != nil {
		return err
	}

	rows, err := s.messages.FindRoomMessagesCursor(ctx, roomID, q)
	if err != nil {
		return helper.Internal(err)
	}

	hasMore := len(rows) > q.Limit
	if hasMore {
		rows = rows[:q.Limit]
	}

	meta := &model.CursorMeta{
		Limit:   q.Limit,
		HasMore: hasMore,
	}

	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		meta.NextCursor = helper.EncodeCursor(last.CreatedAt, last.ID)
	}

	responses := make([]model.MessageResponse, len(rows))
	for i, r := range rows {
		responses[i] = model.MessageResponse{
			ID:        r.ID,
			RoomID:    r.RoomID,
			SenderID:  *r.SenderID,
			Content:   r.Content,
			IsEdited:  r.UpdatedAt.After(r.CreatedAt),
			CreatedAt: r.CreatedAt,
		}
	}

	return helper.SuccessCursor(c, "Pesan obrolan berhasil diambil", responses, meta)
}

func (s *ChatService) UpdateMessage(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("Belum terautentikasi")
	}

	msgID, ok := helper.ParamId64(c)
	if !ok {
		return helper.BadRequest("ID pesan tidak valid")
	}

	existing, err := s.messages.FindByID(ctx, msgID)
	if err != nil {
		return helper.NotFound("Pesan tidak ditemukan")
	}

	if existing.DeletedAt != nil {
		return helper.NotFound("Pesan telah dihapus")
	}

	if existing.SenderID == nil || *existing.SenderID != current.UserID {
		return helper.Forbidden("Anda tidak berhak mengedit pesan milik orang lain")
	}

	var req model.UpdateMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("Body harus berupa JSON yang valid")
	}

	req.Content = strings.TrimSpace(req.Content)
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	updated, err := s.messages.Update(ctx, msgID, req.Content)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "Pesan berhasil diperbarui", model.MessageResponse{
		ID:        updated.ID,
		RoomID:    updated.RoomID,
		SenderID:  *updated.SenderID,
		Content:   updated.Content,
		IsEdited:  true,
		CreatedAt: updated.CreatedAt,
	})
}

func (s *ChatService) DeleteMessage(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("Belum terautentikasi")
	}

	msgID, ok := helper.ParamId64(c)
	if !ok {
		return helper.BadRequest("ID pesan tidak valid")
	}

	existing, err := s.messages.FindByID(ctx, msgID)
	if err != nil {
		return helper.NotFound("Pesan tidak ditemukan")
	}

	if existing.SenderID == nil || *existing.SenderID != current.UserID {
		return helper.Forbidden("Anda tidak berhak menghapus pesan milik orang lain")
	}

	if err := s.messages.SoftDelete(ctx, msgID); err != nil {
		return helper.Internal(err)
	}

	return helper.NoContent(c)
}
