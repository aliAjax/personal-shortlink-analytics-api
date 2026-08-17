package handler

import (
	"context"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/go-chi/chi/v5"

	accessmodel "github.com/example/shortlink-api/internal/access/model"
	redirectservice "github.com/example/shortlink-api/internal/redirect/service"
	shortlinkmodel "github.com/example/shortlink-api/internal/shortlink/model"
)

type contextLinks struct {
	sawCanceled atomic.Bool
}

func (r *contextLinks) FindByCode(ctx context.Context, _ string) (shortlinkmodel.ShortLink, error) {
	if ctx.Err() != nil {
		r.sawCanceled.Store(true)
	}
	return shortlinkmodel.ShortLink{}, ctx.Err()
}

type unusedStats struct{}

func (*unusedStats) Record(context.Context, accessmodel.RecordAccess) error { return nil }

func TestRedirectForwardsCanceledRequestContext(t *testing.T) {
	links := &contextLinks{}
	handler := NewHandler(redirectservice.NewService(links, &unusedStats{}))
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	request := httptest.NewRequest("GET", "/r/code", nil).WithContext(ctx)
	router.ServeHTTP(httptest.NewRecorder(), request)

	if !links.sawCanceled.Load() {
		t.Fatal("repository did not receive the canceled request context")
	}
}
