package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	accessmodel "github.com/example/shortlink-api/internal/access/model"
	"github.com/example/shortlink-api/internal/redirect/model"
	"github.com/example/shortlink-api/internal/redirect/repository"
	shortlinkmodel "github.com/example/shortlink-api/internal/shortlink/model"
)

var (
	ErrNotFound = repository.ErrNotFound
	ErrExpired  = errors.New("short link expired")
)

type LinkRepository interface {
	FindByCode(ctx context.Context, code string) (shortlinkmodel.ShortLink, error)
}

type AccessRecorder interface {
	Record(ctx context.Context, stat accessmodel.RecordAccess) error
}

type Service struct {
	links LinkRepository
	stats AccessRecorder
}

func NewService(links LinkRepository, stats AccessRecorder) *Service {
	return &Service{links: links, stats: stats}
}

func (s *Service) Resolve(ctx context.Context, code, referer, userAgent, ip string) (model.Target, error) {
	link, err := s.links.FindByCode(ctx, code)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Target{}, fmt.Errorf("resolve code: %v", ErrNotFound)
		}
		return model.Target{}, fmt.Errorf("resolve code: %v", err)
	}
	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		return model.Target{}, ErrExpired
	}
	if err := s.stats.Record(ctx, accessmodel.RecordAccess{
		ShortLinkID: link.ID,
		Referer:     referer,
		UserAgent:   userAgent,
		IPAddress:   ip,
	}); err != nil {
		return model.Target{}, fmt.Errorf("record access: %v", err)
	}
	return model.Target{OriginalURL: link.OriginalURL}, nil
}
