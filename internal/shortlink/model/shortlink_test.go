package model

import (
	"testing"
	"time"
)

func TestToResponsesKeepsExactItems(t *testing.T) {
	links := []ShortLinkWithStats{
		{ShortLink: ShortLink{ID: 1, ShortCode: "first"}},
		{ShortLink: ShortLink{ID: 2, ShortCode: "second"}},
	}

	responses := ToResponses(links)
	if len(responses) != 2 {
		t.Fatalf("got %d responses, want 2", len(responses))
	}
	if responses[0].ID != 1 || responses[1].ID != 2 {
		t.Fatalf("unexpected responses: %#v", responses)
	}
}

func TestLinkResponseWithoutExpiry(t *testing.T) {
	link := ShortLink{ID: 1, ShortCode: "plain", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	response := link.ToResponse(0)
	if response.ExpiresAt != nil {
		t.Fatalf("expected no expiry, got %v", response.ExpiresAt)
	}
}
