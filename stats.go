package espn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// Stat is a single named statistic within a StatCategory — e.g.
// "passingYards" for a QB. DisplayValue is ESPN's own formatted string
// (e.g. "3,587"), not the raw number — some stats (completion percentage,
// QBR) are only meaningful as ESPN already formats them. DisplayName/
// ShortDisplayName are ESPN's own human-readable labels ("Passing Yards"/
// "YDS") — Name is the raw camelCase API field ("passingYards"), not a
// good UI label on its own.
type Stat struct {
	Name             string
	DisplayName      string
	ShortDisplayName string
	DisplayValue     string
}

// StatCategory is one category's worth of stats for a single season (see
// SeasonStats) — e.g. "passing" for a QB, "receiving" for a WR. Only
// stats with a non-zero value are included — ESPN attaches every stat in
// every category to every player regardless of position or relevance, so
// e.g. a quarterback's "general" category would otherwise carry a long
// tail of always-zero defensive stats (forced fumbles, tackles, ...)
// alongside the handful that actually apply.
type StatCategory struct {
	Name        string
	DisplayName string
	Stats       []Stat
}

// SeasonStats is a player's full statistical season, from ESPN's
// season-scoped athlete statistics endpoint (regular season only — see
// GetPlayerSeasonStats). Every category is present, but ESPN attaches all
// of them (passing, rushing, receiving, defensive, ...) to every player
// regardless of position — Categories here only includes ones with at
// least one non-zero stat, so a receiver's response won't carry an empty
// "passing" category.
type SeasonStats struct {
	Season     int
	Categories []StatCategory
}

type statsResponse struct {
	Splits struct {
		Categories []struct {
			Name        string `json:"name"`
			DisplayName string `json:"displayName"`
			Stats       []struct {
				Name             string  `json:"name"`
				DisplayName      string  `json:"displayName"`
				ShortDisplayName string  `json:"shortDisplayName"`
				DisplayValue     string  `json:"displayValue"`
				Value            float64 `json:"value"`
			} `json:"stats"`
		} `json:"categories"`
	} `json:"splits"`
}

// GetPlayerSeasonStats fetches an athlete's regular-season statistics for
// a single season. Returns a zero SeasonStats (no error) if ESPN has no
// stats for that player in that season — e.g. a year before they were in
// the league — since that's an expected, common case for callers iterating
// several seasons, not a real failure.
func GetPlayerSeasonStats(client *http.Client, sport Sport, league League, athleteID string, season int) (SeasonStats, error) {
	if client == nil {
		client = http.DefaultClient
	}

	url := strings.NewReplacer(
		"{sport}", string(sport),
		"{league}", string(league),
		"{athleteId}", athleteID,
		"{season}", strconv.Itoa(season),
	).Replace(athleteSeasonStatsURL)

	resp, err := client.Get(url)
	if err != nil {
		return SeasonStats{}, fmt.Errorf("fetching season %d stats for athlete %s: %w", season, athleteID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return SeasonStats{Season: season}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return SeasonStats{}, fmt.Errorf("unexpected status %d fetching season %d stats for athlete %s", resp.StatusCode, season, athleteID)
	}

	var raw statsResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return SeasonStats{}, fmt.Errorf("decoding season %d stats for athlete %s: %w", season, athleteID, err)
	}

	out := SeasonStats{Season: season}
	for _, c := range raw.Splits.Categories {
		var stats []Stat
		for _, s := range c.Stats {
			if s.Value == 0 {
				continue
			}
			stats = append(stats, Stat{
				Name:             s.Name,
				DisplayName:      s.DisplayName,
				ShortDisplayName: s.ShortDisplayName,
				DisplayValue:     s.DisplayValue,
			})
		}
		if len(stats) == 0 {
			continue
		}
		out.Categories = append(out.Categories, StatCategory{
			Name:        c.Name,
			DisplayName: c.DisplayName,
			Stats:       stats,
		})
	}

	return out, nil
}
