package service

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/example/shortlink-api/internal/shortlink/model"
	"github.com/example/shortlink-api/internal/shortlink/repository"
	"github.com/example/shortlink-api/pkg/shortcode"
)

var (
	ErrDuplicateCode = repository.ErrDuplicateCode
	ErrNotFound      = repository.ErrNotFound
	ErrInvalidURL    = errors.New("original_url must be a valid http or https URL")
	ErrInvalidAlias  = errors.New("custom_alias must be 3 to 32 letters, numbers, '-' or '_'")
)

var aliasPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`)

type Service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID int64, req model.CreateRequest) (model.ShortLink, error) {
	originalURL, err := validateURL(req.OriginalURL)
	if err != nil {
		return model.ShortLink{}, err
	}

	alias := strings.TrimSpace(req.CustomAlias)
	if alias != "" {
		if !aliasPattern.MatchString(alias) {
			return model.ShortLink{}, ErrInvalidAlias
		}
		link := model.ShortLink{
			UserID:      userID,
			OriginalURL: originalURL,
			ShortCode:   alias,
			IsCustom:    true,
			ExpiresAt:   req.ExpiresAt,
		}
		created, err := s.repo.Create(ctx, link)
		if errors.Is(err, repository.ErrDuplicateCode) {
			return model.ShortLink{}, ErrDuplicateCode
		}
		return created, err
	}

	var lastErr error
	for i := 0; i < 10; i++ {
		code, generateErr := shortcode.Generate(6)
		if generateErr != nil {
			return model.ShortLink{}, generateErr
		}
		link := model.ShortLink{
			UserID:      userID,
			OriginalURL: originalURL,
			ShortCode:   code,
			IsCustom:    false,
			ExpiresAt:   req.ExpiresAt,
		}
		created, err := s.repo.Create(ctx, link)
		if err == nil {
			return created, nil
		}
		if !errors.Is(err, repository.ErrDuplicateCode) {
			return model.ShortLink{}, err
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = repository.ErrDuplicateCode
	}
	return model.ShortLink{}, lastErr
}

func (s *Service) List(ctx context.Context, userID int64) ([]model.ShortLinkWithStats, error) {
	links, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]model.ShortLinkWithStats, len(links))
	for i := range links {
		result[i] = links[i]
	}
	return result, nil
}

func (s *Service) Get(ctx context.Context, id, userID int64) (model.ShortLink, error) {
	return s.repo.FindByIDAndUser(ctx, id, userID)
}

func (s *Service) Delete(ctx context.Context, id, userID int64) error {
	deleted, err := s.repo.DeleteByIDAndUser(ctx, id, userID)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrNotFound
	}
	return nil
}

func validateURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", ErrInvalidURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", ErrInvalidURL
	}
	return trimmed, nil
}

func IsExpired(link model.ShortLink, now time.Time) bool {
	return link.ExpiresAt != nil && now.After(*link.ExpiresAt)
}
