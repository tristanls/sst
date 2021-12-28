package main

import (
	"fmt"

	"github.com/tristanls/sst"
)

func main() {
	config := &sst.Config{
		Name:     "nation_spacetime",
		Password: "",
		URL:      "http://localhost:8529",
		Username: "root",
	}
	s, err := sst.NewSST(config)
	if err != nil {
		panic(err)
	}
	// Modeling choices:
	// - Nodes are Events
	// - Frags are Persons
	// - Hubs are Locations

	// Mark's journey as a sequential process
	CountryIssuedPassport(s, "Professor Burgess", "UK", "Number 12345")
	CountryIssuedVisa(s, "Professor Burgess", "USA", "Visa Waiver")
	PersonLocation(s, "Professor Burgess", "USA")
	PersonLocation(s, "Professor Burgess", "UK")

	paris := s.MustCreateHub("Paris", "Paris, capital city of France", 1)
	france := s.MustCreateHub("France", "France, country in Europe", 100)

	CountryIssuedVisa(s, "Emily", "France", "Schengen work visa")
	PersonLocation(s, "Emily", "Paris")

	s.MustCreateLink(paris, "part_of", france, 100)

	CountryIssuedVisa(s, "Captain Evil", "USA", "Work Visa")
	PersonLocation(s, "Captain Evil", "UK")
	PersonLocation(s, "Captain Evil", "USA")
}

func PersonLocation(s *sst.SST, person, location string) {
	s.MustCreateFragment(person, person, 0)
	s.MustCreateHub(location, "", 0)
	s.MustNextEvent(person+" in "+location, person+" observed in "+location)
	fmt.Println("Timeline: " + person + " in " + location)
}

func CountryIssuedPassport(s *sst.SST, person, location, passport string) {
	countryHub := s.MustCreateHub(location, "", 0)
	personFrag := s.MustCreateFragment(person, "", 0)
	timeLimit := 1.0
	sst.MustCreateAssociation(&sst.Association{
		Key:          passport,
		SemanticType: sst.Expresses,
		Fwd:          "grants passport to",
		Bwd:          "holds passport from",
		Nfwd:         "did not grant passport to",
		Nbwd:         "does not hold passport from",
	})
	s.MustCreateLink(countryHub, passport, personFrag, timeLimit)
	s.MustNextEvent(location+" grants "+passport+" to "+person, location+" granted passport "+passport+" to "+person)
	fmt.Println("Timeline: " + location + " granted passport " + passport + " to " + person)
}

func CountryIssuedVisa(s *sst.SST, person, location, visa string) {
	countryHub := s.MustCreateHub(location, "", 0)
	personFrag := s.MustCreateFragment(person, "", 0)
	timeLimit := 1.0
	sst.MustCreateAssociation(&sst.Association{
		Key:          visa,
		SemanticType: sst.Expresses,
		Fwd:          "grants visa to",
		Bwd:          "holds visa from",
		Nfwd:         "did not grant visa to",
		Nbwd:         "does not hold visa from",
	})
	s.MustCreateLink(countryHub, visa, personFrag, timeLimit)
	s.MustNextEvent(location+" grants "+visa+" to "+person, location+" granted visa "+visa+" to "+person)
	fmt.Println("Timeline: " + location + " granted visa " + visa + " to " + person)
}
