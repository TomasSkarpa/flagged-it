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
	cachedGeoData      map[string]models.GeoJSON
	countriesOnce      sync.Once
	factsOnce          sync.Once
	geoOnce            sync.Once
	geoMutex           sync.RWMutex
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

// LoadGeoData loads and caches GeoJSON data
func LoadGeoData(cca3 string) (models.GeoJSON, error) {
	// Initialize cache on first call
	geoOnce.Do(func() {
		cachedGeoData = make(map[string]models.GeoJSON)
	})

	// Check cache first (read lock)
	geoMutex.RLock()
	if cached, exists := cachedGeoData[cca3]; exists {
		geoMutex.RUnlock()
		return cached, nil
	}
	geoMutex.RUnlock()

	// Load from embedded FS
	data, err := geoFS.ReadFile("sources/geo/" + cca3 + ".json")
	if err != nil {
		return models.GeoJSON{}, err
	}

	var geoData models.GeoJSON
	err = json.Unmarshal(data, &geoData)
	if err != nil {
		return models.GeoJSON{}, err
	}

	// Cache the result (write lock)
	geoMutex.Lock()
	cachedGeoData[cca3] = geoData
	geoMutex.Unlock()

	return geoData, nil
}

