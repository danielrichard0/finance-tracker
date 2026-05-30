package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"expense-tracker/backend/internal/model"
)

type ExpenseRepository interface {
	Create(ctx context.Context, expense *model.Expense) (int64, error)
	List(ctx context.Context, limit, offset int) ([]model.Expense, error)
	GetByID(ctx context.Context, id int64) (*model.Expense, error)
	Update(ctx context.Context, expense *model.Expense) error
	Delete(ctx context.Context, id int64) error
}

type expenseRepository struct {
	db *sql.DB
}

func NewExpenseRepository(db *sql.DB) ExpenseRepository {
	return &expenseRepository{db: db}
}

func (r *expenseRepository) Create(ctx context.Context, expense *model.Expense) (int64, error) {
	const query = `
		INSERT INTO expenses (title, amount, category, notes, expense_date)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		expense.Title,
		expense.Amount,
		expense.Category,
		nullableString(expense.Notes),
		expense.ExpenseDate,
	)
	if err != nil {
		return 0, fmt.Errorf("create expense: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read insert id: %w", err)
	}

	return id, nil
}

func (r *expenseRepository) List(ctx context.Context, limit, offset int) ([]model.Expense, error) {
	const query = `
		SELECT id, title, amount, category, notes, expense_date, created_at, updated_at
		FROM expenses
		ORDER BY expense_date DESC, id DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list expenses: %w", err)
	}
	defer rows.Close()

	expenses := make([]model.Expense, 0, limit)
	for rows.Next() {
		expense, scanErr := scanExpense(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		expenses = append(expenses, *expense)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate expenses: %w", err)
	}

	return expenses, nil
}

func (r *expenseRepository) GetByID(ctx context.Context, id int64) (*model.Expense, error) {
	const query = `
		SELECT id, title, amount, category, notes, expense_date, created_at, updated_at
		FROM expenses
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)
	expense, err := scanExpense(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("get expense by id: %w", err)
	}

	return expense, nil
}

func (r *expenseRepository) Update(ctx context.Context, expense *model.Expense) error {
	const query = `
		UPDATE expenses
		SET title = ?, amount = ?, category = ?, notes = ?, expense_date = ?, updated_at = NOW()
		WHERE id = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		expense.Title,
		expense.Amount,
		expense.Category,
		nullableString(expense.Notes),
		expense.ExpenseDate,
		expense.ID,
	)
	if err != nil {
		return fmt.Errorf("update expense: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read updated rows: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *expenseRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM expenses WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete expense: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted rows: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanExpense(s scanner) (*model.Expense, error) {
	var expense model.Expense
	var notes sql.NullString

	err := s.Scan(
		&expense.ID,
		&expense.Title,
		&expense.Amount,
		&expense.Category,
		&notes,
		&expense.ExpenseDate,
		&expense.CreatedAt,
		&expense.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if notes.Valid {
		expense.Notes = notes.String
	}

	return &expense, nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
