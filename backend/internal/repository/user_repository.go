package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"expense-tracker/backend/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) (string, error)
	List(ctx context.Context, limit, offset int) ([]model.User, error)
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id string) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) (string, error) {
	const query = `
		INSERT INTO users (email, phone, username)
		VALUES (?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.Email,
		nullableString(user.Phone),
		user.Username,
	)
	if err != nil {
		return "", fmt.Errorf("create user: %w", err)
	}

	fmt.Println(user.Email)
	id, err := r.GetByEmail(ctx, user.Email)

	if err != nil {
		return "", fmt.Errorf("read insert id: %w", err)
	}

	return id.UserID, nil
}

func (r *userRepository) List(ctx context.Context, limit, offset int) ([]model.User, error) {
	const query = `
		SELECT user_id, email, phone, username, created_at, updated_at
		FROM users
		ORDER BY created_at DESC, user_id DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]model.User, 0, limit)
	for rows.Next() {
		user, scanErr := scanUser(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		users = append(users, *user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}

	return users, nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	const query = `
		SELECT user_id, email, phone, username, created_at, updated_at
		FROM users
		WHERE user_id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	const query = `
		SELECT user_id, email, phone, username, created_at, updated_at
		FROM users
		WHERE email = ?
	`

	row := r.db.QueryRowContext(ctx, query, email)
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	const query = `
		UPDATE users
		SET email = ?, phone = ?, username = ?, updated_at = NOW()
		WHERE user_id = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		user.Email,
		nullableString(user.Phone),
		user.Username,
		user.UserID,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
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

func (r *userRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM users WHERE user_id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
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

func scanUser(s scanner) (*model.User, error) {
	var user model.User
	var userID string
	var phone sql.NullString
	var email sql.NullString
	var username sql.NullString

	err := s.Scan(
		&userID,
		&email,
		&phone,
		&username,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	user.UserID = userID
	if email.Valid {
		user.Email = email.String
	}
	if phone.Valid {
		user.Phone = phone.String
	}
	if username.Valid {
		user.Username = username.String
	}

	return &user, nil
}
