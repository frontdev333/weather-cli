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

	// Тест с контекстом
	today, err := client.GetToday(ctx, "Tokyo")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("City: %s\n", today.City)
	fmt.Printf("Temperature: %.1f°C (feels like %.1f°C)\n",
		today.Temperature, today.FeelsLike)
	fmt.Printf("Condition: %s\n", today.Conditions)
	fmt.Printf("Wind: %.1f m/s at %d°\n",
		today.WindSpeed, today.WindDirection)
	fmt.Printf("Humidity: %d%%\n", today.Humidity)
}
