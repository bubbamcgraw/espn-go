package espn

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetDraftPool_Mocked(t *testing.T) {
	const body = `[
		{"id": 4429795, "fullName": "Jahmyr Gibbs", "proTeamId": 8, "defaultPositionId": 2, "draftRanksByRankType": {"STANDARD": {"rank": 1}}},
		{"id": 4430807, "fullName": "Bijan Robinson", "proTeamId": 1, "defaultPositionId": 2, "draftRanksByRankType": {"STANDARD": {"rank": 2}}},
		{"id": 9999999, "fullName": "No Draft Data", "proTeamId": 28, "defaultPositionId": 1, "draftRanksByRankType": {"STANDARD": {"rank": 0}}},
		{"id": 8888888, "fullName": "No Rank At All", "proTeamId": 0, "defaultPositionId": 1, "draftRanksByRankType": {}}
	]`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantPath := "/apis/v3/games/ffl/seasons/2026/players"
		if r.URL.Path != wantPath {
			t.Errorf("unexpected request path: got %q, want %q", r.URL.Path, wantPath)
		}
		if got := r.Header.Get("x-fantasy-filter"); got == "" {
			t.Error("expected x-fantasy-filter header to be set")
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

	pool, err := GetDraftPool(client, "2026")
	if err != nil {
		t.Fatalf("GetDraftPool returned error: %v", err)
	}

	if len(pool) != 4 {
		t.Fatalf("expected 4 pool entries, got %d", len(pool))
	}

	byID := make(map[int]DraftPoolPlayer, len(pool))
	for _, p := range pool {
		byID[p.ID] = p
	}

	gibbs := byID[4429795]
	if gibbs.Rank != 1 {
		t.Errorf("expected Gibbs rank 1, got %d", gibbs.Rank)
	}
	if gibbs.Team != "DET" {
		t.Errorf("expected Gibbs team DET, got %q", gibbs.Team)
	}
	if gibbs.Position != "RB" {
		t.Errorf("expected Gibbs position RB, got %q", gibbs.Position)
	}

	robinson := byID[4430807]
	if robinson.Team != "ATL" {
		t.Errorf("expected Robinson team ATL, got %q", robinson.Team)
	}

	washington := byID[9999999]
	if washington.Team != "WAS" {
		t.Errorf("expected proTeamId 28 to map to WAS, got %q", washington.Team)
	}
	if washington.Rank != 0 {
		t.Errorf("expected rank-0 entry to carry through as 0, got %d", washington.Rank)
	}

	freeAgent := byID[8888888]
	if freeAgent.Team != "" {
		t.Errorf("expected proTeamId 0 to map to no team, got %q", freeAgent.Team)
	}
}

// TestGetDraftPool_Live hits the real ESPN API. Slow, network-dependent,
// and will break if ESPN changes the response shape — skipped with -short.
func TestGetDraftPool_Live(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live network test in -short mode")
	}

	pool, err := GetDraftPool(nil, "2026")
	if err != nil {
		t.Fatalf("GetDraftPool returned error: %v", err)
	}

	if len(pool) == 0 {
		t.Fatal("expected at least one player in the draft pool")
	}

	ranked := 0
	for _, p := range pool {
		if p.Rank > 0 {
			ranked++
		}
	}
	if ranked == 0 {
		t.Error("expected at least one player with a rank")
	}
}
