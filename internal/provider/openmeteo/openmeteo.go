package openmeteo

import (
	"context"
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

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	c := &http.Client{
		Timeout: 5 * time.Second,
	}

	return &Client{
		httpClient: c,
	}
}

func (c *Client) geocode(ctx context.Context, city string) (name string, lat, lon float64, err error) {
	api, err := url.ParseRequestURI(geoCodingAPI)
	if err != nil {
		return "", 0, 0, err
	}

	params := url.Values{}
	params.Set("name", city)
	params.Set("count", "1")

	api.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, api.String(), nil)
	if err != nil {
		return "", 0, 0, err
	}

	res, err := c.httpClient.Do(req)
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

func (c *Client) getCurrentWeather(ctx context.Context, lat, lon float64) (*ForecastResp, error) {
	api, err := url.ParseRequestURI(openMeteoAPI)
	if err != nil {
		return nil, err
	}

	latitude := strconv.FormatFloat(lat, 'f', -1, 64)
	longitude := strconv.FormatFloat(lon, 'f', -1, 64)

	params := url.Values{}
	params.Set("latitude", latitude)
	params.Set("longitude", longitude)
	params.Set("forecast_days", "1")
	params.Set("hourly", "temperature_2m,weather_code,wind_speed_10m,precipitation_probability")
	params.Set("daily", "temperature_2m_max,temperature_2m_min,precipitation_probability_max,wind_speed_10m_max,weather_code")
	params.Set("current", "temperature_2m,visibility,apparent_temperature,relative_humidity_2m,wind_speed_10m,wind_direction_10m,precipitation,pressure_msl,weather_code")

	api.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, api.String(), nil)
	if err != nil {
		return nil, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, errors.New(res.Status)
	}

	var dto ForecastResp

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&dto)

	if err != nil {
		return nil, err
	}

	return &dto, nil
}

func (c *Client) GetToday(ctx context.Context, city string) (domain.Today, error) {
	name, lat, lon, err := c.geocode(ctx, city)
	if err != nil {
		return domain.Today{}, err
	}

	forecast, err := c.getCurrentWeather(ctx, lat, lon)
	if err != nil {
		return domain.Today{}, err
	}

	return domain.Today{
		City:          name,
		Temperature:   forecast.CurrentDTO.Temperature2m,
		FeelsLike:     forecast.CurrentDTO.ApparentTemperature,
		Conditions:    weatherCodeToText(forecast.CurrentDTO.WeatherCode),
		WindSpeed:     forecast.CurrentDTO.WindSpeed10m,
		WindDirection: forecast.CurrentDTO.WindDirection10m,
		Humidity:      forecast.CurrentDTO.RelativeHumidity2m,
		Pressure:      forecast.CurrentDTO.PressureMsl,
		Visibility:    forecast.CurrentDTO.Visibility,
		Precipitation: forecast.CurrentDTO.Precipitation,
		UpdatedAt:     time.Now(),
	}, nil
}

// https://api.open-meteo.com/v1/forecast?latitude=52.52&longitude=13.41&daily=temperature_2m_max,temperature_2m_min,precipitation_probability_max,wind_speed_10m_max,weather_code&hourly=,temperature_2m,weather_code,wind_speed_10m,precipitation_probability&current=temperature_2m,apparent_temperature,relative_humidity_2m,wind_speed_10m,wind_direction_10m,precipitation,pressure_msl,weather_code&forecast_days=1

func (c *Client) GetHourly(ctx context.Context, city string, hours int) ([]domain.HourlyEntry, error) {
	return nil, nil
}

func (c *Client) GetDaily(ctx context.Context, city string, days int) ([]domain.DailyEntry, error) {
	return nil, nil
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
