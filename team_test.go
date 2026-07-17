package espn

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetTeam_Mocked(t *testing.T) {
	const body = `{
		"team": {
			"id": "12",
			"location": "Kansas City",
			"name": "Chiefs",
			"abbreviation": "KC",
			"displayName": "Kansas City Chiefs"
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantPath := "/apis/site/v2/sports/football/nfl/teams/12"
		if r.URL.Path != wantPath {
			t.Errorf("unexpected request path: got %q, want %q", r.URL.Path, wantPath)
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

	team, err := GetTeam(client, Football, NFL, "12")
	if err != nil {
		t.Fatalf("GetTeam returned error: %v", err)
	}

	if team.ID != "12" {
		t.Errorf("expected ID %q, got %q", "12", team.ID)
	}
	if team.DisplayName != "Kansas City Chiefs" {
		t.Errorf("expected DisplayName %q, got %q", "Kansas City Chiefs", team.DisplayName)
	}
}

// TestGetTeam_Live hits the real ESPN API. Slow, network-dependent, and
// will break if ESPN changes the response shape — skipped with -short.
func TestGetTeam_Live(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live network test in -short mode")
	}

	team, err := GetTeam(nil, Football, NFL, "12")
	if err != nil {
		t.Fatalf("GetTeam returned error: %v", err)
	}

	if team.ID != "12" {
		t.Errorf("expected ID %q, got %q", "12", team.ID)
	}
	if team.DisplayName == "" {
		t.Error("expected DisplayName to be populated, got empty string")
	}
}
