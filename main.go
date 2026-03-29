package main

import (
	"context"
	"fmt"
	"frontdev333/weather-cli/internal/provider/openmeteo"
	"frontdev333/weather-cli/internal/ui"
	"log"
)

func main() {
	client := openmeteo.NewClient()
	ctx := context.Background()

	hourly, err := client.GetHourly(ctx, "Paris", 12)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(ui.RenderHourly(hourly))

	daily, err := client.GetDaily(ctx, "Berlin", 7)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(ui.RenderDaily(daily))
}
