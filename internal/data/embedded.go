package data

import (
	"embed"
	"encoding/json"
	"fmt"
	"flagged-it/internal/data/models"
	"sync"
)

//go:embed sources/countries_main.json
var countriesData []byte

//go:embed sources/countries_facts.json
var factsData []byte

//go:embed sources/geo/*.json
var geoFS embed.FS

//go:embed sources/geo.json
var worldGeoData []byte

var (
	cachedCountries    []models.Country
	cachedCountryFacts map[string]models.CountryFacts
	cachedGeoData      map[string]models.GeoJSON
	cachedWorldGeo     *models.GeoJSON
	countriesOnce      sync.Once
	factsOnce          sync.Once
	geoOnce            sync.Once
	worldGeoOnce       sync.Once
	geoMutex           sync.RWMutex
	worldGeoMutex      sync.RWMutex
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

// LoadWorldGeoData loads and caches the world GeoJSON data
func LoadWorldGeoData() (*models.GeoJSON, error) {
	var loadErr error
	worldGeoOnce.Do(func() {
		var worldGeo models.GeoJSON
		if err := json.Unmarshal(worldGeoData, &worldGeo); err != nil {
			loadErr = err
			return
		}
		cachedWorldGeo = &worldGeo
	})

	if loadErr != nil {
		return nil, fmt.Errorf("failed to load world GeoJSON: %w", loadErr)
	}

	if cachedWorldGeo == nil {
		return nil, fmt.Errorf("failed to load world GeoJSON")
	}

	return cachedWorldGeo, nil
}
