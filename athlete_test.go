package espn

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// rewriteTransport redirects every request to target, regardless of the
// host/scheme baked into the request URL. Lets GetPlayer's hardcoded
// athleteURL constant be tested against an httptest.Server without
// changing how the URL is built.
type rewriteTransport struct {
	target *url.URL
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = t.target.Scheme
	req.URL.Host = t.target.Host
	return http.DefaultTransport.RoundTrip(req)
}

func TestGetPlayer_Mocked(t *testing.T) {
	const body = `{
		"athlete": {
			"id": "9999",
			"fullName": "Test Player",
			"displayName": "Test Player",
			"position": {"abbreviation": "QB"},
			"team": {"abbreviation": "TST"}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantPath := "/apis/common/v3/sports/football/nfl/athletes/9999"
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

	athlete, err := GetPlayer(client, Football, NFL, "9999")
	if err != nil {
		t.Fatalf("GetPlayer returned error: %v", err)
	}

	if athlete.ID != "9999" {
		t.Errorf("expected ID %q, got %q", "9999", athlete.ID)
	}
	if athlete.FullName != "Test Player" {
		t.Errorf("expected FullName %q, got %q", "Test Player", athlete.FullName)
	}
	if athlete.Position.Abbreviation != "QB" {
		t.Errorf("expected position %q, got %q", "QB", athlete.Position.Abbreviation)
	}
}

// TestGetPlayer_Live hits the real ESPN API. Slow, network-dependent, and
// will break if ESPN changes the response shape — skipped with -short.
func TestGetPlayer_Live(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live network test in -short mode")
	}

	athlete, err := GetPlayer(nil, Football, NFL, "3139477")
	if err != nil {
		t.Fatalf("GetPlayer returned error: %v", err)
	}

	if athlete.ID != "3139477" {
		t.Errorf("expected ID %q, got %q", "3139477", athlete.ID)
	}
	if athlete.FullName == "" {
		t.Error("expected FullName to be populated, got empty string")
	}
}
