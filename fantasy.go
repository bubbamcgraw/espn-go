package espn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// FantasyLeague is an ESPN Fantasy Football league. Only public leagues are
// supported — private leagues require espn_s2/SWID session cookies that
// this package does not handle.
type FantasyLeague struct {
	ID       int                   `json:"id"`
	SeasonID int                   `json:"seasonId"`
	Settings FantasyLeagueSettings `json:"settings"`
	Teams    []FantasyTeam         `json:"teams"`
	Members  []FantasyMember       `json:"members"`
}

type FantasyLeagueSettings struct {
	Name string `json:"name"`
	Size int    `json:"size"`
}

// FantasyTeam is a single franchise within a FantasyLeague.
type FantasyTeam struct {
	ID     int      `json:"id"`
	Abbrev string   `json:"abbrev"`
	Name   string   `json:"name"`
	Owners []string `json:"owners"`
}

// FantasyMember is a person who owns (or co-owns) a FantasyTeam, referenced
// by FantasyTeam.Owners.
type FantasyMember struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
}

// GetFantasyLeague fetches a public ESPN Fantasy Football league for the
// given season.
func GetFantasyLeague(client *http.Client, season string, leagueID string) (FantasyLeague, error) {
	if client == nil {
		client = http.DefaultClient
	}

	reqURL := strings.NewReplacer(
		"{season}", season,
		"{leagueId}", leagueID,
	).Replace(fantasyLeagueURL)

	q := url.Values{}
	q.Add("view", "mTeam")
	q.Add("view", "mSettings")

	resp, err := client.Get(reqURL + "?" + q.Encode())
	if err != nil {
		return FantasyLeague{}, fmt.Errorf("fetching fantasy league %s: %w", leagueID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return FantasyLeague{}, fmt.Errorf("unexpected status %d fetching fantasy league %s", resp.StatusCode, leagueID)
	}

	var out FantasyLeague
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return FantasyLeague{}, fmt.Errorf("decoding fantasy league %s: %w", leagueID, err)
	}

	return out, nil
}
