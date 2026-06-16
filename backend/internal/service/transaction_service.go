package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"expense-tracker/backend/internal/model"
	"expense-tracker/backend/internal/repository"
)

var (
	ErrNotFound               = errors.New("resource not found")
	ErrInvalidInput           = errors.New("invalid input")
	ErrInvalidTransactionType = errors.New("invalid transaction type")
)

type ListTransactionsInput struct {
	Limit  int
	Offset int
}

type SaveTransactionInput struct {
	Title           string
	Amount          float64
	Category        string
	Notes           string
	TransactionDate time.Time
	Type            string
}

type TransactionService interface {
	CreateTransaction(ctx context.Context, input SaveTransactionInput) (*model.Transaction, error)
	ListTransactions(ctx context.Context, input ListTransactionsInput) ([]model.Transaction, error)
	GetTransactionByID(ctx context.Context, id int64) (*model.Transaction, error)
	UpdateTransaction(ctx context.Context, id int64, input SaveTransactionInput) (*model.Transaction, error)
	DeleteTransaction(ctx context.Context, id int64) error
}

type transactionService struct {
	repo repository.TransactionRepository
}

func NewTransactionService(repo repository.TransactionRepository) TransactionService {
	return &transactionService{repo: repo}
}

func (s *transactionService) CreateTransaction(ctx context.Context, input SaveTransactionInput) (*model.Transaction, error) {
	transaction, err := buildTransactionModel(input)
	if err != nil {
		return nil, err
	}

	id, err := s.repo.Create(ctx, transaction)
	if err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, id)
}

func (s *transactionService) ListTransactions(ctx context.Context, input ListTransactionsInput) ([]model.Transaction, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := input.Offset
	if offset < 0 {
		offset = 0
	}

	return s.repo.List(ctx, limit, offset)
}

func (s *transactionService) GetTransactionByID(ctx context.Context, id int64) (*model.Transaction, error) {
	transaction, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return transaction, nil
}

func (s *transactionService) UpdateTransaction(ctx context.Context, id int64, input SaveTransactionInput) (*model.Transaction, error) {
	transaction, err := buildTransactionModel(input)
	if err != nil {
		return nil, err
	}
	transaction.ID = id

	err = s.repo.Update(ctx, transaction)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return s.repo.GetByID(ctx, id)
}

func (s *transactionService) DeleteTransaction(ctx context.Context, id int64) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func buildTransactionModel(input SaveTransactionInput) (*model.Transaction, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, ErrInvalidInput
	}
	if input.Amount <= 0 {
		return nil, ErrInvalidInput
	}

	category := strings.TrimSpace(input.Category)
	if category == "" {
		category = "general"
	}

	date := input.TransactionDate.UTC()
	if input.TransactionDate.IsZero() {
		date = time.Now().UTC()
	}

	transactionType, err := normalizeTransactionType(input.Type)
	if err != nil {
		return nil, err
	}

	return &model.Transaction{
		Title:           title,
		Amount:          input.Amount,
		Category:        category,
		Notes:           strings.TrimSpace(input.Notes),
		TransactionDate: date,
		Type:            transactionType,
	}, nil
}

func normalizeTransactionType(value string) (string, error) {
	transactionType := strings.ToUpper(strings.TrimSpace(value))
	switch transactionType {
	case "E", "I":
		return transactionType, nil
	default:
		return "", ErrInvalidTransactionType
	}
}
