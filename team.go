package espn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type TeamResponse struct {
	Team Team `json:"team"`
}

func GetTeam(client *http.Client, sport Sport, league League, teamID string) (Team, error) {
	if client == nil {
		client = http.DefaultClient
	}

	url := strings.NewReplacer(
		"{sport}", string(sport),
		"{league}", string(league),
		"{teamId}", teamID,
	).Replace(teamURL)

	resp, err := client.Get(url)
	if err != nil {
		return Team{}, fmt.Errorf("fetching team %s: %w", teamID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Team{}, fmt.Errorf("unexpected status %d fetching team %s", resp.StatusCode, teamID)
	}

	var out TeamResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Team{}, fmt.Errorf("decoding team %s: %w", teamID, err)
	}

	return out.Team, nil
}
