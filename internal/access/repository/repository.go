package repository

import (
	"context"
	"database/sql"

	"github.com/example/shortlink-api/internal/access/model"
)

type Repository interface {
	Create(ctx context.Context, stat model.RecordAccess) error
	CountByLink(ctx context.Context, linkID int64) (int64, error)
	ListRecentByLink(ctx context.Context, linkID int64, limit int) ([]model.AccessStat, error)
	ListRecentByUser(ctx context.Context, userID int64, limit int) ([]model.AccessStat, error)
	CountByUser(ctx context.Context, userID int64) (int64, error)
	CountLinksByUser(ctx context.Context, userID int64) (int64, error)
	ListTopLinksByUser(ctx context.Context, userID int64, limit int) ([]model.TopLink, error)
}

type mysqlRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &mysqlRepository{db: db}
}

func (r *mysqlRepository) Create(ctx context.Context, stat model.RecordAccess) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO access_stats (short_link_id, referer, user_agent, ip_address)
		VALUES (?, ?, ?, ?)
	`, stat.ShortLinkID, stat.Referer, stat.UserAgent, stat.IPAddress)
	return err
}

func (r *mysqlRepository) CountByLink(ctx context.Context, linkID int64) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM access_stats WHERE short_link_id = ?", linkID).Scan(&count)
	return count, err
}

func (r *mysqlRepository) ListRecentByLink(ctx context.Context, linkID int64, limit int) ([]model.AccessStat, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT a.id, a.short_link_id, l.short_code, l.original_url, a.referer, a.user_agent, a.ip_address, a.accessed_at
		FROM access_stats a
		JOIN short_links l ON l.id = a.short_link_id
		WHERE a.short_link_id = ?
		ORDER BY a.accessed_at DESC
		LIMIT ?
	`, linkID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make([]model.AccessStat, 0)
	for rows.Next() {
		var item model.AccessStat
		if err := rows.Scan(
			&item.ID,
			&item.ShortLinkID,
			&item.ShortCode,
			&item.OriginalURL,
			&item.Referer,
			&item.UserAgent,
			&item.IPAddress,
			&item.AccessedAt,
		); err != nil {
			return nil, err
		}
		stats = append(stats, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *mysqlRepository) ListRecentByUser(ctx context.Context, userID int64, limit int) ([]model.AccessStat, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT a.id, a.short_link_id, l.short_code, l.original_url, a.referer, a.user_agent, a.ip_address, a.accessed_at
		FROM access_stats a
		JOIN short_links l ON l.id = a.short_link_id
		WHERE l.user_id = ?
		ORDER BY a.accessed_at DESC
		LIMIT ?
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make([]model.AccessStat, 0)
	for rows.Next() {
		var item model.AccessStat
		if err := rows.Scan(
			&item.ID,
			&item.ShortLinkID,
			&item.ShortCode,
			&item.OriginalURL,
			&item.Referer,
			&item.UserAgent,
			&item.IPAddress,
			&item.AccessedAt,
		); err != nil {
			return nil, err
		}
		stats = append(stats, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *mysqlRepository) CountByUser(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM access_stats a
		JOIN short_links l ON l.id = a.short_link_id
		WHERE l.user_id = ?
	`, userID).Scan(&count)
	return count, err
}

func (r *mysqlRepository) CountLinksByUser(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM short_links WHERE user_id = ?", userID).Scan(&count)
	return count, err
}

func (r *mysqlRepository) ListTopLinksByUser(ctx context.Context, userID int64, limit int) ([]model.TopLink, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT l.short_code, l.original_url, COUNT(a.id) AS visit_count, MAX(a.accessed_at) AS last_access
		FROM short_links l
		LEFT JOIN access_stats a ON a.short_link_id = l.id
		WHERE l.user_id = ?
		GROUP BY l.short_code, l.original_url
		ORDER BY visit_count DESC, last_access DESC
		LIMIT ?
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make([]model.TopLink, 0)
	for rows.Next() {
		var item model.TopLink
		var lastAccess sql.NullTime
		if err := rows.Scan(&item.ShortCode, &item.OriginalURL, &item.VisitCount, &lastAccess); err != nil {
			return nil, err
		}
		if lastAccess.Valid {
			value := lastAccess.Time
			item.LastAccess = &value
		}
		links = append(links, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return links, nil
}
