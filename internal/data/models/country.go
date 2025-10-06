package models

type GeoPosition struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Country struct {
	CountryName       string      `json:"country_name"`
	CountryCode       string      `json:"country_code"`
	Continent         string      `json:"continent"`
	Region            string      `json:"region"`
	Population        int         `json:"population"`
	Capital           string      `json:"capital"`
	GeoPosition       GeoPosition `json:"geo_position"`
	OfficialLanguages []string    `json:"official_languages"`
	Facts             []string    `json:"facts"`
}
