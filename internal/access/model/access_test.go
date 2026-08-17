package model

import "testing"

func TestNewDashboardResponseKeepsExactItems(t *testing.T) {
	top := []TopLink{{ShortCode: "one"}, {ShortCode: "two"}}
	recent := []AccessResponse{{ID: 1}, {ID: 2}}

	result := NewDashboardResponse(2, 7, top, recent)
	if len(result.TopLinks) != 2 || result.TopLinks[0].ShortCode != "one" {
		t.Fatalf("unexpected top links: %#v", result.TopLinks)
	}
	if len(result.RecentAccess) != 2 || result.RecentAccess[0].ID != 1 {
		t.Fatalf("unexpected recent access: %#v", result.RecentAccess)
	}

	top[0].ShortCode = "changed"
	recent[0].ID = 99
	if result.TopLinks[0].ShortCode != "one" || result.RecentAccess[0].ID != 1 {
		t.Fatal("dashboard response shares caller-owned slices")
	}
}
