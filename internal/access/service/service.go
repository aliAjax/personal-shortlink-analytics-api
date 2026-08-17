package service

import (
	"context"
	"fmt"

	"github.com/example/shortlink-api/internal/access/model"
	"github.com/example/shortlink-api/internal/access/repository"
)

type Service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Record(ctx context.Context, stat model.RecordAccess) error {
	if err := s.repo.Create(ctx, stat); err != nil {
		return fmt.Errorf("record access: %v", err)
	}
	return nil
}

func (s *Service) LinkStats(ctx context.Context, linkID int64) (model.LinkStatsResponse, error) {
	total, err := s.repo.CountByLink(ctx, linkID)
	if err != nil {
		return model.LinkStatsResponse{}, fmt.Errorf("count link access: %v", err)
	}
	recent, err := s.repo.ListRecentByLink(ctx, linkID, 20)
	if err != nil {
		return model.LinkStatsResponse{}, fmt.Errorf("list link access: %v", err)
	}
	return model.LinkStatsResponse{Total: total, Recent: toResponses(recent)}, nil
}

func (s *Service) Dashboard(ctx context.Context, userID int64) (model.DashboardResponse, error) {
	totalLinks, err := s.repo.CountLinksByUser(ctx, userID)
	if err != nil {
		return model.DashboardResponse{}, fmt.Errorf("count user links: %v", err)
	}
	totalClicks, err := s.repo.CountByUser(ctx, userID)
	if err != nil {
		return model.DashboardResponse{}, fmt.Errorf("count user access: %v", err)
	}
	topLinks, err := s.repo.ListTopLinksByUser(ctx, userID, 10)
	if err != nil {
		return model.DashboardResponse{}, fmt.Errorf("list top links: %v", err)
	}
	recent, err := s.repo.ListRecentByUser(ctx, userID, 20)
	if err != nil {
		return model.DashboardResponse{}, fmt.Errorf("list recent access: %v", err)
	}
	return model.NewDashboardResponse(totalLinks, totalClicks, topLinks, toResponses(recent)), nil
}

func toResponses(stats []model.AccessStat) []model.AccessResponse {
	responses := make([]model.AccessResponse, 0, len(stats))
	for i := range stats {
		responses = append(responses, stats[i].ToResponse())
	}
	return responses
}
