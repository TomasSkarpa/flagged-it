package data

import (
	_ "embed"
	"encoding/json"
	"flagged-it/internal/data/models"
	"sync"
)

//go:embed sources/countries_main.json
var countriesData []byte

//go:embed sources/countries_facts.json
var factsData []byte

//go:embed sources/geo.json
var geoData []byte

var (
	cachedCountries    []models.Country
	cachedCountryFacts map[string]models.CountryFacts
	cachedGeoData      models.GeoJSON
	countriesOnce      sync.Once
	factsOnce          sync.Once
	geoDataOnce        sync.Once
)

func LoadCountries() []models.Country {
	countriesOnce.Do(func() {
		json.Unmarshal(countriesData, &cachedCountries)
	})
	return cachedCountries
}

func LoadCountryFacts() map[string]models.CountryFacts {
	factsOnce.Do(func() {
		json.Unmarshal(factsData, &cachedCountryFacts)
	})
	return cachedCountryFacts
}

func LoadGeoData() models.GeoJSON {
	geoDataOnce.Do(func() {
		json.Unmarshal(geoData, &cachedGeoData)
	})
	return cachedGeoData
}
