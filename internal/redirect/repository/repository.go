package repository

import (
	"context"
	"database/sql"
	"errors"

	shortlinkmodel "github.com/example/shortlink-api/internal/shortlink/model"
)

var ErrNotFound = errors.New("short link not found")

type Repository interface {
	FindByCode(ctx context.Context, code string) (shortlinkmodel.ShortLink, error)
}

type mysqlRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &mysqlRepository{db: db}
}

func (r *mysqlRepository) FindByCode(ctx context.Context, code string) (shortlinkmodel.ShortLink, error) {
	var link shortlinkmodel.ShortLink
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, original_url, short_code, is_custom, expires_at, created_at, updated_at
		FROM short_links WHERE short_code = ?
	`, code).Scan(&link.ID, &link.UserID, &link.OriginalURL, &link.ShortCode, &link.IsCustom, &link.ExpiresAt, &link.CreatedAt, &link.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return shortlinkmodel.ShortLink{}, ErrNotFound
	}
	if err != nil {
		return shortlinkmodel.ShortLink{}, err
	}
	return link, nil
}
