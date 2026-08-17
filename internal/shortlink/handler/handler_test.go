package handler

import (
	"context"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/example/shortlink-api/internal/shortlink/model"
	shortlinkservice "github.com/example/shortlink-api/internal/shortlink/service"
)

type contextRepository struct {
	sawCanceled atomic.Bool
	links       []model.ShortLinkWithStats
}

func (r *contextRepository) Create(context.Context, model.ShortLink) (model.ShortLink, error) {
	return model.ShortLink{}, nil
}
func (r *contextRepository) FindByCode(context.Context, string) (model.ShortLink, error) {
	return model.ShortLink{}, nil
}
func (r *contextRepository) FindByIDAndUser(context.Context, int64, int64) (model.ShortLink, error) {
	return model.ShortLink{}, nil
}
func (r *contextRepository) ListByUser(ctx context.Context, _ int64) ([]model.ShortLinkWithStats, error) {
	if ctx.Err() != nil {
		r.sawCanceled.Store(true)
	}
	return r.links, ctx.Err()
}
func (r *contextRepository) DeleteByIDAndUser(context.Context, int64, int64) (bool, error) {
	return false, nil
}

func TestListForwardsCanceledRequestContext(t *testing.T) {
	repo := &contextRepository{}
	handler := NewHandler(shortlinkservice.NewService(repo))
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	request := httptest.NewRequest("GET", "/links/", nil).WithContext(ctx)
	router.ServeHTTP(httptest.NewRecorder(), request)

	if !repo.sawCanceled.Load() {
		t.Fatal("repository did not receive the canceled request context")
	}
}

func TestListWithoutExpiryReturnsOneItem(t *testing.T) {
	repo := &contextRepository{links: []model.ShortLinkWithStats{{ShortLink: model.ShortLink{ID: 1, ShortCode: "plain"}}}}
	handler := NewHandler(shortlinkservice.NewService(repo))
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest("GET", "/links/", nil))
	if response.Code != 200 {
		t.Fatalf("got status %d, body %s", response.Code, response.Body.String())
	}
}
