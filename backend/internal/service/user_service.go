package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"expense-tracker/backend/internal/model"
	"expense-tracker/backend/internal/repository"
)

type ListUsersInput struct {
	Limit  int
	Offset int
}

type SaveUserInput struct {
	Email    string
	Phone    string
	Username string
}

type UserService interface {
	CreateUser(ctx context.Context, input SaveUserInput) (*model.User, error)
	ListUsers(ctx context.Context, input ListUsersInput) ([]model.User, error)
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	UpdateUser(ctx context.Context, id string, input SaveUserInput) (*model.User, error)
	DeleteUser(ctx context.Context, id string) error
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, input SaveUserInput) (*model.User, error) {
	user, err := buildUserModel(input)
	if err != nil {
		return nil, err
	}

	id, err := s.repo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, id)
}

func (s *userService) ListUsers(ctx context.Context, input ListUsersInput) ([]model.User, error) {
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

func (s *userService) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, id string, input SaveUserInput) (*model.User, error) {
	user, err := buildUserModel(input)
	if err != nil {
		return nil, err
	}
	user.UserID = id

	err = s.repo.Update(ctx, user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return s.repo.GetByID(ctx, id)
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func buildUserModel(input SaveUserInput) (*model.User, error) {
	email := strings.TrimSpace(input.Email)
	username := strings.TrimSpace(input.Username)

	if email == "" || username == "" {
		return nil, ErrInvalidInput
	}

	return &model.User{
		Email:    email,
		Phone:    strings.TrimSpace(input.Phone),
		Username: username,
	}, nil
}
