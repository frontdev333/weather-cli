package openmeteo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"frontdev333/weather-cli/internal/domain"
	"frontdev333/weather-cli/internal/retry"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	geoCodingAPI = "https://geocoding-api.open-meteo.com/v1/search"
	openMeteoAPI = "https://api.open-meteo.com/v1/forecast"
)

var wmoCodes = map[uint]string{
	0:  "Ясно",
	1:  "Преимущественно ясно",
	2:  "Переменная облачность",
	3:  "Пасмурно",
	45: "Туман",
	48: "Туман с изморозью",
	51: "Легкая морось",
	53: "Морось",
	55: "Сильная морось",
	56: "Ледяная морось (легкая)",
	57: "Ледяная морось (сильная)",
	61: "Легкий дождь",
	63: "Дождь",
	65: "Сильный дождь",
	66: "Ледяной дождь (легкий)",
	67: "Ледяной дождь (сильный)",
	71: "Легкий снег",
	73: "Снег",
	75: "Сильный снег",
	77: "Снежная крупа",
	80: "Легкий ливневый дождь",
	81: "Ливневый дождь",
	82: "Сильный ливневый дождь",
	85: "Легкие снеговые ливни",
	86: "Снеговые ливни",
	95: "Гроза",
	96: "Гроза с градом (легкая)",
	99: "Гроза с градом (сильная)",
}

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

	var res *http.Response

	err = retry.Do(ctx, 2, 250*time.Millisecond, func() error {
		var req *http.Request

		req, err = http.NewRequestWithContext(ctx, http.MethodGet, api.String(), nil)
		if err != nil {
			return err
		}

		res, err = c.httpClient.Do(req)
		return err
	})

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
		return "", 0, 0, errors.New("город не найден")
	}

	return fmt.Sprintf("%s, %s, %s", dto.Results[0].Name, dto.Results[0].Admin1, dto.Results[0].Country), dto.Results[0].Latitude, dto.Results[0].Longitude, err
}

func (c *Client) getCurrentWeather(ctx context.Context, lat, lon float64, days, hours int) (*ForecastResp, error) {
	api, err := url.ParseRequestURI(openMeteoAPI)
	if err != nil {
		return nil, err
	}

	latitude := strconv.FormatFloat(lat, 'f', -1, 64)
	longitude := strconv.FormatFloat(lon, 'f', -1, 64)

	params := url.Values{}
	params.Set("latitude", latitude)
	params.Set("longitude", longitude)
	params.Set("timezone", "auto")
	params.Set("forecast_days", strconv.Itoa(days))
	params.Set("forecast_hours", strconv.Itoa(hours))
	params.Set("hourly", "temperature_2m,weather_code,wind_speed_10m,precipitation_probability")
	params.Set("daily", "temperature_2m_max,temperature_2m_min,precipitation_probability_max,wind_speed_10m_max,weather_code")
	params.Set("current", "temperature_2m,visibility,apparent_temperature,relative_humidity_2m,wind_speed_10m,wind_direction_10m,precipitation,pressure_msl,weather_code")

	api.RawQuery = params.Encode()

	var res *http.Response

	err = retry.Do(ctx, 2, 250*time.Millisecond, func() error {
		var req *http.Request

		req, err = http.NewRequestWithContext(ctx, http.MethodGet, api.String(), nil)
		if err != nil {
			return err
		}

		res, err = c.httpClient.Do(req)
		return err
	})
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

	forecast, err := c.getCurrentWeather(ctx, lat, lon, 1, 1)
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

func (c *Client) GetHourly(ctx context.Context, city string, hours int) ([]domain.HourlyEntry, error) {
	_, lat, lon, err := c.geocode(ctx, city)
	if err != nil {
		return nil, err
	}

	forecast, err := c.getCurrentWeather(ctx, lat, lon, 0, hours)
	if err != nil {
		return nil, err
	}

	res := make([]domain.HourlyEntry, len(forecast.HourlyDTO.Time))

	for i := 0; i < len(forecast.HourlyDTO.Time); i++ {

		timestamp, err := parseISOTime(forecast.HourlyDTO.Time[i])
		if err != nil {
			return nil, err
		}

		hour := domain.HourlyEntry{
			Temperature:              forecast.HourlyDTO.Temperature2m[i],
			Conditions:               weatherCodeToText(forecast.HourlyDTO.WeatherCode[i]),
			WindSpeed:                forecast.HourlyDTO.WindSpeed10m[i],
			PrecipitationProbability: forecast.HourlyDTO.PrecipitationProbability[i],
			Time:                     timestamp,
		}

		res[i] = hour
	}

	return res, nil
}

func (c *Client) GetDaily(ctx context.Context, city string, days int) ([]domain.DailyEntry, error) {
	city, lat, lon, err := c.geocode(ctx, city)
	if err != nil {
		return nil, err
	}

	forecast, err := c.getCurrentWeather(ctx, lat, lon, days, 0)
	if err != nil {
		return nil, err
	}

	res := make([]domain.DailyEntry, len(forecast.DailyDTO.Time))

	for i := 0; i < len(forecast.DailyDTO.Time); i++ {

		timestamp, err := parseISOTime(forecast.DailyDTO.Time[i])
		if err != nil {
			return nil, err
		}

		day := domain.DailyEntry{
			MinTemperature:           forecast.DailyDTO.Temperature2mMin[i],
			MaxTemperature:           forecast.DailyDTO.Temperature2mMax[i],
			PrecipitationProbability: forecast.DailyDTO.PrecipitationProbabilityMax[i],
			WindSpeed:                forecast.DailyDTO.WindSpeed10mMax[i],
			Conditions:               weatherCodeToText(forecast.DailyDTO.WeatherCode[i]),
			Date:                     timestamp,
		}

		res[i] = day
	}

	return res, nil
}

func parseISOTime(s string) (time.Time, error) {
	if strings.Contains(s, "T") {
		return time.Parse("2006-01-02T15:04", s)
	}

	return time.Parse("2006-01-02", s)
}

func weatherCodeToText(code uint) string {
	if text, ok := wmoCodes[code]; ok {
		return text
	}
	return "Неизвестные погодные условия"
}
