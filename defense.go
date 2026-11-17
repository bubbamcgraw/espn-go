package espn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// TeamDefenseWeek is one game's derived team-defense stat line, from
// GetTeamDefenseGameLog. ESPN has no dedicated team-defense stats endpoint
// — fantasy D/ST scoring is built here from the standard game recap
// boxscore: this team's defensive production is its *opponent's* offensive
// negatives that game (their interceptions thrown and fumbles lost become
// this team's takeaways, their sacks-taken become this team's sacks),
// plus this team's own defensive/return touchdowns and the opponent's
// final score (points allowed).
type TeamDefenseWeek struct {
	Game
	Sacks               float64
	Interceptions       float64
	FumbleRecoveries    float64
	DefensiveTouchdowns float64
	// PointsAllowed is the opponent's final score — also available as
	// Game.OpponentScore (a string), duplicated here as an int since
	// points-allowed scoring is tier/threshold-based, not a per-unit rate.
	PointsAllowed int
}

type gameSummaryResponse struct {
	Boxscore struct {
		Teams []struct {
			Team struct {
				ID string `json:"id"`
			} `json:"team"`
			Statistics []struct {
				Name         string `json:"name"`
				DisplayValue string `json:"displayValue"`
			} `json:"statistics"`
		} `json:"teams"`
	} `json:"boxscore"`
}

// teamStats returns teamID's own boxscore stats, by name -> DisplayValue.
func (r gameSummaryResponse) teamStats(teamID string) map[string]string {
	for _, t := range r.Boxscore.Teams {
		if t.Team.ID == teamID {
			out := make(map[string]string, len(t.Statistics))
			for _, s := range t.Statistics {
				out[s.Name] = s.DisplayValue
			}
			return out
		}
	}
	return nil
}

// GetTeamDefenseGameLog fetches a team's game-by-game defensive stat line
// for a season, derived one game summary at a time (see TeamDefenseWeek) —
// there's no single-request team gamelog the way GetPlayerGameLog has for
// athletes, so this costs one request for the season schedule plus one per
// played game. Fetches sequentially, not concurrently, since callers
// looping over several teams' defenses are expected to bound their own
// concurrency across teams (see fantasy-api's leaguePositionPerformance).
func GetTeamDefenseGameLog(client *http.Client, sport Sport, league League, teamID string, season int) ([]TeamDefenseWeek, error) {
	if client == nil {
		client = http.DefaultClient
	}

	games, err := GetTeamScheduleForSeason(client, sport, league, teamID, season)
	if err != nil {
		return nil, fmt.Errorf("fetching schedule for team %s season %d: %w", teamID, season, err)
	}

	var weeks []TeamDefenseWeek
	for _, g := range games {
		if !g.Played {
			continue
		}

		week, err := fetchTeamDefenseWeek(client, sport, league, teamID, g)
		if err != nil {
			return nil, err
		}
		if week != nil {
			weeks = append(weeks, *week)
		}
	}

	return weeks, nil
}

func fetchTeamDefenseWeek(client *http.Client, sport Sport, league League, teamID string, g Game) (*TeamDefenseWeek, error) {
	url := strings.NewReplacer(
		"{sport}", string(sport),
		"{league}", string(league),
		"{eventId}", g.ID,
	).Replace(gameSummaryURL)

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching game summary %s: %w", g.ID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d fetching game summary %s", resp.StatusCode, g.ID)
	}

	var raw gameSummaryResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding game summary %s: %w", g.ID, err)
	}

	own := raw.teamStats(teamID)
	if own == nil {
		return nil, nil
	}

	// The opponent is whichever boxscore team block isn't ours.
	var opp map[string]string
	for _, t := range raw.Boxscore.Teams {
		if t.Team.ID != teamID {
			opp = raw.teamStats(t.Team.ID)
			break
		}
	}
	if opp == nil {
		return nil, nil
	}

	pointsAllowed, _ := strconv.Atoi(g.OpponentScore)

	return &TeamDefenseWeek{
		Game:                g,
		Sacks:               madeCount(opp["sacksYardsLost"]),
		Interceptions:       numericStat(opp["interceptions"]),
		FumbleRecoveries:    numericStat(opp["fumblesLost"]),
		DefensiveTouchdowns: numericStat(own["defensiveTouchdowns"]),
		PointsAllowed:       pointsAllowed,
	}, nil
}

// madeCount parses ESPN's "N-M" combined boxscore stats (e.g.
// "sacksYardsLost": "4-9" — 4 sacks for 9 yards) and returns the first
// number. Returns 0 for anything that doesn't parse.
func madeCount(displayValue string) float64 {
	before, _, found := strings.Cut(displayValue, "-")
	if !found {
		return numericStat(displayValue)
	}
	v, err := strconv.ParseFloat(before, 64)
	if err != nil {
		return 0
	}
	return v
}

func numericStat(displayValue string) float64 {
	v, err := strconv.ParseFloat(strings.ReplaceAll(displayValue, ",", ""), 64)
	if err != nil {
		return 0
	}
	return v
}
