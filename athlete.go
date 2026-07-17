package espn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func GetPlayer(client *http.Client, sport Sport, league League, athleteID string) (Athlete, error) {
	if client == nil {
		client = http.DefaultClient
	}

	url := strings.NewReplacer(
		"{sport}", string(sport),
		"{league}", string(league),
		"{athleteId}", athleteID,
	).Replace(athleteURL)

	resp, err := client.Get(url)
	if err != nil {
		return Athlete{}, fmt.Errorf("fetching athlete %s: %w", athleteID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Athlete{}, fmt.Errorf("unexpected status %d fetching athlete %s", resp.StatusCode, athleteID)
	}

	var out AthleteResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Athlete{}, fmt.Errorf("decoding athlete %s: %w", athleteID, err)
	}

	return out.Athlete, nil
}
