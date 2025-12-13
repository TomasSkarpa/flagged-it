package models

type CountryName struct {
	Common   string `json:"common"`
	Official string `json:"official"`
}

type Country struct {
	Name       CountryName       `json:"name"`
	CCA2       string            `json:"cca2"`
	CCA3       string            `json:"cca3"`
	Capital    []string          `json:"capital"`
	Region     string            `json:"region"`
	Subregion  string            `json:"subregion"`
	Languages  map[string]string `json:"languages"`
	Latlng     []float64         `json:"latlng"`
	Population int               `json:"population"`
	Area       float64           `json:"area"`
}

type CountryFacts struct {
	Name  string   `json:"name"`
	Facts []string `json:"facts"`
}
