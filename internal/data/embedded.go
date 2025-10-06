package data

import (
	_ "embed"
	"encoding/json"
	"flagged-it/internal/data/models"
)

//go:embed sources/countries.json
var countriesData []byte

func LoadCountries() []models.Country {
	var countries []models.Country
	json.Unmarshal(countriesData, &countries)
	return countries
}
