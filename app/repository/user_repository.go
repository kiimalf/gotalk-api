package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gotalk-api/app/model"
)

var (
	ErrNotFound  = errors.New("Data tidak ditemukan")
	ErrDuplicate = errors.New("Data sudah ada")
)

type UserRepository interface {
	FindAll(ctx context.Context, q model.ListQuery) ([]model.User, int, error)
	FindAfterCursor(ctx context.Context, q model.CursorQuery) ([]model.User, error)
	FindByID(ctx context.Context, id int) (model.User, error)
	FindByUsername(ctx context.Context, username string) (model.User, error)
	FindByEmail(ctx context.Context, email string) (model.User, error)
	SearchOtherUsers(ctx context.Context, currentUserID int, search string, limit int) ([]model.UserPublicResponse, error)
	Create(ctx context.Context, u model.User) (model.User, error)
	Update(ctx context.Context, u model.User) (model.User, error)
	UpdateRole(ctx context.Context, id int, role string) (model.User, error)
	Delete(ctx context.Context, id int) error
}

type userPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userPostgresRepository{pool: pool}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func buildFilter(search string, isActive *bool) (string, []any) {
	where := " WHERE 1 = 1"
	args := []any{}

	if search != "" {
		where += fmt.Sprintf(" AND (username ILIKE $%d OR email ILIKE $%d)",
			len(args)+1, len(args)+1)
		args = append(args, "%"+search+"%")
	}

	if isActive != nil {
		where += fmt.Sprintf(" AND is_active = $%d", len(args)+1)
		args = append(args, *isActive)
	}

	return where, args
}

func (r *userPostgresRepository) FindAll(ctx context.Context, q model.ListQuery) ([]model.User, int, error) {
	where, args := buildFilter(q.Search, q.IsActive)

	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users"+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("Menghitung total user: %w", err)
	}

	query := fmt.Sprintf(
		`SELECT id, username, email, password_hash, role, is_active, created_at, updated_at
		 FROM users%s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, len(args)+1, len(args)+2,
	)
	args = append(args, q.Limit, q.Offset())

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("Mengambil daftar user: %w", err)
	}
	defer rows.Close()

	users := []model.User{}
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}

func (r *userPostgresRepository) FindAfterCursor(ctx context.Context, q model.CursorQuery) ([]model.User, error) {
	where, args := buildFilter(q.Search, q.IsActive)

	if q.After != nil {
		where += fmt.Sprintf(" AND (created_at, id) < ($%d, $%d)", len(args)+1, len(args)+2)
		args = append(args, q.After.CreatedAt, q.After.ID)
	}

	args = append(args, q.Limit+1)
	query := fmt.Sprintf(
		`SELECT id, username, email, password_hash, role, is_active, created_at, updated_at
		 FROM users%s ORDER BY created_at DESC, id DESC LIMIT $%d`,
		where, len(args),
	)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("Mengambil cursor user: %w", err)
	}
	defer rows.Close()

	users := []model.User{}
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *userPostgresRepository) FindByID(ctx context.Context, id int) (model.User, error) {
	var u model.User
	query := `SELECT id, username, email, password_hash, role, is_active, created_at, updated_at
	          FROM users WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("Mengambil user by id: %w", err)
	}
	return u, nil
}

func (r *userPostgresRepository) FindByUsername(ctx context.Context, username string) (model.User, error) {
	var u model.User
	query := `SELECT id, username, email, password_hash, role, is_active, created_at, updated_at
	          FROM users WHERE LOWER(username) = LOWER($1)`
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("Mengambil user by username: %w", err)
	}
	return u, nil
}

func (r *userPostgresRepository) FindByEmail(ctx context.Context, email string) (model.User, error) {
	var u model.User
	query := `SELECT id, username, email, password_hash, role, is_active, created_at, updated_at
	          FROM users WHERE LOWER(email) = LOWER($1)`
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("Mengambil user by email: %w", err)
	}
	return u, nil
}

func (r *userPostgresRepository) SearchOtherUsers(ctx context.Context, currentUserID int, search string, limit int) ([]model.UserPublicResponse, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	query := `SELECT id, username, created_at FROM users
	          WHERE id != $1 AND is_active = true AND username ILIKE $2
	          ORDER BY username ASC LIMIT $3`
	rows, err := r.pool.Query(ctx, query, currentUserID, "%"+search+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("Mencari user publik: %w", err)
	}
	defer rows.Close()

	result := []model.UserPublicResponse{}
	for rows.Next() {
		var u model.UserPublicResponse
		if err := rows.Scan(&u.ID, &u.Username, &u.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, u)
	}
	return result, rows.Err()
}

func (r *userPostgresRepository) Create(ctx context.Context, u model.User) (model.User, error) {
	query := `INSERT INTO users (username, email, password_hash, role, is_active)
	          VALUES ($1, $2, $3, $4, $5)
	          RETURNING id, created_at, updated_at`
	err := r.pool.QueryRow(ctx, query, u.Username, u.Email, u.Password, u.Role, u.IsActive).
		Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return model.User{}, ErrDuplicate
		}
		return model.User{}, fmt.Errorf("Menyimpan user: %w", err)
	}
	return u, nil
}

func (r *userPostgresRepository) Update(ctx context.Context, u model.User) (model.User, error) {
	query := `UPDATE users SET username = $1, email = $2, role = $3, is_active = $4, updated_at = NOW()
	          WHERE id = $5
	          RETURNING id, username, email, password_hash, role, is_active, created_at, updated_at`
	err := r.pool.QueryRow(ctx, query, u.Username, u.Email, u.Role, u.IsActive, u.ID).Scan(
		&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return model.User{}, ErrDuplicate
		}
		return model.User{}, fmt.Errorf("Memperbarui user: %w", err)
	}
	return u, nil
}

func (r *userPostgresRepository) UpdateRole(ctx context.Context, id int, role string) (model.User, error) {
	var u model.User
	query := `UPDATE users SET role = $1, updated_at = NOW()
	          WHERE id = $2
	          RETURNING id, username, email, password_hash, role, is_active, created_at, updated_at`
	err := r.pool.QueryRow(ctx, query, role, id).Scan(
		&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("Mengubah role user: %w", err)
	}
	return u, nil
}

func (r *userPostgresRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM users where id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("Menghapus user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
