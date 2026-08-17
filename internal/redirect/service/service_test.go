package service

import (
	"context"
	"errors"
	"testing"

	accessmodel "github.com/example/shortlink-api/internal/access/model"
	shortlinkmodel "github.com/example/shortlink-api/internal/shortlink/model"
)

type fakeLinks struct {
	link shortlinkmodel.ShortLink
	err  error
}

func (f *fakeLinks) FindByCode(context.Context, string) (shortlinkmodel.ShortLink, error) {
	return f.link, f.err
}

type fakeStats struct {
	err error
}

func (f *fakeStats) Record(context.Context, accessmodel.RecordAccess) error { return f.err }

func TestResolveWithoutExpiry(t *testing.T) {
	service := NewService(&fakeLinks{link: shortlinkmodel.ShortLink{ID: 1, OriginalURL: "https://example.com"}}, &fakeStats{})
	target, err := service.Resolve(context.Background(), "plain", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if target.OriginalURL != "https://example.com" {
		t.Fatalf("unexpected target: %#v", target)
	}
}

func TestResolvePreservesDependencyErrors(t *testing.T) {
	sentinel := errors.New("dependency stopped")
	if _, err := NewService(&fakeLinks{err: sentinel}, &fakeStats{}).Resolve(context.Background(), "x", "", "", ""); !errors.Is(err, sentinel) {
		t.Fatalf("link error lost identity: %v", err)
	}
	links := &fakeLinks{link: shortlinkmodel.ShortLink{ID: 1, OriginalURL: "https://example.com"}}
	if _, err := NewService(links, &fakeStats{err: sentinel}).Resolve(context.Background(), "x", "", "", ""); !errors.Is(err, sentinel) {
		t.Fatalf("record error lost identity: %v", err)
	}
}
