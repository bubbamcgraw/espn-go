package espn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// RosterAthlete is one player on a team's current roster (see
// GetTeamRoster). DateOfBirth is trimmed to YYYY-MM-DD; ESPN sends a full
// timestamp.
type RosterAthlete struct {
	ID          string
	FullName    string
	Position    string // abbreviation, e.g. "RB"
	Jersey      string
	DateOfBirth string
}

type rosterResponse struct {
	Athletes []struct {
		Items []struct {
			ID          string `json:"id"`
			FullName    string `json:"fullName"`
			Jersey      string `json:"jersey"`
			DateOfBirth string `json:"dateOfBirth"`
			Position    struct {
				Abbreviation string `json:"abbreviation"`
			} `json:"position"`
		} `json:"items"`
	} `json:"athletes"`
}

// GetTeamRoster fetches a team's current roster, flattened across ESPN's
// offense/defense/special teams groups. Useful for mapping another
// provider's players to ESPN athlete IDs by name, team and birth date.
func GetTeamRoster(client *http.Client, sport Sport, league League, teamID string) ([]RosterAthlete, error) {
	if client == nil {
		client = http.DefaultClient
	}

	url := strings.NewReplacer(
		"{sport}", string(sport),
		"{league}", string(league),
		"{teamId}", teamID,
	).Replace(teamRosterURL)

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching roster for team %s: %w", teamID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d fetching roster for team %s", resp.StatusCode, teamID)
	}

	var raw rosterResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding roster for team %s: %w", teamID, err)
	}

	var out []RosterAthlete
	for _, group := range raw.Athletes {
		for _, a := range group.Items {
			out = append(out, RosterAthlete{
				ID:          a.ID,
				FullName:    a.FullName,
				Position:    a.Position.Abbreviation,
				Jersey:      a.Jersey,
				DateOfBirth: a.DateOfBirth[:min(len(a.DateOfBirth), 10)],
			})
		}
	}
	return out, nil
}
