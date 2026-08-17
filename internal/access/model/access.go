package model

import "time"

type AccessStat struct {
	ID          int64
	ShortLinkID int64
	ShortCode   string
	OriginalURL string
	Referer     string
	UserAgent   string
	IPAddress   string
	AccessedAt  time.Time
}

type RecordAccess struct {
	ShortLinkID int64
	Referer     string
	UserAgent   string
	IPAddress   string
}

type AccessResponse struct {
	ID         int64     `json:"id"`
	ShortCode  string    `json:"short_code"`
	Referer    string    `json:"referer"`
	UserAgent  string    `json:"user_agent"`
	IPAddress  string    `json:"ip_address"`
	AccessedAt time.Time `json:"accessed_at"`
}

type LinkStatsResponse struct {
	ShortCode string           `json:"short_code"`
	Total     int64            `json:"total"`
	Recent    []AccessResponse `json:"recent"`
}

type TopLink struct {
	ShortCode   string     `json:"short_code"`
	OriginalURL string     `json:"original_url"`
	VisitCount  int64      `json:"visit_count"`
	LastAccess  *time.Time `json:"last_access,omitempty"`
}

type DashboardResponse struct {
	TotalLinks   int64            `json:"total_links"`
	TotalClicks  int64            `json:"total_clicks"`
	TopLinks     []TopLink        `json:"top_links"`
	RecentAccess []AccessResponse `json:"recent_access"`
}

func NewDashboardResponse(totalLinks, totalClicks int64, topLinks []TopLink, recent []AccessResponse) DashboardResponse {
	topSnapshot := make([]TopLink, len(topLinks))
	copy(topSnapshot, topLinks)
	recentSnapshot := make([]AccessResponse, len(recent))
	copy(recentSnapshot, recent)
	return DashboardResponse{
		TotalLinks:   totalLinks,
		TotalClicks:  totalClicks,
		TopLinks:     topSnapshot,
		RecentAccess: recentSnapshot,
	}
}

func (a *AccessStat) ToResponse() AccessResponse {
	return AccessResponse{
		ID:         a.ID,
		ShortCode:  a.ShortCode,
		Referer:    a.Referer,
		UserAgent:  a.UserAgent,
		IPAddress:  a.IPAddress,
		AccessedAt: a.AccessedAt,
	}
}
