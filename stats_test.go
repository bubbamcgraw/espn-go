package espn

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetPlayerSeasonStats_Mocked(t *testing.T) {
	const body = `{
		"splits": {
			"categories": [
				{
					"name": "passing",
					"displayName": "Passing",
					"stats": [
						{"name": "passingYards", "displayName": "Passing Yards", "shortDisplayName": "YDS", "displayValue": "3,587", "value": 3587.0},
						{"name": "passingTouchdowns", "displayName": "Passing Touchdowns", "shortDisplayName": "TD", "displayValue": "22", "value": 22.0},
						{"name": "twoPtPass", "displayName": "Two Point Passes", "shortDisplayName": "2PP", "displayValue": "0", "value": 0.0}
					]
				},
				{
					"name": "defensive",
					"displayName": "Defensive",
					"stats": [
						{"name": "soloTackles", "displayName": "Solo Tackles", "shortDisplayName": "SOLO", "displayValue": "0", "value": 0.0},
						{"name": "sacks", "displayName": "Sacks", "shortDisplayName": "SACK", "displayValue": "0", "value": 0.0}
					]
				}
			]
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantPath := "/v2/sports/football/leagues/nfl/seasons/2025/types/2/athletes/3139477/statistics"
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

	stats, err := GetPlayerSeasonStats(client, Football, NFL, "3139477", 2025)
	if err != nil {
		t.Fatalf("GetPlayerSeasonStats returned error: %v", err)
	}

	if stats.Season != 2025 {
		t.Errorf("expected season 2025, got %d", stats.Season)
	}
	if len(stats.Categories) != 1 {
		t.Fatalf("expected 1 category (defensive should be dropped, all-zero), got %d: %+v", len(stats.Categories), stats.Categories)
	}
	if stats.Categories[0].Name != "passing" {
		t.Errorf("expected passing category, got %q", stats.Categories[0].Name)
	}
	if len(stats.Categories[0].Stats) != 2 {
		t.Fatalf("expected 2 stats (twoPtPass should be dropped, zero value), got %d: %+v", len(stats.Categories[0].Stats), stats.Categories[0].Stats)
	}
	yards := stats.Categories[0].Stats[0]
	if yards.Name != "passingYards" || yards.DisplayValue != "3,587" {
		t.Errorf("expected passingYards '3,587', got %+v", yards)
	}
	if yards.DisplayName != "Passing Yards" || yards.ShortDisplayName != "YDS" {
		t.Errorf("expected display labels to carry through, got %+v", yards)
	}
}

func TestGetPlayerSeasonStats_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"message":"No stats found.","code":404}}`))
	}))
	defer server.Close()

	target, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parsing server URL: %v", err)
	}
	client := &http.Client{Transport: &rewriteTransport{target: target}}

	stats, err := GetPlayerSeasonStats(client, Football, NFL, "9999999", 2018)
	if err != nil {
		t.Fatalf("expected no error for a 404 (no stats that season), got: %v", err)
	}
	if len(stats.Categories) != 0 {
		t.Errorf("expected no categories, got %d", len(stats.Categories))
	}
	if stats.Season != 2018 {
		t.Errorf("expected season 2018 to still be set, got %d", stats.Season)
	}
}

// TestGetPlayerSeasonStats_Live hits the real ESPN API. Slow,
// network-dependent, and will break if ESPN changes the response shape —
// skipped with -short.
func TestGetPlayerSeasonStats_Live(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live network test in -short mode")
	}

	stats, err := GetPlayerSeasonStats(nil, Football, NFL, "3139477", 2025)
	if err != nil {
		t.Fatalf("GetPlayerSeasonStats returned error: %v", err)
	}

	if len(stats.Categories) == 0 {
		t.Error("expected at least one stat category")
	}
}
