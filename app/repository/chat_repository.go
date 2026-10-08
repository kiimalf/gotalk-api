package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gotalk-api/app/model"
)

type ChatRepository interface {
	FindOrCreateRoom(ctx context.Context, uA, uB int) (model.ChatRoom, bool, error)
	FindUserRooms(ctx context.Context, userID int) ([]model.ChatRoomResponse, error)
	IsParticipant(ctx context.Context, roomID int64, userID int) (bool, error)
}

type chatPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewChatRepository(pool *pgxpool.Pool) ChatRepository {
	return &chatPostgresRepository{pool: pool}
}

func CanonicalOrdering(userA, userB int) (int, int) {
	if userA < userB {
		return userA, userB
	}
	return userB, userA
}

func (r *chatPostgresRepository) FindOrCreateRoom(
	ctx context.Context, userA, userB int,
) (model.ChatRoom, bool, error) {
	u1, u2 := CanonicalOrdering(userA, userB)
	var room model.ChatRoom

	err := r.pool.QueryRow(ctx,
		`SELECT id, user1_id, user2_id, created_at, updated_at
	    FROM chat_rooms WHERE user1_id = $1 AND user2_id = $2`, u1, u2,
	).Scan(&room.ID, &room.User1ID, &room.User2ID, &room.CreatedAt, &room.UpdatedAt)

	if err == nil {
		return room, false, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return model.ChatRoom{}, false, fmt.Errorf("Mencari roomchat: %w", err)
	}

	err = r.pool.QueryRow(ctx,
		`INSERT INTO chat_rooms (user1_id, user2_id)
		VALUES ($1, $2)
		ON CONFLICT (user1_id, user2_id) DO UPDATE SET updated_at = NOW()
		RETURNING id, user1_id, user2_id, created_at, updated_at`, u1, u2,
	).Scan(&room.ID, &room.User1ID, &room.User2ID, &room.CreatedAt, &room.UpdatedAt)

	if err != nil {
		return model.ChatRoom{}, false, fmt.Errorf("Membuat room: %w", err)
	}

	return room, true, nil
}

func (r *chatPostgresRepository) FindUserRooms(ctx context.Context, userID int) ([]model.ChatRoomResponse, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT c.id, c.created_at, c.updated_at, u.id, u.username, u.created_at
		FROM chat_rooms c
		LEFT JOIN users u ON (u.id = CASE WHEN c.user1_id = $1 THEN c.user2_id ELSE c.user1_id END)
		WHERE c.user1_id = $1 OR c.user2_id = $1
		ORDER BY c.updated_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("Mengambil daftar room: %w", err)
	}
	defer rows.Close()

	results := []model.ChatRoomResponse{}
	for rows.Next() {
		var item model.ChatRoomResponse
		var partnerID *int
		var partnerUsername *string
		var partnerCreatedAt *time.Time
		if err := rows.Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt,
			&partnerID, &partnerUsername, &partnerCreatedAt); err != nil {
			return nil, err
		}
		if partnerID != nil && partnerUsername != nil && partnerCreatedAt != nil {
			item.Partner = &model.UserPublicResponse{
				ID:        *partnerID,
				Username:  *partnerUsername,
				CreatedAt: *partnerCreatedAt,
			}
		} else {
			item.Partner = &model.UserPublicResponse{
				ID:       0,
				Username: "Pengguna Dihapus",
			}
		}
		results = append(results, item)
	}
	return results, rows.Err()
}

func (r *chatPostgresRepository) IsParticipant(ctx context.Context, roomID int64, userID int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(
	            SELECT 1 FROM chat_rooms WHERE id = $1 AND (user1_id = $2 OR user2_id = $2)
	          )`
	err := r.pool.QueryRow(ctx, query, roomID, userID).Scan(&exists)
	return exists, err
}
