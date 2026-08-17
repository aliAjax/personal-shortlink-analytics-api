package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/example/shortlink-api/internal/access/model"
)

type fakeRepository struct {
	recordErr     error
	countLinksErr error
}

func (f *fakeRepository) Create(context.Context, model.RecordAccess) error  { return f.recordErr }
func (f *fakeRepository) CountByLink(context.Context, int64) (int64, error) { return 2, nil }
func (f *fakeRepository) ListRecentByLink(context.Context, int64, int) ([]model.AccessStat, error) {
	return []model.AccessStat{{ID: 1}, {ID: 2}}, nil
}
func (f *fakeRepository) ListRecentByUser(_ context.Context, userID int64, _ int) ([]model.AccessStat, error) {
	return []model.AccessStat{{ID: userID, ShortCode: fmt.Sprintf("link-%d", userID)}}, nil
}
func (f *fakeRepository) CountByUser(_ context.Context, userID int64) (int64, error) {
	return userID * 10, nil
}
func (f *fakeRepository) CountLinksByUser(_ context.Context, userID int64) (int64, error) {
	if f.countLinksErr != nil {
		return 0, f.countLinksErr
	}
	return userID, nil
}
func (f *fakeRepository) ListTopLinksByUser(_ context.Context, userID int64, _ int) ([]model.TopLink, error) {
	return []model.TopLink{{ShortCode: fmt.Sprintf("link-%d", userID)}}, nil
}

func TestDashboardKeepsExactRows(t *testing.T) {
	result, err := NewService(&fakeRepository{}).Dashboard(context.Background(), 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.TopLinks) != 1 || len(result.RecentAccess) != 1 {
		t.Fatalf("unexpected dashboard rows: %#v", result)
	}
	if result.TopLinks[0].ShortCode != "link-3" || result.RecentAccess[0].ID != 3 {
		t.Fatalf("unexpected dashboard data: %#v", result)
	}
}

func TestServicePreservesRepositoryErrors(t *testing.T) {
	sentinel := errors.New("storage unavailable")
	service := NewService(&fakeRepository{recordErr: sentinel, countLinksErr: sentinel})
	if err := service.Record(context.Background(), model.RecordAccess{}); !errors.Is(err, sentinel) {
		t.Fatalf("record error lost identity: %v", err)
	}
	if _, err := service.Dashboard(context.Background(), 1); !errors.Is(err, sentinel) {
		t.Fatalf("dashboard error lost identity: %v", err)
	}
}

func TestConcurrentDashboardResultsStayIsolated(t *testing.T) {
	service := NewService(&fakeRepository{})
	var wg sync.WaitGroup
	errCh := make(chan error, 64)
	for userID := int64(1); userID <= 64; userID++ {
		userID := userID
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := service.Dashboard(context.Background(), userID)
			if err != nil {
				errCh <- err
				return
			}
			want := fmt.Sprintf("link-%d", userID)
			if len(result.TopLinks) != 1 || result.TopLinks[0].ShortCode != want {
				errCh <- fmt.Errorf("user %d got top links %#v", userID, result.TopLinks)
			}
			if len(result.RecentAccess) != 1 || result.RecentAccess[0].ShortCode != want {
				errCh <- fmt.Errorf("user %d got recent %#v", userID, result.RecentAccess)
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Error(err)
	}
}
