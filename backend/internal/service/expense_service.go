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
	ErrNotFound     = errors.New("resource not found")
	ErrInvalidInput = errors.New("invalid input")
)

type ListExpensesInput struct {
	Limit  int
	Offset int
}

type SaveExpenseInput struct {
	Title       string
	Amount      float64
	Category    string
	Notes       string
	ExpenseDate time.Time
}

type ExpenseService interface {
	CreateExpense(ctx context.Context, input SaveExpenseInput) (*model.Expense, error)
	ListExpenses(ctx context.Context, input ListExpensesInput) ([]model.Expense, error)
	GetExpenseByID(ctx context.Context, id int64) (*model.Expense, error)
	UpdateExpense(ctx context.Context, id int64, input SaveExpenseInput) (*model.Expense, error)
	DeleteExpense(ctx context.Context, id int64) error
}

type expenseService struct {
	repo repository.ExpenseRepository
}

func NewExpenseService(repo repository.ExpenseRepository) ExpenseService {
	return &expenseService{repo: repo}
}

func (s *expenseService) CreateExpense(ctx context.Context, input SaveExpenseInput) (*model.Expense, error) {
	expense, err := buildExpenseModel(input)
	if err != nil {
		return nil, err
	}

	id, err := s.repo.Create(ctx, expense)
	if err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, id)
}

func (s *expenseService) ListExpenses(ctx context.Context, input ListExpensesInput) ([]model.Expense, error) {
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

func (s *expenseService) GetExpenseByID(ctx context.Context, id int64) (*model.Expense, error) {
	expense, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return expense, nil
}

func (s *expenseService) UpdateExpense(ctx context.Context, id int64, input SaveExpenseInput) (*model.Expense, error) {
	expense, err := buildExpenseModel(input)
	if err != nil {
		return nil, err
	}
	expense.ID = id

	err = s.repo.Update(ctx, expense)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return s.repo.GetByID(ctx, id)
}

func (s *expenseService) DeleteExpense(ctx context.Context, id int64) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func buildExpenseModel(input SaveExpenseInput) (*model.Expense, error) {
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

	date := input.ExpenseDate.UTC()
	if input.ExpenseDate.IsZero() {
		date = time.Now().UTC()
	}

	return &model.Expense{
		Title:       title,
		Amount:      input.Amount,
		Category:    category,
		Notes:       strings.TrimSpace(input.Notes),
		ExpenseDate: date,
	}, nil
}
