package config

import (
	"errors"
	"strings"

	"github.com/joho/godotenv"
)

const defaultTransactionAPIBaseURL = "http://127.0.0.1:8080/api/v1"

type BotConfig struct {
	DiscordToken          string
	DiscordGuildID        string
	DiscordAppID          string
	TransactionAPIBaseURL string
}

func LoadBot() (BotConfig, error) {
	_ = godotenv.Load()

	cfg := BotConfig{
		DiscordToken:          strings.TrimSpace(getEnv("DISCORD_TOKEN", "")),
		DiscordGuildID:        strings.TrimSpace(getEnv("DISCORD_GUILD_ID", "")),
		DiscordAppID:          strings.TrimSpace(getEnv("DISCORD_APP_ID", "")),
		TransactionAPIBaseURL: strings.TrimRight(strings.TrimSpace(getEnv("TRANSACTION_API_BASE_URL", defaultTransactionAPIBaseURL)), "/"),
	}

	if cfg.DiscordToken == "" {
		return BotConfig{}, errors.New("DISCORD_TOKEN is required")
	}

	return cfg, nil
}
