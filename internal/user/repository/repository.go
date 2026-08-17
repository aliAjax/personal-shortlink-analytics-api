package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/example/shortlink-api/internal/user/model"
)

var (
	ErrDuplicateUsername = errors.New("username already exists")
	ErrNotFound          = errors.New("user not found")
)

type Repository interface {
	Create(ctx context.Context, username, passwordHash string) (model.User, error)
	FindByUsername(ctx context.Context, username string) (model.User, error)
	FindByID(ctx context.Context, id int64) (model.User, error)
}

type mysqlRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &mysqlRepository{db: db}
}

func (r *mysqlRepository) Create(ctx context.Context, username, passwordHash string) (model.User, error) {
	result, err := r.db.ExecContext(ctx, "INSERT INTO users (username, password_hash) VALUES (?, ?)", username, passwordHash)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return model.User{}, ErrDuplicateUsername
		}
		return model.User{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return model.User{}, err
	}
	return r.FindByID(ctx, id)
}

func (r *mysqlRepository) FindByUsername(ctx context.Context, username string) (model.User, error) {
	var user model.User
	err := r.db.QueryRowContext(ctx, "SELECT id, username, password_hash, created_at, updated_at FROM users WHERE username = ?", username).
		Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, ErrNotFound
	}
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (r *mysqlRepository) FindByID(ctx context.Context, id int64) (model.User, error) {
	var user model.User
	err := r.db.QueryRowContext(ctx, "SELECT id, username, password_hash, created_at, updated_at FROM users WHERE id = ?", id).
		Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, ErrNotFound
	}
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}
