package discordbot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"expense-tracker/backend/internal/transactionapi"

	"github.com/bwmarrin/discordgo"
)

const interactionTimeout = 10 * time.Second

var transactionTypeChoices = []*discordgo.ApplicationCommandOptionChoice{
	{Name: "Expense", Value: "E"},
	{Name: "Income", Value: "I"},
}

var transactionCommand = &discordgo.ApplicationCommand{
	Name:        "transaction",
	Description: "Manage transactions",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "add",
			Description: "Add a new transaction",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "title",
					Description: "Transaction title",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "amount",
					Description: "Amount, for example 12.50",
					Required:    true,
					MinValue:    floatPtr(0.01),
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "type",
					Description: "Expense or income",
					Required:    true,
					Choices:     transactionTypeChoices,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "category",
					Description: "Optional category",
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "notes",
					Description: "Optional notes",
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "transaction_date",
					Description: "Optional date in YYYY-MM-DD format",
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "list",
			Description: "List recent transactions",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "limit",
					Description: "Number of transactions to show",
					MinValue:    floatPtr(1),
					MaxValue:    20,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "get",
			Description: "Get one transaction by ID",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "id",
					Description: "Transaction ID",
					Required:    true,
					MinValue:    floatPtr(1),
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "update",
			Description: "Update a transaction",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "id",
					Description: "Transaction ID",
					Required:    true,
					MinValue:    floatPtr(1),
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "title",
					Description: "New title",
				},
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "amount",
					Description: "New amount, for example 12.50",
					MinValue:    floatPtr(0.01),
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "type",
					Description: "New expense or income type",
					Choices:     transactionTypeChoices,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "category",
					Description: "New category",
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "notes",
					Description: "New notes",
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "transaction_date",
					Description: "New date in YYYY-MM-DD format",
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "delete",
			Description: "Delete a transaction",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "id",
					Description: "Transaction ID",
					Required:    true,
					MinValue:    floatPtr(1),
				},
			},
		},
	},
}

func registerCommands(session *discordgo.Session, appID string, guildID string) error {
	_, err := session.ApplicationCommandCreate(appID, guildID, transactionCommand)
	return err
}

func (b *Bot) onInteractionCreate(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if interaction.Type != discordgo.InteractionApplicationCommand {
		return
	}

	if interaction.ApplicationCommandData().Name != "transaction" {
		return
	}

	if err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}); err != nil {
		fmt.Printf("failed to respond to interaction: %v\n", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), interactionTimeout)
	defer cancel()

	content, err := b.handleTransactionCommand(ctx, interaction.ApplicationCommandData())
	if err != nil {
		content = fmt.Sprintf("Request failed: %s", err.Error())
	}

	if _, err := session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	}); err != nil {
		fmt.Printf("failed to edit interaction response: %v\n", err)
	}
}

func (b *Bot) handleTransactionCommand(ctx context.Context, data discordgo.ApplicationCommandInteractionData) (string, error) {
	if len(data.Options) == 0 {
		return "Choose a transaction subcommand.", nil
	}

	subcommand := data.Options[0]
	options := optionMap(subcommand.Options)

	switch subcommand.Name {
	case "add":
		return b.addTransaction(ctx, options)
	case "list":
		return b.listTransactions(ctx, options)
	case "get":
		return b.getTransaction(ctx, options)
	case "update":
		return b.updateTransaction(ctx, options)
	case "delete":
		return b.deleteTransaction(ctx, options)
	default:
		return "Unknown transaction subcommand.", nil
	}
}

func (b *Bot) addTransaction(ctx context.Context, options map[string]*discordgo.ApplicationCommandInteractionDataOption) (string, error) {
	title := strings.TrimSpace(stringOption(options, "title"))
	if title == "" {
		return "Title cannot be empty.", nil
	}

	transactionDate, err := optionalTransactionDate(options, "transaction_date")
	if err != nil {
		return "", err
	}

	payload := transactionapi.SaveTransactionRequest{
		Title:           title,
		Amount:          numberOption(options, "amount"),
		Category:        stringValue(optionalString(options, "category")),
		Notes:           stringValue(optionalString(options, "notes")),
		TransactionDate: transactionDate,
		Type:            stringOption(options, "type"),
	}

	transaction, err := b.apiClient.CreateTransaction(ctx, payload)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Added transaction #%d: %s - %.2f (%s)", transaction.ID, transaction.Title, transaction.Amount, transactionTypeLabel(transaction.Type)), nil
}

func (b *Bot) listTransactions(ctx context.Context, options map[string]*discordgo.ApplicationCommandInteractionDataOption) (string, error) {
	limit := 5
	if option, ok := options["limit"]; ok {
		limit = int(option.IntValue())
	}

	transactions, err := b.apiClient.ListTransactions(ctx, limit, 0)
	if err != nil {
		return "", err
	}

	if len(transactions) == 0 {
		return "No transactions found.", nil
	}

	lines := make([]string, 0, len(transactions)+1)
	lines = append(lines, "Recent transactions:")
	for _, transaction := range transactions {
		lines = append(lines, fmt.Sprintf(
			"#%d [%s/%s] %s - %.2f on %s",
			transaction.ID,
			transactionTypeLabel(transaction.Type),
			transaction.Category,
			transaction.Title,
			transaction.Amount,
			transaction.TransactionDate.Format("2006-01-02"),
		))
	}

	return strings.Join(lines, "\n"), nil
}

