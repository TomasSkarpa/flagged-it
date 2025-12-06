package data

import (
	"embed"
	"encoding/json"
	"flagged-it/internal/data/models"
	"sync"
)

//go:embed sources/countries_main.json
var countriesData []byte

//go:embed sources/countries_facts.json
var factsData []byte

//go:embed sources/geo/*.json
var geoFS embed.FS

var (
	cachedCountries    []models.Country
	cachedCountryFacts map[string]models.CountryFacts
	countriesOnce      sync.Once
	factsOnce          sync.Once
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

func LoadGeoData(cca3 string) (models.GeoJSON, error) {
	data, err := geoFS.ReadFile("sources/geo/" + cca3 + ".json")
	if err != nil {
		return models.GeoJSON{}, err
	}
	var geoData models.GeoJSON
	err = json.Unmarshal(data, &geoData)
	return geoData, err
}
