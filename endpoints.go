package espn

const (
	baseTeamsURL    = "https://site.api.espn.com/apis/site/v2/sports/{sport}/{league}/teams"
	teamURL         = baseTeamsURL + "/{teamId}"
	teamRosterURL   = teamURL + "/roster"
	teamScheduleURL = teamURL + "/schedule"
	// gameSummaryURL is the same host/API family as the team endpoints
	// above (not the athlete gamelog's site.web.api host) — the full game
	// recap payload, including both teams' boxscore stats. See
	// GetTeamDefenseGameLog, which derives D/ST fantasy stats from it:
	// ESPN has no dedicated team-defense stats endpoint.
	gameSummaryURL = "https://site.api.espn.com/apis/site/v2/sports/{sport}/{league}/summary?event={eventId}"
)

const (
	// athleteSeasonStatsURL is season-scoped (regular season, "types/2") —
	// see GetPlayerSeasonStats. The unscoped .../athletes/{id}/statistics
	// only gives ESPN's own idea of "current" stats, not a specific season.
	athleteSeasonStatsURL = "https://sports.core.api.espn.com/v2/sports/{sport}/leagues/{league}/seasons/{season}/types/2/athletes/{athleteId}/statistics"
	athleteEventURL       = "https://sports.core.api.espn.com/v2/sports/{sport}/leagues/{league}/athletes/{athleteId}/eventlog"
	athleteURL            = "https://site.web.api.espn.com/apis/common/v3/sports/{sport}/{league}/athletes/{athleteId}"
	// athleteGameLogURL is the same host as athleteURL (the "friendly" site
	// API, not the raw sports.core.api.espn.com one) — see
	// GetPlayerGameLog. {season} is appended as a query param, not a path
	// segment.
	athleteGameLogURL = "https://site.web.api.espn.com/apis/common/v3/sports/{sport}/{league}/athletes/{athleteId}/gamelog"
)

// fantasyLeagueURL is ESPN's fantasy football API, a separate host and
// undocumented v3 API distinct from the real-world sports data above.
// Public leagues only require no auth here; the old fantasy.espn.com host
// 302-redirects unauthenticated requests, so this must stay on
// lm-api-reads.
const fantasyLeagueURL = "https://lm-api-reads.fantasy.espn.com/apis/v3/games/ffl/seasons/{season}/segments/0/leagues/{leagueId}"

// fantasyPlayersURL is the same undocumented fantasy API's global player
// pool for a season (not scoped to a league) — used for draft prep data
// like average draft position, via the x-fantasy-filter header (see
// GetADP).
const fantasyPlayersURL = "https://lm-api-reads.fantasy.espn.com/apis/v3/games/ffl/seasons/{season}/players"

type Sport string

const (
	Football Sport = "football"
)

type League string

const (
	NFL League = "nfl"
)
