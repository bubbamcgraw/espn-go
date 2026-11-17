package espn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// proTeamAbbreviations maps ESPN's internal numeric NFL team ID
// (proTeamId in the fantasy players API) to the standard abbreviation used
// everywhere else in this codebase (sleeper-go's Player.Team, this
// package's own Team.Abbreviation) — notably "WAS" for Washington, not
// ESPN's own "WSH", so callers can match directly against sleeper-go data
// without a second translation. ESPN doesn't document these IDs; this is
// the same fixed mapping used across the open-source ESPN fantasy API
// ecosystem. 0 (and any ID not in this map) means a free agent/no team.
var proTeamAbbreviations = map[int]string{
	1: "ATL", 2: "BUF", 3: "CHI", 4: "CIN", 5: "CLE", 6: "DAL", 7: "DEN",
	8: "DET", 9: "GB", 10: "TEN", 11: "IND", 12: "KC", 13: "LV", 14: "LAR",
	15: "MIA", 16: "MIN", 17: "NE", 18: "NO", 19: "NYG", 20: "NYJ",
	21: "PHI", 22: "ARI", 23: "PIT", 24: "LAC", 25: "SF", 26: "SEA",
	27: "TB", 28: "WAS", 29: "CAR", 30: "JAX", 33: "BAL", 34: "HOU",
}

// teamIDByAbbreviation is the inverse of proTeamAbbreviations, built once.
// Empirically, ESPN reuses the same internal numeric team ID across its
// fantasy API (proTeamId, above) and its separate site API that GetTeam/
// GetTeamSchedule call (teams/{id}) — e.g. Detroit is 8 and Green Bay is 9
// in both — so this one table serves both.
var teamIDByAbbreviation = func() map[string]int {
	m := make(map[string]int, len(proTeamAbbreviations))
	for id, abbr := range proTeamAbbreviations {
		m[abbr] = id
	}
	return m
}()

// TeamAbbreviations returns every NFL team abbreviation TeamID knows.
func TeamAbbreviations() []string {
	out := make([]string, 0, len(proTeamAbbreviations))
	for _, abbr := range proTeamAbbreviations {
		out = append(out, abbr)
	}
	return out
}

// TeamID looks up ESPN's internal numeric team ID for a standard NFL team
// abbreviation (e.g. "SF") — for GetTeam/GetTeamSchedule, which key off
// that numeric ID rather than the abbreviation everything else in this
// codebase uses.
func TeamID(abbreviation string) (int, bool) {
	id, ok := teamIDByAbbreviation[strings.ToUpper(strings.TrimSpace(abbreviation))]
	return id, ok
}

// positionAbbreviations maps ESPN's internal numeric position ID
// (defaultPositionId) to the abbreviation sleeper-go's Player.Position
// uses.
var positionAbbreviations = map[int]string{
	1: "QB", 2: "RB", 3: "WR", 4: "TE", 5: "K", 16: "DEF",
}

// DraftPoolPlayer is one player from ESPN's season-wide draft-prep pool
// (see GetDraftPool) — enough to rank a player, and, when a Sleeper player
// has no espn_id of their own, to match them to this ESPN record by name,
// team, and position instead.
type DraftPoolPlayer struct {
	ID       int
	FullName string
	Team     string // NFL team abbreviation ("" for a free agent), e.g. "SF"
	Position string // e.g. "RB"; "" if ESPN's position ID isn't one we map
	Rank     int    // ESPN's consensus draft rank; 0 if it has none
}

type draftPoolPlayerResponse struct {
	ID                   int    `json:"id"`
	FullName             string `json:"fullName"`
	ProTeamID            int    `json:"proTeamId"`
	DefaultPositionID    int    `json:"defaultPositionId"`
	DraftRanksByRankType struct {
		Standard struct {
			Rank int `json:"rank"`
		} `json:"STANDARD"`
	} `json:"draftRanksByRankType"`
}

// GetDraftPool fetches ESPN's whole draft-prep player pool for a season,
// keyed by ESPN athlete ID (the same ID GetPlayer takes, and the id
// sleeper-go's Player.EspnID bridges to). Meaningful only for the season
// currently being drafted — ESPN's rankings reset once a season starts, so
// a past or in-progress season's data isn't reliable draft-prep data.
//
// Rank (DraftPoolPlayer.Rank) is deliberately sourced from
// draftRanksByRankType.STANDARD.rank rather than the response's own
// ownership.averageDraftPosition field. The latter looks like real ADP but
// is only meaningful for players with heavy real draft volume behind them
// (observed: needs roughly 50%+ ESPN ownership before it stops being a
// noisy, near-constant ~170 shared by thousands of unrelated bench/waiver
// players) — this early in a preseason, that's a few hundred players at
// most. rank is a full, non-degenerate ordinal across the entire draftable
// pool and tracks averageDraftPosition closely wherever the latter is
// trustworthy (spot-checked: top ~15 by real ADP have ranks 1-20, in almost
// the same order), so it's a much more complete proxy without the
// sparse-data artifact.
//
// This is the same undocumented endpoint ESPN's own draft-prep UI calls; it
// takes no leagueId; the x-fantasy-filter header (rather than a query
// param) is what selects/sorts the player set.
func GetDraftPool(client *http.Client, season string) ([]DraftPoolPlayer, error) {
	if client == nil {
		client = http.DefaultClient
	}

	reqURL := strings.NewReplacer("{season}", season).Replace(fantasyPlayersURL)

	q := url.Values{}
	q.Add("scoringPeriodId", "0")
	q.Add("view", "kona_player_info")

	req, err := http.NewRequest(http.MethodGet, reqURL+"?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("building draft pool request: %w", err)
	}
	// sortPriority/sortAsc pick STANDARD draft rank as the sort; the limit
	// is generous rather than exact — ESPN returns its whole draftable pool
	// (several thousand players) regardless of the requested limit.
	req.Header.Set("x-fantasy-filter", `{"players":{"limit":3000,"sortDraftRanks":{"sortPriority":100,"sortAsc":true,"value":"STANDARD"}}}`)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching draft pool for season %s: %w", season, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d fetching draft pool for season %s", resp.StatusCode, season)
	}

	var raw []draftPoolPlayerResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding draft pool for season %s: %w", season, err)
	}

	pool := make([]DraftPoolPlayer, len(raw))
	for i, p := range raw {
		pool[i] = DraftPoolPlayer{
			ID:       p.ID,
			FullName: p.FullName,
			Team:     proTeamAbbreviations[p.ProTeamID],
			Position: positionAbbreviations[p.DefaultPositionID],
			Rank:     p.DraftRanksByRankType.Standard.Rank,
		}
	}

	return pool, nil
}
