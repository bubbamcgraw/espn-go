package espn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// Game is a single game on an NFL team's season schedule (see
// GetTeamSchedule) — both past results and the remaining, not-yet-played
// games.
type Game struct {
	ID                   string
	Date                 string
	Week                 int
	SeasonType           string // "Preseason", "Regular Season", "Postseason"
	OpponentAbbreviation string
	OpponentName         string
	HomeAway             string // "home" or "away"
	// Played is false for a game that hasn't happened yet — Won/TeamScore/
	// OpponentScore are all zero-valued in that case.
	Played        bool
	Won           bool
	TeamScore     string
	OpponentScore string
}

type scheduleResponse struct {
	Events []struct {
		ID         string `json:"id"`
		Date       string `json:"date"`
		SeasonType struct {
			Name string `json:"name"`
		} `json:"seasonType"`
		Week struct {
			Number int `json:"number"`
		} `json:"week"`
		Competitions []struct {
			Competitors []struct {
				HomeAway string `json:"homeAway"`
				Winner   bool   `json:"winner"`
				Team     struct {
					ID           string `json:"id"`
					Abbreviation string `json:"abbreviation"`
					DisplayName  string `json:"displayName"`
				} `json:"team"`
				Score struct {
					DisplayValue string `json:"displayValue"`
				} `json:"score"`
			} `json:"competitors"`
		} `json:"competitions"`
	} `json:"events"`
}

// GetTeamSchedule fetches an NFL team's full current-season schedule —
// past results and the remaining games — ordered as ESPN returns it
// (chronological).
func GetTeamSchedule(client *http.Client, sport Sport, league League, teamID string) ([]Game, error) {
	return fetchTeamSchedule(client, sport, league, teamID, "")
}

// GetTeamScheduleForSeason is GetTeamSchedule pinned to a specific season —
// GetTeamSchedule itself has no season parameter and always follows
// whatever ESPN currently considers "the" schedule for that team (which,
// for much of the year, is a future/in-progress season, not the season
// callers usually mean). See GetTeamDefenseGameLog, which needs a specific
// past season's results.
func GetTeamScheduleForSeason(client *http.Client, sport Sport, league League, teamID string, season int) ([]Game, error) {
	return fetchTeamSchedule(client, sport, league, teamID, "?season="+strconv.Itoa(season))
}

func fetchTeamSchedule(client *http.Client, sport Sport, league League, teamID, querySuffix string) ([]Game, error) {
	if client == nil {
		client = http.DefaultClient
	}

	url := strings.NewReplacer(
		"{sport}", string(sport),
		"{league}", string(league),
		"{teamId}", teamID,
	).Replace(teamScheduleURL) + querySuffix

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching schedule for team %s: %w", teamID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d fetching schedule for team %s", resp.StatusCode, teamID)
	}

	var raw scheduleResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding schedule for team %s: %w", teamID, err)
	}

	games := make([]Game, 0, len(raw.Events))
	for _, e := range raw.Events {
		if len(e.Competitions) == 0 || len(e.Competitions[0].Competitors) != 2 {
			continue
		}

		var mine, opp struct {
			homeAway string
			winner   bool
			name     string
			abbr     string
			score    string
		}
		for _, c := range e.Competitions[0].Competitors {
			if c.Team.ID == teamID {
				mine.homeAway, mine.winner, mine.score = c.HomeAway, c.Winner, c.Score.DisplayValue
			} else {
				opp.name, opp.abbr, opp.score = c.Team.DisplayName, c.Team.Abbreviation, c.Score.DisplayValue
			}
		}

		games = append(games, Game{
			ID:                   e.ID,
			Date:                 e.Date,
			Week:                 e.Week.Number,
			SeasonType:           e.SeasonType.Name,
			OpponentAbbreviation: opp.abbr,
			OpponentName:         opp.name,
			HomeAway:             mine.homeAway,
			Played:               mine.score != "",
			Won:                  mine.winner,
			TeamScore:            mine.score,
			OpponentScore:        opp.score,
		})
	}

	return games, nil
}
