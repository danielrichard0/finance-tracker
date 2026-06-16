package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"expense-tracker/backend/internal/model"
)

type TransactionRepository interface {
	Create(ctx context.Context, transaction *model.Transaction) (int64, error)
	List(ctx context.Context, limit, offset int) ([]model.Transaction, error)
	GetByID(ctx context.Context, id int64) (*model.Transaction, error)
	Update(ctx context.Context, transaction *model.Transaction) error
	Delete(ctx context.Context, id int64) error
}

type transactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, transaction *model.Transaction) (int64, error) {
	const query = `
		INSERT INTO transactions (title, amount, category, notes, transaction_date, type)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		transaction.Title,
		transaction.Amount,
		transaction.Category,
		nullableString(transaction.Notes),
		transaction.TransactionDate,
		transaction.Type,
	)
	if err != nil {
		return 0, fmt.Errorf("create transaction: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read insert id: %w", err)
	}

	return id, nil
}

func (r *transactionRepository) List(ctx context.Context, limit, offset int) ([]model.Transaction, error) {
	const query = `
		SELECT id, title, amount, category, notes, transaction_date, type, created_at, updated_at
		FROM transactions
		ORDER BY transaction_date DESC, id DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}
	defer rows.Close()

	transactions := make([]model.Transaction, 0, limit)
	for rows.Next() {
		transaction, scanErr := scanTransaction(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		transactions = append(transactions, *transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transactions: %w", err)
	}

	return transactions, nil
}

func (r *transactionRepository) GetByID(ctx context.Context, id int64) (*model.Transaction, error) {
	const query = `
		SELECT id, title, amount, category, notes, transaction_date, type, created_at, updated_at
		FROM transactions
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)
	transaction, err := scanTransaction(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("get transaction by id: %w", err)
	}

	return transaction, nil
}

func (r *transactionRepository) Update(ctx context.Context, transaction *model.Transaction) error {
	const query = `
		UPDATE transactions
		SET title = ?, amount = ?, category = ?, notes = ?, transaction_date = ?, type = ?, updated_at = NOW()
		WHERE id = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		transaction.Title,
		transaction.Amount,
		transaction.Category,
		nullableString(transaction.Notes),
		transaction.TransactionDate,
		transaction.Type,
		transaction.ID,
	)
	if err != nil {
		return fmt.Errorf("update transaction: %w", err)
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

func (r *transactionRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM transactions WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete transaction: %w", err)
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

func scanTransaction(s scanner) (*model.Transaction, error) {
	var transaction model.Transaction
	var notes sql.NullString

	err := s.Scan(
		&transaction.ID,
		&transaction.Title,
		&transaction.Amount,
		&transaction.Category,
		&notes,
		&transaction.TransactionDate,
		&transaction.Type,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if notes.Valid {
		transaction.Notes = notes.String
	}

	return &transaction, nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
