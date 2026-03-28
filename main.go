package main

import (
	"context"
	"fmt"
	"frontdev333/weather-cli/internal/provider/openmeteo"
	"log"
)

func main() {
	client := openmeteo.NewClient()
	ctx := context.Background()

	hourly, err := client.GetHourly(ctx, "Paris", 12)
	if err != nil {
		log.Fatal(err)
	}
	for _, h := range hourly {
		fmt.Printf("%s: %.1f°C, precipitation %.0f%%\n",
			h.Time.Format("15:04"), h.Temperature, h.PrecipitationProbability)
	}

	daily, err := client.GetDaily(ctx, "Berlin", 7)
	if err != nil {
		log.Fatal(err)
	}
	for _, d := range daily {
		fmt.Printf("%s: %.1f°C - %.1f°C, %s\n",
			d.Date.Format("Jan 02"), d.MinTemperature, d.MaxTemperature, d.Conditions)
	}
}
