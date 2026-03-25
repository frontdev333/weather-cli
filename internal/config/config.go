package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DefaultCity string `yaml:"default_city"`
}

func Load() (Config, error) {
	bytes, err := os.ReadFile(getConfigPath())
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return Config{}, err
		}

		cfg := Config{
			DefaultCity: "Moscow",
		}

		if err = Save(cfg); err != nil {
			return Config{}, err
		}

		return cfg, nil
	}

	var cfg Config

	if err = yaml.Unmarshal(bytes, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func Save(cfg Config) error {
	bytes, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(getConfigPath(), bytes, 0644)
}

func getConfigPath() string {
	return "config.yaml"
}
