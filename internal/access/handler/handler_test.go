package handler

import (
	"context"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/go-chi/chi/v5"

	accessmodel "github.com/example/shortlink-api/internal/access/model"
	accessservice "github.com/example/shortlink-api/internal/access/service"
	shortlinkmodel "github.com/example/shortlink-api/internal/shortlink/model"
	shortlinkservice "github.com/example/shortlink-api/internal/shortlink/service"
)

type contextAccessRepository struct {
	sawCanceled atomic.Bool
}

func (r *contextAccessRepository) Create(context.Context, accessmodel.RecordAccess) error { return nil }
func (r *contextAccessRepository) CountByLink(context.Context, int64) (int64, error)      { return 0, nil }
func (r *contextAccessRepository) ListRecentByLink(context.Context, int64, int) ([]accessmodel.AccessStat, error) {
	return nil, nil
}
func (r *contextAccessRepository) ListRecentByUser(context.Context, int64, int) ([]accessmodel.AccessStat, error) {
	return nil, nil
}
func (r *contextAccessRepository) CountByUser(context.Context, int64) (int64, error) { return 0, nil }
func (r *contextAccessRepository) CountLinksByUser(ctx context.Context, _ int64) (int64, error) {
	if ctx.Err() != nil {
		r.sawCanceled.Store(true)
	}
	return 0, ctx.Err()
}
func (r *contextAccessRepository) ListTopLinksByUser(context.Context, int64, int) ([]accessmodel.TopLink, error) {
	return nil, nil
}

type unusedShortlinkRepository struct{}

func (*unusedShortlinkRepository) Create(context.Context, shortlinkmodel.ShortLink) (shortlinkmodel.ShortLink, error) {
	return shortlinkmodel.ShortLink{}, nil
}
func (*unusedShortlinkRepository) FindByCode(context.Context, string) (shortlinkmodel.ShortLink, error) {
	return shortlinkmodel.ShortLink{}, nil
}
func (*unusedShortlinkRepository) FindByIDAndUser(context.Context, int64, int64) (shortlinkmodel.ShortLink, error) {
	return shortlinkmodel.ShortLink{}, nil
}
func (*unusedShortlinkRepository) ListByUser(context.Context, int64) ([]shortlinkmodel.ShortLinkWithStats, error) {
	return nil, nil
}
func (*unusedShortlinkRepository) DeleteByIDAndUser(context.Context, int64, int64) (bool, error) {
	return false, nil
}

func TestDashboardForwardsCanceledRequestContext(t *testing.T) {
	repo := &contextAccessRepository{}
	handler := NewHandler(accessservice.NewService(repo), shortlinkservice.NewService(&unusedShortlinkRepository{}))
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	request := httptest.NewRequest("GET", "/dashboard", nil).WithContext(ctx)
	router.ServeHTTP(httptest.NewRecorder(), request)

	if !repo.sawCanceled.Load() {
		t.Fatal("repository did not receive the canceled request context")
	}
}
