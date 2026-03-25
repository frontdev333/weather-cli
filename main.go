package main

import (
	"fmt"
	"frontdev333/weather-cli/internal/config"
	"log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Default city: %s\n", cfg.DefaultCity)
}
