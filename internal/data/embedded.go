package data

import (
	_ "embed"
	"encoding/json"
	"flagged-it/internal/data/models"
)

//go:embed sources/countries_main.json
var countriesData []byte

//go:embed sources/countries_facts.json
var factsData []byte

//go:embed sources/geo.json
var geoData []byte

func LoadCountries() []models.Country {
	var countries []models.Country
	json.Unmarshal(countriesData, &countries)
	return countries
}

func LoadCountryFacts() map[string]models.CountryFacts {
	var facts map[string]models.CountryFacts
	json.Unmarshal(factsData, &facts)
	return facts
}

func LoadGeoData() models.GeoJSON {
	var geo models.GeoJSON
	json.Unmarshal(geoData, &geo)
	return geo
}
