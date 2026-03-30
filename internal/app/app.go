package app

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"frontdev333/weather-cli/internal/cache"
	"frontdev333/weather-cli/internal/config"
	"frontdev333/weather-cli/internal/domain"
	"frontdev333/weather-cli/internal/provider"
	"frontdev333/weather-cli/internal/provider/openmeteo"
	"frontdev333/weather-cli/internal/ui"
	"os"
	"strings"
	"time"
)

type Application struct {
	Config   *config.Config
	Cache    *cache.TTLCache
	City     string
	Provider provider.WeatherProvider
}

func Run() error {

	cfg, err := config.Load()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}

		var city string

		fmt.Println("Введите город по умолчанию: ")
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			city = strings.TrimSpace(scanner.Text())
			if len(city) == 0 {
				fmt.Println("Город не может быть пустым.")
				continue
			}
			break
		}

		cfg = config.Config{city}
		if err = config.Save(cfg); err != nil {
			return err
		}
	}

	openMeteoProvider := openmeteo.NewClient()

	application := Application{
		Config:   &cfg,
		Cache:    cache.New(),
		City:     cfg.DefaultCity,
		Provider: openMeteoProvider,
	}

	return application.loop()
}

func (a *Application) loop() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Print("\033[H\033[2J")
	err := a.renderToday(ctx)
	if err != nil {
		return err
	}
	fmt.Print(ui.RenderMenu())
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Print("\033[H\033[2J")
		switch strings.ToLower(strings.TrimSpace(scanner.Text())) {
		case "1":
			if err = a.renderHourly(ctx); err != nil {
				return err
			}
		case "2":
			if err = a.renderDaily(ctx); err != nil {
				return err
			}
		case "c":
			if err = a.changeCity(ctx, scanner); err != nil {
				return err
			}
		case "r":
			a.Cache.Clear()
			if err = a.renderToday(ctx); err != nil {
				return err
			}
		case "q":
			cancel()
			return nil
		default:
			fmt.Println("неизвестная команда.")
			if err = a.renderToday(ctx); err != nil {
				return err
			}
		}
		fmt.Print(ui.RenderMenu())
	}
	return nil
}

func (a *Application) renderToday(ctx context.Context) error {
	val, fetchedAt, isCached := a.Cache.Get(a.City + ":today")
	if !isCached {
		t, err := a.Provider.GetToday(ctx, a.City)
		if err != nil {
			return err
		}
		a.Cache.Set(a.City+":today", t, 5*time.Minute)
		fmt.Print(ui.Header(a.City, isCached, t.UpdatedAt))
		fmt.Print(ui.RenderToday(t))
		return nil
	}

	t, ok := val.(domain.Today)
	if !ok {
		return errors.New("unable to get today forecast")
	}

	fmt.Print(ui.Header(a.City, isCached, fetchedAt))
	fmt.Print(ui.RenderToday(t))
	return nil
}

func (a *Application) renderHourly(ctx context.Context) error {
	val, fethedAt, isCached := a.Cache.Get(a.City + ":hourly")
	if !isCached {
		h, err := a.Provider.GetHourly(ctx, a.City, 12)
		if err != nil {
			return err
		}
		a.Cache.Set(a.City+":hourly", h, 15*time.Minute)
		fmt.Print(ui.Header(a.City, isCached, time.Now()))
		fmt.Print(ui.RenderHourly(h))
		return nil
	}

	h, ok := val.([]domain.HourlyEntry)
	if !ok {
		return errors.New("unable to get hourly forecast")
	}

	fmt.Print(ui.Header(a.City, isCached, fethedAt))
	fmt.Print(ui.RenderHourly(h))
	return nil
}

func (a *Application) renderDaily(ctx context.Context) error {
	val, fetchedAt, isCached := a.Cache.Get(a.City + ":daily")
	if !isCached {
		d, err := a.Provider.GetDaily(ctx, a.City, 7)
		if err != nil {
			return err
		}
		a.Cache.Set(a.City+":daily", d, 5*time.Minute)
		fmt.Print(ui.Header(a.City, isCached, time.Now()))
		fmt.Print(ui.RenderDaily(d))
		return nil
	}

	d, ok := val.([]domain.DailyEntry)
	if !ok {
		return errors.New("unable to get daily forecast")
	}

	fmt.Print(ui.Header(a.City, isCached, fetchedAt))
	fmt.Print(ui.RenderDaily(d))
	return nil
}

func (a *Application) changeCity(ctx context.Context, scanner *bufio.Scanner) error {
	fmt.Println("Введите новый город: ")
	for scanner.Scan() {
		city := strings.TrimSpace(scanner.Text())
		if len(city) == 0 {
			fmt.Println("Город не может быть пустым.")
			return nil
		}

		if _, err := a.Provider.GetToday(ctx, city); err != nil {
			fmt.Println("Город введен некорректно.")
			return nil
		}

		a.Cache.Clear()
		a.City = city
		a.Config.DefaultCity = city
		return config.Save(*a.Config)
	}
	return nil
}
