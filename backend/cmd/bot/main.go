package main

import (
	"log"

	"expense-tracker/backend/internal/config"
	"expense-tracker/backend/internal/discordbot"
	"expense-tracker/backend/internal/transactionapi"
)

func main() {
	cfg, err := config.LoadBot()
	if err != nil {
		log.Fatal(err)
	}

	apiClient := transactionapi.NewClient(cfg.TransactionAPIBaseURL)
	bot := discordbot.New(cfg, apiClient)

	if err := bot.Run(); err != nil {
		log.Fatal(err)
	}
}
