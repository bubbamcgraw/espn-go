package main

import (
	"fmt"
	"log"
	"os"

	espn "github.com/bubbamcgraw/espn-go"
)

func usage() {
	log.Fatalf("usage: %s player <athleteID> | %s team <teamID>", os.Args[0], os.Args[0])
}

func main() {
	if len(os.Args) != 3 {
		usage()
	}

	switch os.Args[1] {
	case "player":
		athlete, err := espn.GetPlayer(nil, espn.Football, espn.NFL, os.Args[2])
		if err != nil {
			log.Fatalf("GetPlayer failed: %v", err)
		}
		fmt.Printf("%s (%s) — %s, %s\n", athlete.FullName, athlete.ID, athlete.Position.Abbreviation, athlete.Team.Abbreviation)

	case "team":
		team, err := espn.GetTeam(nil, espn.Football, espn.NFL, os.Args[2])
		if err != nil {
			log.Fatalf("GetTeam failed: %v", err)
		}
		fmt.Printf("%s (%s)\n", team.DisplayName, team.ID)

	default:
		usage()
	}
}
