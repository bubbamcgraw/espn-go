package espn

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetTeamRoster_Mocked(t *testing.T) {
	const body = `{
		"athletes": [
			{"position": "offense", "items": [
				{"id": "4430737", "fullName": "Kyren Williams", "jersey": "23", "dateOfBirth": "2000-08-26T07:00Z", "position": {"abbreviation": "RB"}}
			]},
			{"position": "defense", "items": [
				{"id": "1", "fullName": "No Birthday", "position": {"abbreviation": "LB"}}
			]}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantPath := "/apis/site/v2/sports/football/nfl/teams/14/roster"
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

	roster, err := GetTeamRoster(client, Football, NFL, "14")
	if err != nil {
		t.Fatalf("GetTeamRoster returned error: %v", err)
	}
	if len(roster) != 2 {
		t.Fatalf("got %d athletes, want 2 (groups flattened)", len(roster))
	}
	want := RosterAthlete{ID: "4430737", FullName: "Kyren Williams", Position: "RB", Jersey: "23", DateOfBirth: "2000-08-26"}
	if roster[0] != want {
		t.Errorf("roster[0] = %+v, want %+v", roster[0], want)
	}
	if roster[1].DateOfBirth != "" {
		t.Errorf("missing birth date = %q, want empty", roster[1].DateOfBirth)
	}
}
