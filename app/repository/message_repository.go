package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gotalk-api/app/model"
)

type MessageRepository interface {
	Create(ctx context.Context, m model.Message) (model.Message, error)
	FindRoomMessagesCursor(ctx context.Context, roomID int64, q model.CursorQuery) ([]model.Message, error)
	FindByID(ctx context.Context, id int64) (model.Message, error)
	Update(ctx context.Context, id int64, content string) (model.Message, error)
	SoftDelete(ctx context.Context, id int64) error
}

type messagePostgresRepository struct {
	pool *pgxpool.Pool
}

func NewMessageRepository(pool *pgxpool.Pool) MessageRepository {
	return &messagePostgresRepository{pool: pool}
}

func (r *messagePostgresRepository) Create(ctx context.Context, m model.Message) (model.Message, error) {
	query := `INSERT INTO messages (room_id, sender_id, content)
	          VALUES ($1, $2, $3)
	          RETURNING id, created_at, updated_at`
	err := r.pool.QueryRow(ctx, query, m.RoomID, m.SenderID, m.Content).
		Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return model.Message{}, fmt.Errorf("Menyimpan pesan: %w", err)
	}

	_, _ = r.pool.Exec(ctx, `UPDATE chat_rooms SET updated_at = NOW() WHERE id = $1`, m.RoomID)

	return m, nil
}

func (r *messagePostgresRepository) FindRoomMessagesCursor(ctx context.Context, roomID int64, q model.CursorQuery) ([]model.Message, error) {
	where := " WHERE room_id = $1 AND deleted_at IS NULL"
	args := []any{roomID}

	if q.After != nil {
		where += fmt.Sprintf(" AND (created_at, id) < ($%d, $%d)", len(args)+1, len(args)+2)
		args = append(args, q.After.CreatedAt, q.After.ID)
	}

	args = append(args, q.Limit+1)
	query := fmt.Sprintf(
		`SELECT id, room_id, sender_id, content, created_at, updated_at
		 FROM messages%s
		 ORDER BY created_at DESC, id DESC
		 LIMIT $%d`,
		where, len(args),
	)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("Mengambil riwayat pesan: %w", err)
	}
	defer rows.Close()

	messages := []model.Message{}
	for rows.Next() {
		var m model.Message
		if err := rows.Scan(&m.ID, &m.RoomID, &m.SenderID, &m.Content, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

func (r *messagePostgresRepository) FindByID(ctx context.Context, id int64) (model.Message, error) {
	var m model.Message
	query := `SELECT id, room_id, sender_id, content, created_at, updated_at, deleted_at
	          FROM messages WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&m.ID, &m.RoomID, &m.SenderID, &m.Content, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Message{}, ErrNotFound
		}
		return model.Message{}, fmt.Errorf("Mengambil pesan: %w", err)
	}
	return m, nil
}

func (r *messagePostgresRepository) Update(ctx context.Context, id int64, content string) (model.Message, error) {
	var m model.Message
	query := `UPDATE messages SET content = $1, updated_at = NOW()
	          WHERE id = $2 AND deleted_at IS NULL
	          RETURNING id, room_id, sender_id, content, created_at, updated_at`
	err := r.pool.QueryRow(ctx, query, content, id).Scan(
		&m.ID, &m.RoomID, &m.SenderID, &m.Content, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Message{}, ErrNotFound
		}
		return model.Message{}, fmt.Errorf("Memperbarui pesan: %w", err)
	}
	return m, nil
}

func (r *messagePostgresRepository) SoftDelete(ctx context.Context, id int64) error {
	query := `UPDATE messages SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("Menghapus pesan: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
