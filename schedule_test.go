package espn

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetTeamSchedule_Mocked(t *testing.T) {
	const body = `{
		"events": [
			{
				"id": "401772722",
				"date": "2025-09-07T20:25Z",
				"seasonType": {"name": "Regular Season"},
				"week": {"number": 1},
				"competitions": [
					{
						"competitors": [
							{
								"homeAway": "home",
								"winner": true,
								"team": {"id": "9", "abbreviation": "GB", "displayName": "Green Bay Packers"},
								"score": {"displayValue": "27"}
							},
							{
								"homeAway": "away",
								"winner": false,
								"team": {"id": "8", "abbreviation": "DET", "displayName": "Detroit Lions"},
								"score": {"displayValue": "13"}
							}
						]
					}
				]
			},
			{
				"id": "401772723",
				"date": "2025-09-14T20:25Z",
				"seasonType": {"name": "Regular Season"},
				"week": {"number": 2},
				"competitions": [
					{
						"competitors": [
							{
								"homeAway": "home",
								"winner": false,
								"team": {"id": "8", "abbreviation": "DET", "displayName": "Detroit Lions"},
								"score": {"displayValue": ""}
							},
							{
								"homeAway": "away",
								"winner": false,
								"team": {"id": "6", "abbreviation": "DAL", "displayName": "Dallas Cowboys"},
								"score": {"displayValue": ""}
							}
						]
					}
				]
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantPath := "/apis/site/v2/sports/football/nfl/teams/8/schedule"
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

	games, err := GetTeamSchedule(client, Football, NFL, "8")
	if err != nil {
		t.Fatalf("GetTeamSchedule returned error: %v", err)
	}

	if len(games) != 2 {
		t.Fatalf("expected 2 games, got %d", len(games))
	}

	played := games[0]
	if played.OpponentAbbreviation != "GB" {
		t.Errorf("expected opponent GB, got %q", played.OpponentAbbreviation)
	}
	if played.HomeAway != "away" {
		t.Errorf("expected away, got %q", played.HomeAway)
	}
	if !played.Played {
		t.Error("expected Played true")
	}
	if played.Won {
		t.Error("expected Won false (Lions lost)")
	}
	if played.TeamScore != "13" || played.OpponentScore != "27" {
		t.Errorf("expected scores 13/27, got %s/%s", played.TeamScore, played.OpponentScore)
	}

	upcoming := games[1]
	if upcoming.Played {
		t.Error("expected Played false for a future game")
	}
	if upcoming.OpponentAbbreviation != "DAL" {
		t.Errorf("expected opponent DAL, got %q", upcoming.OpponentAbbreviation)
	}
}

func TestTeamID(t *testing.T) {
	id, ok := TeamID("DET")
	if !ok || id != 8 {
		t.Errorf("expected DET -> 8, got %d ok=%v", id, ok)
	}
	if id, ok := TeamID("sf"); !ok || id != 25 {
		t.Errorf("expected case-insensitive sf -> 25, got %d ok=%v", id, ok)
	}
	if _, ok := TeamID("XXX"); ok {
		t.Error("expected unknown abbreviation to miss")
	}
}

// TestGetTeamSchedule_Live hits the real ESPN API. Slow, network-dependent,
// and will break if ESPN changes the response shape — skipped with -short.
func TestGetTeamSchedule_Live(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live network test in -short mode")
	}

	games, err := GetTeamSchedule(nil, Football, NFL, "8")
	if err != nil {
		t.Fatalf("GetTeamSchedule returned error: %v", err)
	}

	if len(games) == 0 {
		t.Error("expected at least one game")
	}
}
