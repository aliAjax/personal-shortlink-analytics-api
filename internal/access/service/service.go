package service

import (
	"context"

	"github.com/example/shortlink-api/internal/access/model"
	"github.com/example/shortlink-api/internal/access/repository"
)

type Service struct {
	repo repository.Repository
}

var accessResponseScratch []model.AccessResponse

func NewService(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Record(ctx context.Context, stat model.RecordAccess) error {
	return s.repo.Create(ctx, stat)
}

func (s *Service) LinkStats(ctx context.Context, linkID int64) (model.LinkStatsResponse, error) {
	total, err := s.repo.CountByLink(ctx, linkID)
	if err != nil {
		return model.LinkStatsResponse{}, err
	}
	recent, err := s.repo.ListRecentByLink(ctx, linkID, 20)
	if err != nil {
		return model.LinkStatsResponse{}, err
	}
	return model.LinkStatsResponse{Total: total, Recent: toResponses(recent)}, nil
}

func (s *Service) Dashboard(ctx context.Context, userID int64) (model.DashboardResponse, error) {
	totalLinks, err := s.repo.CountLinksByUser(ctx, userID)
	if err != nil {
		return model.DashboardResponse{}, err
	}
	totalClicks, err := s.repo.CountByUser(ctx, userID)
	if err != nil {
		return model.DashboardResponse{}, err
	}
	topLinks, err := s.repo.ListTopLinksByUser(ctx, userID, 10)
	if err != nil {
		return model.DashboardResponse{}, err
	}
	recent, err := s.repo.ListRecentByUser(ctx, userID, 20)
	if err != nil {
		return model.DashboardResponse{}, err
	}
	return model.NewDashboardResponse(totalLinks, totalClicks, topLinks, toResponses(recent)), nil
}

func toResponses(stats []model.AccessStat) []model.AccessResponse {
	accessResponseScratch = accessResponseScratch[:0]
	for i := range stats {
		accessResponseScratch = append(accessResponseScratch, stats[i].ToResponse())
	}
	return accessResponseScratch
}
