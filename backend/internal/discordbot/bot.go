package discordbot

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"expense-tracker/backend/internal/config"
	"expense-tracker/backend/internal/transactionapi"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	cfg          config.BotConfig
	apiClient    *transactionapi.Client
	session      *discordgo.Session
	registerOnce sync.Once
}

func New(cfg config.BotConfig, apiClient *transactionapi.Client) *Bot {
	return &Bot{
		cfg:       cfg,
		apiClient: apiClient,
	}
}

func (b *Bot) Run() error {
	session, err := discordgo.New("Bot " + b.cfg.DiscordToken)
	if err != nil {
		return err
	}

	b.session = session
	b.session.AddHandler(b.onReady)
	b.session.AddHandler(b.onInteractionCreate)

	if err := b.session.Open(); err != nil {
		return err
	}
	defer b.session.Close()

	log.Println("Discord bot is running. Press Ctrl+C to stop.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Discord bot stopped.")
	return nil
}

func (b *Bot) onReady(session *discordgo.Session, ready *discordgo.Ready) {
	log.Printf("Logged in as %s#%s", ready.User.Username, ready.User.Discriminator)

	b.registerOnce.Do(func() {
		appID := b.cfg.DiscordAppID
		if appID == "" {
			appID = ready.User.ID
		}

		if err := registerCommands(session, appID, b.cfg.DiscordGuildID); err != nil {
			log.Printf("Failed to register commands: %v", err)
			return
		}

		if b.cfg.DiscordGuildID == "" {
			log.Println("Registered global slash commands")
			return
		}

		log.Printf("Registered slash commands for guild %s", b.cfg.DiscordGuildID)
	})
}
