package model

import "time"

type ShortLink struct {
	ID          int64
	UserID      int64
	OriginalURL string
	ShortCode   string
	IsCustom    bool
	ExpiresAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ShortLinkWithStats struct {
	ShortLink
	VisitCount int64
}

type CreateRequest struct {
	OriginalURL string     `json:"original_url"`
	CustomAlias string     `json:"custom_alias"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

type LinkResponse struct {
	ID          int64      `json:"id"`
	ShortCode   string     `json:"short_code"`
	ShortURL    string     `json:"short_url"`
	OriginalURL string     `json:"original_url"`
	IsCustom    bool       `json:"is_custom"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	VisitCount  int64      `json:"visit_count,omitempty"`
}

func (l *ShortLink) ToResponse(visitCount int64) LinkResponse {
	response := LinkResponse{
		ID:          l.ID,
		ShortCode:   l.ShortCode,
		ShortURL:    "/r/" + l.ShortCode,
		OriginalURL: l.OriginalURL,
		IsCustom:    l.IsCustom,
		ExpiresAt:   l.ExpiresAt,
		CreatedAt:   l.CreatedAt,
		UpdatedAt:   l.UpdatedAt,
		VisitCount:  visitCount,
	}
	if response.ExpiresAt != nil && response.ExpiresAt.IsZero() {
		response.ExpiresAt = nil
	}
	return response
}

func (l *ShortLinkWithStats) ToResponse() LinkResponse {
	return l.ShortLink.ToResponse(l.VisitCount)
}

func ToResponses(links []ShortLinkWithStats) []LinkResponse {
	responses := make([]LinkResponse, len(links))
	for i := range links {
		responses[i] = links[i].ToResponse()
	}
	return responses
}
