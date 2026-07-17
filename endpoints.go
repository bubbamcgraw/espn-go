package espn

const (
	baseTeamsURL    = "https://site.api.espn.com/apis/site/v2/sports/{sport}/{league}/teams"
	teamURL         = baseTeamsURL + "/{teamId}"
	teamRosterURL   = teamURL + "/roster"
	teamScheduleURL = teamURL + "/schedule"
)

const (
	athleteStatsURL = "https://sports.core.api.espn.com/v2/sports/{sport}/leagues/{league}/athletes/{athleteId}/statistics"
	athleteEventURL = "https://sports.core.api.espn.com/v2/sports/{sport}/leagues/{league}/athletes/{athleteId}/eventlog"
	athleteURL      = "https://site.web.api.espn.com/apis/common/v3/sports/{sport}/{league}/athletes/{athleteId}"
)

type Sport string

const (
	Football Sport = "football"
)

type League string

const (
	NFL League = "nfl"
)
