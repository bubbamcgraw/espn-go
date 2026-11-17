package espn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// GameLogWeek is one game's stat line from a player's game log (see
// GetPlayerGameLog). Embeds Game for the same per-game context
// GetTeamSchedule already exposes (opponent, date, W/L, score) — Played is
// always true here, since an unplayed game (bye week, future game) simply
// isn't present in the underlying feed at all, unlike GetTeamSchedule's
// full season list which includes games yet to happen.
type GameLogWeek struct {
	Game

	// Stats is this player's full stat line for the game — every category
	// ESPN tracks for their position that game (e.g. passing and rushing
	// together for a QB) unified into one flat list, positionally the same
	// Stat shape SeasonStats/StatCategory use. Unlike season stats, there's
	// no sub-category split and no filtering of zero-value stats — ESPN's
	// game log already returns just the columns relevant to this player.
	Stats []Stat
}

type gameLogResponse struct {
	Names        []string                 `json:"names"`
	Labels       []string                 `json:"labels"`
	DisplayNames []string                 `json:"displayNames"`
	Events       map[string]gameLogEvent  `json:"events"`
	SeasonTypes  []gameLogSeasonTypeEntry `json:"seasonTypes"`
}

type gameLogEvent struct {
	ID            string `json:"id"`
	Week          int    `json:"week"`
	GameDate      string `json:"gameDate"`
	HomeTeamID    string `json:"homeTeamId"`
	AwayTeamID    string `json:"awayTeamId"`
	HomeTeamScore string `json:"homeTeamScore"`
	AwayTeamScore string `json:"awayTeamScore"`
	GameResult    string `json:"gameResult"`
	Team          struct {
		ID string `json:"id"`
	} `json:"team"`
	Opponent struct {
		Abbreviation string `json:"abbreviation"`
		DisplayName  string `json:"displayName"`
	} `json:"opponent"`
}

type gameLogSeasonTypeEntry struct {
	DisplayName string `json:"displayName"`
	Categories  []struct {
		Events []struct {
			EventID string   `json:"eventId"`
			Stats   []string `json:"stats"`
		} `json:"events"`
	} `json:"categories"`
}

// GetPlayerGameLog fetches a player's game-by-game stat log for a season —
// every season type ESPN has data for that year (regular season, and
// postseason once/if their team made the playoffs) is returned together.
// Returns (nil, nil), not an error, if ESPN has no game log for this
// athlete in that season — e.g. a year before they were in the league —
// since that's an expected, common case for callers iterating several
// seasons, not a real failure (same convention as GetPlayerSeasonStats).
func GetPlayerGameLog(client *http.Client, sport Sport, league League, athleteID string, season int) ([]GameLogWeek, error) {
	if client == nil {
		client = http.DefaultClient
	}

	url := strings.NewReplacer(
		"{sport}", string(sport),
		"{league}", string(league),
		"{athleteId}", athleteID,
	).Replace(athleteGameLogURL) + "?season=" + strconv.Itoa(season)

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching game log for athlete %s season %d: %w", athleteID, season, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d fetching game log for athlete %s season %d", resp.StatusCode, athleteID, season)
	}

	var raw gameLogResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding game log for athlete %s season %d: %w", athleteID, season, err)
	}

	var weeks []GameLogWeek
	for _, st := range raw.SeasonTypes {
		seasonType := seasonTypeFromGameLogDisplayName(st.DisplayName)
		for _, cat := range st.Categories {
			for _, ev := range cat.Events {
				meta, ok := raw.Events[ev.EventID]
				if !ok {
					continue
				}
				weeks = append(weeks, buildGameLogWeek(raw.Names, raw.Labels, raw.DisplayNames, seasonType, meta, ev.Stats))
			}
		}
	}

	return weeks, nil
}

func buildGameLogWeek(names, labels, displayNames []string, seasonType string, meta gameLogEvent, values []string) GameLogWeek {
	stats := make([]Stat, 0, len(names))
	for i, name := range names {
		if i >= len(values) {
			break
		}
		stats = append(stats, Stat{
			Name:             name,
			DisplayName:      stringAt(displayNames, i),
			ShortDisplayName: stringAt(labels, i),
			DisplayValue:     values[i],
		})
	}

	homeAway := "away"
	teamScore, opponentScore := meta.AwayTeamScore, meta.HomeTeamScore
	if meta.Team.ID == meta.HomeTeamID {
		homeAway = "home"
		teamScore, opponentScore = meta.HomeTeamScore, meta.AwayTeamScore
	}

	return GameLogWeek{
		Game: Game{
			ID:                   meta.ID,
			Date:                 meta.GameDate,
			Week:                 meta.Week,
			SeasonType:           seasonType,
			OpponentAbbreviation: meta.Opponent.Abbreviation,
			OpponentName:         meta.Opponent.DisplayName,
			HomeAway:             homeAway,
			Played:               true,
			Won:                  meta.GameResult == "W",
			TeamScore:            teamScore,
			OpponentScore:        opponentScore,
		},
		Stats: stats,
	}
}

func stringAt(s []string, i int) string {
	if i < len(s) {
		return s[i]
	}
	return ""
}

// seasonTypeFromGameLogDisplayName maps a game log seasonType's own
// displayName (e.g. "2025 Regular Season", "2024 Postseason") to the same
// "Preseason"/"Regular Season"/"Postseason" strings GetTeamSchedule uses,
// so every Game value in this package agrees on that vocabulary regardless
// of which endpoint produced it.
func seasonTypeFromGameLogDisplayName(displayName string) string {
	switch {
	case strings.Contains(displayName, "Preseason"):
		return "Preseason"
	case strings.Contains(displayName, "Postseason"):
		return "Postseason"
	default:
		return "Regular Season"
	}
}
