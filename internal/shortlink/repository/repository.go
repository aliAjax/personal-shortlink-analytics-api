package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/example/shortlink-api/internal/shortlink/model"
)

var (
	ErrDuplicateCode = errors.New("short code already exists")
	ErrNotFound      = errors.New("short link not found")
)

type Repository interface {
	Create(ctx context.Context, link model.ShortLink) (model.ShortLink, error)
	FindByCode(ctx context.Context, code string) (model.ShortLink, error)
	FindByIDAndUser(ctx context.Context, id, userID int64) (model.ShortLink, error)
	ListByUser(ctx context.Context, userID int64) ([]model.ShortLinkWithStats, error)
	DeleteByIDAndUser(ctx context.Context, id, userID int64) (bool, error)
}

type mysqlRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &mysqlRepository{db: db}
}

func (r *mysqlRepository) Create(ctx context.Context, link model.ShortLink) (model.ShortLink, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO short_links (user_id, original_url, short_code, is_custom, expires_at)
		VALUES (?, ?, ?, ?, ?)
	`, link.UserID, link.OriginalURL, link.ShortCode, link.IsCustom, link.ExpiresAt)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return model.ShortLink{}, ErrDuplicateCode
		}
		return model.ShortLink{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return model.ShortLink{}, err
	}
	return r.FindByIDAndUser(ctx, id, link.UserID)
}

func (r *mysqlRepository) FindByCode(ctx context.Context, code string) (model.ShortLink, error) {
	var link model.ShortLink
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, original_url, short_code, is_custom, expires_at, created_at, updated_at
		FROM short_links WHERE short_code = ?
	`, code).Scan(&link.ID, &link.UserID, &link.OriginalURL, &link.ShortCode, &link.IsCustom, &link.ExpiresAt, &link.CreatedAt, &link.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.ShortLink{}, ErrNotFound
	}
	if err != nil {
		return model.ShortLink{}, err
	}
	return link, nil
}

func (r *mysqlRepository) FindByIDAndUser(ctx context.Context, id, userID int64) (model.ShortLink, error) {
	var link model.ShortLink
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, original_url, short_code, is_custom, expires_at, created_at, updated_at
		FROM short_links WHERE id = ? AND user_id = ?
	`, id, userID).Scan(&link.ID, &link.UserID, &link.OriginalURL, &link.ShortCode, &link.IsCustom, &link.ExpiresAt, &link.CreatedAt, &link.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.ShortLink{}, ErrNotFound
	}
	if err != nil {
		return model.ShortLink{}, err
	}
	return link, nil
}

func (r *mysqlRepository) ListByUser(ctx context.Context, userID int64) ([]model.ShortLinkWithStats, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT l.id, l.user_id, l.original_url, l.short_code, l.is_custom, l.expires_at, l.created_at, l.updated_at,
		       COALESCE(COUNT(a.id), 0) AS visit_count
		FROM short_links l
		LEFT JOIN access_stats a ON a.short_link_id = l.id
		WHERE l.user_id = ?
		GROUP BY l.id, l.user_id, l.original_url, l.short_code, l.is_custom, l.expires_at, l.created_at, l.updated_at
		ORDER BY l.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make([]model.ShortLinkWithStats, 0)
	for rows.Next() {
		var item model.ShortLinkWithStats
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.OriginalURL,
			&item.ShortCode,
			&item.IsCustom,
			&item.ExpiresAt,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.VisitCount,
		); err != nil {
			return nil, err
		}
		links = append(links, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return links, nil
}

func (r *mysqlRepository) DeleteByIDAndUser(ctx context.Context, id, userID int64) (bool, error) {
	result, err := r.db.ExecContext(ctx, "DELETE FROM short_links WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}
