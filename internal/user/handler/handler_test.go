package handler

import (
	"context"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/example/shortlink-api/internal/user/model"
	userservice "github.com/example/shortlink-api/internal/user/service"
)

type contextRepository struct {
	sawCanceled atomic.Bool
}

func (r *contextRepository) Create(context.Context, string, string) (model.User, error) {
	return model.User{}, nil
}
func (r *contextRepository) FindByUsername(ctx context.Context, _ string) (model.User, error) {
	if ctx.Err() != nil {
		r.sawCanceled.Store(true)
	}
	return model.User{}, ctx.Err()
}
func (r *contextRepository) FindByID(context.Context, int64) (model.User, error) {
	return model.User{}, nil
}

func TestLoginForwardsCanceledRequestContext(t *testing.T) {
	repo := &contextRepository{}
	handler := NewHandler(userservice.NewService(repo), "secret", 1)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	request := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(`{"username":"alice","password":"password"}`)).WithContext(ctx)
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(httptest.NewRecorder(), request)

	if !repo.sawCanceled.Load() {
		t.Fatal("repository did not receive the canceled request context")
	}
}
