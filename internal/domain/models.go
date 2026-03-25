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
	Pressure      uint
	Visibility    uint
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
