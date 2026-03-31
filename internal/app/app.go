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

type application struct {
	config   *config.Config
	cache    *cache.TTLCache
	city     string
	provider provider.WeatherProvider
}

func Run() error {

	cfg, err := config.Load()
	openMeteoProvider := openmeteo.NewClient()
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

			if _, err = openMeteoProvider.GetDaily(context.Background(), city, 1); err != nil {
				fmt.Println("Город не найден. Проверьте на ошибки ввода.")
				continue
			}
			break
		}

		cfg = config.Config{city}
		if err = config.Save(cfg); err != nil {
			return err
		}
	}

	application := application{
		config:   &cfg,
		cache:    cache.New(),
		city:     cfg.DefaultCity,
		provider: openMeteoProvider,
	}

	return application.loop()
}

func (a *application) loop() error {
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
				fmt.Printf("Ошибка: %v\n", err)
				continue
			}
		case "2":
			if err = a.renderDaily(ctx); err != nil {
				fmt.Printf("Ошибка: %v\n", err)
				continue
			}
		case "c":
			if err = a.changeCity(ctx, scanner); err != nil {
				fmt.Printf("Ошибка: %v\n", err)
				continue
			}
		case "r":
			a.cache.Clear()
			if err = a.renderToday(ctx); err != nil {
				fmt.Printf("Ошибка: %v\n", err)
				continue
			}
		case "q":
			cancel()
			return nil
		default:
			fmt.Println("неизвестная команда.")
			if err = a.renderToday(ctx); err != nil {
				fmt.Printf("Ошибка: %v\n", err)
				continue
			}
		}
		fmt.Print(ui.RenderMenu())
	}
	return nil
}

func (a *application) renderToday(ctx context.Context) error {
	val, fetchedAt, isCached := a.cache.Get(a.city + ":today")
	if !isCached {
		t, err := a.provider.GetToday(ctx, a.city)
		if err != nil {
			return err
		}
		a.cache.Set(a.city+":today", t, 5*time.Minute)
		fmt.Print(ui.Header(a.city, isCached, t.UpdatedAt))
		fmt.Print(ui.RenderToday(t))
		return nil
	}

	t, ok := val.(domain.Today)
	if !ok {
		return errors.New("unable to get today forecast")
	}

	fmt.Print(ui.Header(a.city, isCached, fetchedAt))
	fmt.Print(ui.RenderToday(t))
	return nil
}

func (a *application) renderHourly(ctx context.Context) error {
	val, fethedAt, isCached := a.cache.Get(a.city + ":hourly")
	if !isCached {
		h, err := a.provider.GetHourly(ctx, a.city, 12)
		if err != nil {
			return err
		}
		a.cache.Set(a.city+":hourly", h, 15*time.Minute)
		fmt.Print(ui.Header(a.city, isCached, time.Now()))
		fmt.Print(ui.RenderHourly(h))
		return nil
	}

	h, ok := val.([]domain.HourlyEntry)
	if !ok {
		return errors.New("unable to get hourly forecast")
	}

	fmt.Print(ui.Header(a.city, isCached, fethedAt))
	fmt.Print(ui.RenderHourly(h))
	return nil
}

func (a *application) renderDaily(ctx context.Context) error {
	val, fetchedAt, isCached := a.cache.Get(a.city + ":daily")
	if !isCached {
		d, err := a.provider.GetDaily(ctx, a.city, 7)
		if err != nil {
			return err
		}
		a.cache.Set(a.city+":daily", d, 30*time.Minute)
		fmt.Print(ui.Header(a.city, isCached, time.Now()))
		fmt.Print(ui.RenderDaily(d))
		return nil
	}

	d, ok := val.([]domain.DailyEntry)
	if !ok {
		return errors.New("unable to get daily forecast")
	}

	fmt.Print(ui.Header(a.city, isCached, fetchedAt))
	fmt.Print(ui.RenderDaily(d))
	return nil
}

func (a *application) changeCity(ctx context.Context, scanner *bufio.Scanner) error {
	fmt.Println("Введите новый город: ")
	for scanner.Scan() {
		city := strings.TrimSpace(scanner.Text())
		if len(city) == 0 {
			fmt.Println("Город не может быть пустым.")
			continue
		}

		if _, err := a.provider.GetToday(ctx, city); err != nil {
			fmt.Println("Город введен некорректно.")
			continue
		}

		a.cache.Clear()
		a.city = city
		a.config.DefaultCity = city
		return config.Save(*a.config)
	}
	return nil
}
