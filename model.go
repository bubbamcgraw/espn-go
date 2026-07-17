package espn

type AthleteResponse struct {
	Athlete Athlete `json:"athlete"`
}

type Athlete struct {
	ID            string     `json:"id"`
	UID           string     `json:"uid"`
	GUID          string     `json:"guid"`
	FirstName     string     `json:"firstName"`
	LastName      string     `json:"lastName"`
	FullName      string     `json:"fullName"`
	DisplayName   string     `json:"displayName"`
	ShortName     string     `json:"shortName"`
	Weight        float64    `json:"weight"`
	DisplayWeight string     `json:"displayWeight"`
	Height        float64    `json:"height"`
	DisplayHeight string     `json:"displayHeight"`
	Age           int        `json:"age"`
	DateOfBirth   string     `json:"dateOfBirth"`
	BirthPlace    BirthPlace `json:"birthPlace"`
	College       College    `json:"college"`
	Slug          string     `json:"slug"`
	Headshot      Headshot   `json:"headshot"`
	Jersey        string     `json:"jersey"`
	Position      Position   `json:"position"`
	Team          Team       `json:"team"`
	Status        Status     `json:"status"`
	Active        bool       `json:"active"`
}

type BirthPlace struct {
	City    string `json:"city"`
	State   string `json:"state"`
	Country string `json:"country"`
}

type College struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	Mascot    string `json:"mascot"`
}

type Headshot struct {
	Href string `json:"href"`
	Alt  string `json:"alt"`
}

type Position struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	DisplayName  string `json:"displayName"`
	Abbreviation string `json:"abbreviation"`
}

type Team struct {
	ID           string `json:"id"`
	Location     string `json:"location"`
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
	DisplayName  string `json:"displayName"`
}

type Status struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Abbreviation string `json:"abbreviation"`
}
