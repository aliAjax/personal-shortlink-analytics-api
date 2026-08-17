package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/example/shortlink-api/internal/shortlink/model"
	"github.com/example/shortlink-api/internal/shortlink/repository"
)

type fakeRepository struct {
	listErr   error
	createErr error
	getErr    error
	deleteErr error
}

func (f *fakeRepository) Create(_ context.Context, link model.ShortLink) (model.ShortLink, error) {
	if f.createErr != nil {
		return model.ShortLink{}, f.createErr
	}
	return link, nil
}
func (f *fakeRepository) FindByCode(context.Context, string) (model.ShortLink, error) {
	return model.ShortLink{}, repository.ErrNotFound
}
func (f *fakeRepository) FindByIDAndUser(context.Context, int64, int64) (model.ShortLink, error) {
	return model.ShortLink{}, f.getErr
}
func (f *fakeRepository) ListByUser(_ context.Context, userID int64) ([]model.ShortLinkWithStats, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []model.ShortLinkWithStats{{ShortLink: model.ShortLink{ID: userID, ShortCode: fmt.Sprintf("code-%d", userID)}}}, nil
}
func (f *fakeRepository) DeleteByIDAndUser(context.Context, int64, int64) (bool, error) {
	return false, f.deleteErr
}

func TestListKeepsExactRows(t *testing.T) {
	links, err := NewService(&fakeRepository{}).List(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || links[0].ID != 7 {
		t.Fatalf("unexpected links: %#v", links)
	}
}

func TestServicePreservesRepositoryErrors(t *testing.T) {
	sentinel := errors.New("database offline")
	service := NewService(&fakeRepository{listErr: sentinel, createErr: repository.ErrDuplicateCode, getErr: sentinel, deleteErr: sentinel})
	if _, err := service.List(context.Background(), 1); !errors.Is(err, sentinel) {
		t.Fatalf("list error lost identity: %v", err)
	}
	if _, err := service.Get(context.Background(), 1, 1); !errors.Is(err, sentinel) {
		t.Fatalf("get error lost identity: %v", err)
	}
	if err := service.Delete(context.Background(), 1, 1); !errors.Is(err, sentinel) {
		t.Fatalf("delete error lost identity: %v", err)
	}
	if _, err := service.Create(context.Background(), 1, model.CreateRequest{OriginalURL: "https://example.com", CustomAlias: "taken"}); !errors.Is(err, ErrDuplicateCode) {
		t.Fatalf("duplicate error lost identity: %v", err)
	}
}

func TestConcurrentListsStayIsolated(t *testing.T) {
	service := NewService(&fakeRepository{})
	var wg sync.WaitGroup
	errCh := make(chan error, 64)
	for userID := int64(1); userID <= 64; userID++ {
		userID := userID
		wg.Add(1)
		go func() {
			defer wg.Done()
			links, err := service.List(context.Background(), userID)
			if err != nil {
				errCh <- err
				return
			}
			if len(links) != 1 || links[0].ID != userID {
				errCh <- fmt.Errorf("user %d got %#v", userID, links)
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Error(err)
	}
}

func TestIsExpiredWithoutExpiry(t *testing.T) {
	if IsExpired(model.ShortLink{}, time.Now()) {
		t.Fatal("link without expiry reported expired")
	}
}
