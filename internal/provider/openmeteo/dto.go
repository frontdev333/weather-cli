package openmeteo

type ResultsGeocode struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Country   string  `json:"country"`
	Admin1    string  `json:"admin1"`
}

type GeocodeWrapper struct {
	Results []ResultsGeocode `json:"results"`
}

type OpenMeteoWrapper struct {
	Current CurrentDTO `json:"current"`
}

type ForecastResp struct {
	CurrentDTO `json:"current"`
	HourlyDTO  `json:"hourly"`
	DailyDTO   `json:"daily"`
}

type CurrentDTO struct {
	Visibility          float64 `json:"visibility"`
	WeatherCode         uint    `json:"weather_code"`
	PressureMsl         float64 `json:"pressure_msl"`
	WindSpeed10m        float64 `json:"wind_speed_10m"`
	Temperature2m       float64 `json:"temperature_2m"`
	Precipitation       float64 `json:"precipitation"`
	WindDirection10m    uint    `json:"wind_direction_10m"`
	RelativeHumidity2m  uint    `json:"relative_humidity_2m"`
	ApparentTemperature float64 `json:"apparent_temperature"`
}

type HourlyDTO struct {
	Time                     []string  `json:"time"`
	WeatherCode              []uint    `json:"weather_code"`
	Temperature2m            []float64 `json:"temperature_2m"`
	WindSpeed10m             []float64 `json:"wind_speed_10m"`
	PrecipitationProbability []float64 `json:"precipitation_probability"`
}

type DailyDTO struct {
	Time                        []string  `json:"time"`
	WeatherCode                 []uint    `json:"weather_code"`
	Temperature2mMin            []float64 `json:"temperature_2m_min"`
	Temperature2mMax            []float64 `json:"temperature_2m_max"`
	WindSpeed10mMax             []float64 `json:"wind_speed_10m_max"`
	PrecipitationProbabilityMax []float64 `json:"precipitation_probability_max"`
}
