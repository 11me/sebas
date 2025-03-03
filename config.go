package main

import (
	"fmt"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	Binance `envPrefix:"BINANCE_"`
	TgBot   `envPrefix:"TG_BOT_"`
}

type Binance struct {
	Key    string `env:"API_KEY"`
	Secret string `env:"API_SECRET"`
}

type TgBot struct {
	Token              string `env:"TOKEN,notEmpty"`
	DelistingChannelID string `env:"DELISTING_CHANNEL_ID,notEmpty"`
}

func NewConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to read .env: %w", err)
	}

	conf := &Config{}
	if err := env.Parse(conf); err != nil {
		return nil, err
	}

	return conf, nil
}
