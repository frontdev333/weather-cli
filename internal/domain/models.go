package domain

import "time"

type Today struct {
	City          string
	Temperature   float64
	FeelsLike     float64
	Conditions    string
	WindSpeed     float64
	WindDirection uint
	Humidity      uint
	Pressure      float64
	Visibility    float64
	Precipitation float64
	UpdatedAt     time.Time
}

type HourlyEntry struct {
	Temperature              float64
	Conditions               string
	WindSpeed                float64
	PrecipitationProbability float64
	UpdatedAt                time.Time
}

type DailyEntry struct {
	MinTemperature           float64
	MaxTemperature           float64
	PrecipitationProbability float64
	WindSpeed                float64
	Conditions               string
	Date                     time.Time
}

// https://api.open-meteo.com/v1/forecast?latitude=52.52&longitude=13.41&daily=temperature_2m_min,temperature_2m_max,precipitation_probability_max,wind_speed_10m_max,weather_code&forecast_days=1
