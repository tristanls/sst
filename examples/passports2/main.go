package main

import (
	"fmt"

	"github.com/tristanls/sst"
)

type NodeType string

var (
	Location NodeType = "Location"
	Country  NodeType = "Country"
	Person   NodeType = "Person"
	Event    NodeType = "Event"
)

func main() {
	sst.MustCreateAssociation(&sst.Association{
		Key:          "happened_in",
		SemanticType: sst.Follows,
		Fwd:          "happened in",
		Bwd:          "was the location of",
		Nfwd:         "did not happen in",
		Nbwd:         "was not the location of",
	})
	sst.MustCreateAssociation(&sst.Association{
		Key:          "involved",
		SemanticType: sst.Follows,
		Fwd:          "involved",
		Bwd:          "was involved in",
		Nfwd:         "did not involve",
		Nbwd:         "was not involved in",
	})
	config := &sst.Config{
		Name:            "nation_spacetime_2",
		NodeCollections: []string{string(Location), string(Country), string(Person), string(Event)},
		Password:        "",
		URL:             "http://localhost:8529",
		Username:        "root",
	}
	s, err := sst.NewSST(config)
	if err != nil {
		panic(err)
	}

	mb1 := CreatePerson(s, "markburgess_osl", "Professor Mark Burgess", 123456, 0)
	mb2 := CreatePerson(s, "Professor Burgess", "Professor Mark Burgess", 123456, 0)

	s.MustCreateLink(mb1, "alias", mb2, 0)

	CreateCountry(s, "USA", "United States of America")
	CreateCountry(s, "UK", "United Kingdom")

	CreateLocation(s, "London", "London, capital city in England")
	CreateLocation(s, "Washington DC", "Washington, capital city in USA")
	CreateLocation(s, "New York", "Capital of the World")

	LocationCountry(s, "Washington DC", "USA")
	LocationCountry(s, "New York", "USA")
	LocationCountry(s, "London", "UK")

	france := CreateCountry(s, "France", "France, country in Europe")
	paris := CreateLocation(s, "Paris", "Paris, capital city in France")

	s.MustCreateLink(paris, "part_of", france, 0)

	// Mark's journey as a sequential process

	CountryIssuedPassport(s, "markburgess_osl", "UK", "Number 12345")
	CountryIssuedVisa(s, "markburgess_osl", "USA", "Visa Waiver")

	PersonLocation(s, "markburgess_osl", "New York")
	PersonLocation(s, "markburgess_osl", "London")

	CountryIssuedVisa(s, "Emily", "France", "Schengen work visa")
	PersonLocation(s, "Emily", "Paris")

	CountryIssuedVisa(s, "Captain Evil", "USA", "Work Visa")

	PersonLocation(s, "Captain Evil", "London")
	PersonLocation(s, "Captain Evil", "Washington DC")
}

func CreatePerson(s *sst.SST, short, description string, number int, weight float64) *sst.Node {
	return s.MustCreateNode(string(Person), short, map[string]interface{}{"description": description, "number": number}, weight)
}

func CreateCountry(s *sst.SST, short, description string) *sst.Node {
	var data map[string]interface{}
	if description != "" {
		data = map[string]interface{}{"description": description}
	}
	return s.MustCreateNode(string(Country), short, data, 0)
}

func CreateLocation(s *sst.SST, short, description string) *sst.Node {
	var data map[string]interface{}
	if description != "" {
		data = map[string]interface{}{"description": description}
	}
	return s.MustCreateNode(string(Location), short, data, 0)
}

func CreateEvent(s *sst.SST, short, description string) *sst.Node {
	var data map[string]interface{}
	if description != "" {
		data = map[string]interface{}{"description": description}
	}
	return s.MustCreateNode(string(Event), short, data, 0)
}

func LocationCountry(s *sst.SST, location, country string) {
	loc := CreateLocation(s, location, "")
	c := CreateCountry(s, country, "")

	s.CreateLink(c, "contains", loc, 1)

	fmt.Println("Location: ", loc.Key, "is in", country)
}

func PersonLocation(s *sst.SST, person, location string) {
	prsn := CreatePerson(s, person, "", 0, 0)
	loc := CreateLocation(s, location, "")

	short := person + " in " + location
	long := person + " observed in " + location

	minihub := CreateEvent(s, short, long)

	s.CreateLink(minihub, "happened_in", loc, 1)
	s.CreateLink(minihub, "involved", prsn, 1)

	s.MustNextEvent(string(Event), short, map[string]interface{}{"description": long})
	fmt.Println("Timeline: " + short)
}

func CountryIssuedPassport(s *sst.SST, person, country, passport string) {
	ctry := CreateCountry(s, country, "")
	prsn := CreatePerson(s, person, "", 0, 0)
	timeLimit := 1.0
	sst.MustCreateAssociation(&sst.Association{
		Key:          passport,
		SemanticType: sst.Expresses,
		Fwd:          "grants passport to",
		Bwd:          "holds passport from",
		Nfwd:         "did not grant passport to",
		Nbwd:         "does not hold passport from",
	})
	s.MustCreateLink(ctry, passport, prsn, timeLimit)
	s.MustNextEvent(
		string(Event),
		country+" grants "+passport+" to "+person,
		map[string]interface{}{"description": country + " granted passport " + passport + " to " + person},
	)
	fmt.Println("Timeline: " + country + " granted passport " + passport + " to " + person)
}

func CountryIssuedVisa(s *sst.SST, person, country, visa string) {
	ctry := CreateCountry(s, country, "")
	prsn := CreatePerson(s, person, "", 0, 0)
	timeLimit := 1.0
	sst.MustCreateAssociation(&sst.Association{
		Key:          visa,
		SemanticType: sst.Expresses,
		Fwd:          "grants visa to",
		Bwd:          "holds visa from",
		Nfwd:         "did not grant visa to",
		Nbwd:         "does not hold visa from",
	})
	s.MustCreateLink(ctry, visa, prsn, timeLimit)
	s.MustNextEvent(
		string(Event),
		country+" grants "+visa+" to "+person,
		map[string]interface{}{"description": country + " granted visa " + visa + " to " + person},
	)
	fmt.Println("Timeline: " + country + " granted visa " + visa + " to " + person)
}
