package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"frontdev333/weather-cli/internal/domain"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	geoCodingAPI = "https://geocoding-api.open-meteo.com/v1/search"
	openMeteoAPI = "https://api.open-meteo.com/v1/forecast"
)

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

type CurrentWeatherDTO struct {
	Temperature2m       float64 `json:"temperature_2m"`
	ApparentTemperature float64 `json:"apparent_temperature"`
	RelativeHumidity2m  uint    `json:"relative_humidity_2m"`
	WindSpeed10m        float64 `json:"wind_speed_10m"`
	WindDirection10m    uint    `json:"wind_direction_10m"`
	Precipitation       float64 `json:"precipitation"`
	PressureMsl         float64 `json:"pressure_msl"`
	WeatherCode         uint    `json:"weather_code"`
	Visibility          float64 `json:"visibility"`
}

type OpenMeteoWrapper struct {
	Current CurrentWeatherDTO `json:"current"`
}

func Geocode(city string) (name string, lat, lon float64, err error) {
	api, err := url.ParseRequestURI(geoCodingAPI)
	if err != nil {
		return "", 0, 0, err
	}

	params := url.Values{}
	params.Set("name", city)
	params.Set("count", "1")

	api.RawQuery = params.Encode()

	res, err := http.Get(api.String())
	if err != nil {
		return "", 0, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", 0, 0, errors.New(res.Status)
	}

	var dto GeocodeWrapper

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&dto)

	if err != nil {
		return "", 0, 0, err
	}

	if len(dto.Results) == 0 {
		return "", 0, 0, errors.New("city not found")
	}

	return fmt.Sprintf("%s, %s, %s", dto.Results[0].Name, dto.Results[0].Admin1, dto.Results[0].Country), dto.Results[0].Latitude, dto.Results[0].Longitude, err
}

func GetCurrentWeather(lat, lon float64) (domain.Today, error) {
	api, err := url.ParseRequestURI(openMeteoAPI)
	if err != nil {
		return domain.Today{}, err
	}

	latitude := strconv.FormatFloat(lat, 'f', -1, 64)
	longitude := strconv.FormatFloat(lon, 'f', -1, 64)

	params := url.Values{}
	params.Set("latitude", latitude)
	params.Set("longitude", longitude)
	params.Set("forecast_days", "1")
	params.Set("current", "temperature_2m,visibility,apparent_temperature,relative_humidity_2m,wind_speed_10m,wind_direction_10m,precipitation,pressure_msl,weather_code")

	api.RawQuery = params.Encode()

	res, err := http.Get(api.String())
	if err != nil {
		return domain.Today{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return domain.Today{}, errors.New(res.Status)
	}

	var dto OpenMeteoWrapper

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&dto)

	if err != nil {
		return domain.Today{}, err
	}

	toReturn := domain.Today{
		City:          "",
		Temperature:   dto.Current.Temperature2m,
		FeelsLike:     dto.Current.ApparentTemperature,
		Conditions:    weatherCodeToText(dto.Current.WeatherCode),
		WindSpeed:     dto.Current.WindSpeed10m,
		WindDirection: dto.Current.WindDirection10m,
		Humidity:      dto.Current.RelativeHumidity2m,
		Pressure:      dto.Current.PressureMsl,
		Visibility:    dto.Current.Visibility,
		Precipitation: dto.Current.Precipitation,
		UpdatedAt:     time.Now(),
	}
	return toReturn, nil
}

func weatherCodeToText(code uint) string {
	switch {
	case code == 0:
		return "Ясно"
	case code >= 1 && code <= 3:
		return "Преимущественно ясно / переменная облачность / пасмурно"
	case code == 45 || code == 48:
		return "Туман / туман с изморозью"
	case code >= 51 && code <= 57:
		return "Морось (в т.ч. ледяная)"
	case code >= 61 && code <= 67:
		return "Дождь (в т.ч. ледяной)"
	case code >= 71 && code <= 77:
		return "Снег / снежная крупа"
	case code >= 80 && code <= 82:
		return "Ливневый дождь"
	case code >= 85 && code <= 86:
		return "Снеговые ливни"
	case code == 95:
		return "Гроза"
	case code == 96 || code == 99:
		return "Гроза с градом"
	default:
		return "Неизвестные погодные условия"
	}
}
