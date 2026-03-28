package main

import (
	"fmt"
	"frontdev333/weather-cli/internal/domain"
	"frontdev333/weather-cli/internal/ui"
	"time"
)

func main() {
	today := domain.Today{
		City:          "Moscow, Moscow, Russia",
		Temperature:   -5.2,
		FeelsLike:     -9.1,
		Conditions:    "Снег",
		WindSpeed:     3.5,
		WindDirection: 180,
		Humidity:      85,
		Pressure:      1015,
		Precipitation: 2.3,
		UpdatedAt:     time.Now().Add(-3 * time.Minute),
	}

	fmt.Print(ui.Header("Moscow", true, today.UpdatedAt))
	fmt.Print(ui.RenderToday(today))
	fmt.Print(ui.RenderMenu())
}
