package espn

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetFantasyLeague_Mocked(t *testing.T) {
	const body = `{
		"id": 899513,
		"seasonId": 2023,
		"settings": {"name": "Pino Posse", "size": 10},
		"teams": [
			{"id": 1, "abbrev": "HRTS", "name": "Chubby Daddy Hurts So Good", "owners": ["{E29D59B3-F170-4B8A-9D59-B3F1702B8AD5}"]}
		],
		"members": [
			{"id": "{E29D59B3-F170-4B8A-9D59-B3F1702B8AD5}", "displayName": "someuser", "firstName": "Some", "lastName": "User"}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantPath := "/apis/v3/games/ffl/seasons/2023/segments/0/leagues/899513"
		if r.URL.Path != wantPath {
			t.Errorf("unexpected request path: got %q, want %q", r.URL.Path, wantPath)
		}
		views := r.URL.Query()["view"]
		if len(views) != 2 || views[0] != "mTeam" || views[1] != "mSettings" {
			t.Errorf("unexpected view params: got %v", views)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
	defer server.Close()

	target, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parsing server URL: %v", err)
	}
	client := &http.Client{Transport: &rewriteTransport{target: target}}

	league, err := GetFantasyLeague(client, "2023", "899513")
	if err != nil {
		t.Fatalf("GetFantasyLeague returned error: %v", err)
	}

	if league.ID != 899513 {
		t.Errorf("expected ID %d, got %d", 899513, league.ID)
	}
	if league.Settings.Name != "Pino Posse" {
		t.Errorf("expected Settings.Name %q, got %q", "Pino Posse", league.Settings.Name)
	}
	if len(league.Teams) != 1 || league.Teams[0].Name != "Chubby Daddy Hurts So Good" {
		t.Fatalf("unexpected teams: %+v", league.Teams)
	}
	if len(league.Members) != 1 || league.Members[0].DisplayName != "someuser" {
		t.Fatalf("unexpected members: %+v", league.Members)
	}
}

// TestGetFantasyLeague_Live hits the real ESPN Fantasy API against a known
// public league. Slow, network-dependent, and will break if ESPN changes
// the response shape or the league goes private — skipped with -short.
func TestGetFantasyLeague_Live(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live network test in -short mode")
	}

	league, err := GetFantasyLeague(nil, "2023", "899513")
	if err != nil {
		t.Fatalf("GetFantasyLeague returned error: %v", err)
	}

	if league.ID != 899513 {
		t.Errorf("expected ID %d, got %d", 899513, league.ID)
	}
	if league.Settings.Name == "" {
		t.Error("expected Settings.Name to be populated, got empty string")
	}
	if len(league.Teams) == 0 {
		t.Error("expected at least one team")
	}
}