func (b *Bot) getTransaction(ctx context.Context, options map[string]*discordgo.ApplicationCommandInteractionDataOption) (string, error) {
	transaction, err := b.apiClient.GetTransaction(ctx, int64Option(options, "id"))
	if err != nil {
		return "", err
	}

	return formatTransaction(transaction), nil
}

func (b *Bot) updateTransaction(ctx context.Context, options map[string]*discordgo.ApplicationCommandInteractionDataOption) (string, error) {
	id := int64Option(options, "id")
	current, err := b.apiClient.GetTransaction(ctx, id)
	if err != nil {
		return "", err
	}

	payload := transactionapi.SaveTransactionRequest{
		Title:           current.Title,
		Amount:          current.Amount,
		Category:        current.Category,
		Notes:           current.Notes,
		TransactionDate: current.TransactionDate.Format("2006-01-02"),
		Type:            current.Type,
	}

	changed := false
	if value := optionalString(options, "title"); value != nil {
		payload.Title = *value
		changed = true
	}
	if option, ok := options["amount"]; ok {
		payload.Amount = option.FloatValue()
		changed = true
	}
	if option, ok := options["type"]; ok {
		payload.Type = option.StringValue()
		changed = true
	}
	if value := optionalString(options, "category"); value != nil {
		payload.Category = *value
		changed = true
	}
	if value := optionalString(options, "notes"); value != nil {
		payload.Notes = *value
		changed = true
	}
	if _, ok := options["transaction_date"]; ok {
		transactionDate, err := optionalTransactionDate(options, "transaction_date")
		if err != nil {
			return "", err
		}
		payload.TransactionDate = transactionDate
		changed = true
	}

	if !changed {
		return "Provide at least one field to update.", nil
	}

	transaction, err := b.apiClient.UpdateTransaction(ctx, id, payload)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Updated transaction #%d: %s - %.2f (%s)", transaction.ID, transaction.Title, transaction.Amount, transactionTypeLabel(transaction.Type)), nil
}

func (b *Bot) deleteTransaction(ctx context.Context, options map[string]*discordgo.ApplicationCommandInteractionDataOption) (string, error) {
	id := int64Option(options, "id")
	if err := b.apiClient.DeleteTransaction(ctx, id); err != nil {
		return "", err
	}

	return fmt.Sprintf("Deleted transaction #%d.", id), nil
}

func formatTransaction(transaction transactionapi.Transaction) string {
	lines := []string{
		fmt.Sprintf("Transaction #%d", transaction.ID),
		fmt.Sprintf("Title: %s", transaction.Title),
		fmt.Sprintf("Amount: %.2f", transaction.Amount),
		fmt.Sprintf("Type: %s", transactionTypeLabel(transaction.Type)),
		fmt.Sprintf("Category: %s", transaction.Category),
		fmt.Sprintf("Date: %s", transaction.TransactionDate.Format("2006-01-02")),
	}

	if transaction.Notes != "" {
		lines = append(lines, fmt.Sprintf("Notes: %s", transaction.Notes))
	}

	return strings.Join(lines, "\n")
}

func transactionTypeLabel(value string) string {
	switch strings.ToUpper(value) {
	case "E":
		return "Expense"
	case "I":
		return "Income"
	default:
		return value
	}
}

func optionMap(options []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	result := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, option := range options {
		result[option.Name] = option
	}
	return result
}

func stringOption(options map[string]*discordgo.ApplicationCommandInteractionDataOption, name string) string {
	option, ok := options[name]
	if !ok {
		return ""
	}
	return option.StringValue()
}

func optionalString(options map[string]*discordgo.ApplicationCommandInteractionDataOption, name string) *string {
	value := strings.TrimSpace(stringOption(options, name))
	if value == "" {
		return nil
	}
	return &value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func numberOption(options map[string]*discordgo.ApplicationCommandInteractionDataOption, name string) float64 {
	option, ok := options[name]
	if !ok {
		return 0
	}
	return option.FloatValue()
}

func int64Option(options map[string]*discordgo.ApplicationCommandInteractionDataOption, name string) int64 {
	option, ok := options[name]
	if !ok {
		return 0
	}
	return option.IntValue()
}

func optionalTransactionDate(options map[string]*discordgo.ApplicationCommandInteractionDataOption, name string) (string, error) {
	value := stringValue(optionalString(options, name))
	if value == "" {
		return "", nil
	}

	if _, err := time.Parse("2006-01-02", value); err != nil {
		return "", fmt.Errorf("%s must use YYYY-MM-DD format", name)
	}

	return value, nil
}

func floatPtr(value float64) *float64 {
	return &value
}
