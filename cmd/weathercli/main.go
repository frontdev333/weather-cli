package main

import (
	"frontdev333/weather-cli/internal/app"
	"log/slog"
)

func main() {
	err := app.Run()
	if err != nil {
		slog.Error(err.Error())
		return
	}
}
