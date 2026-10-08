package service

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode"

	"todo/internal/model"
	"todo/internal/repository"
)

var (
	// ErrEmptyTitle is returned when creating or updating a todo with an empty title.
	ErrEmptyTitle = errors.New("title is required")
)

// Service defines the business logic operations for Todo items.
type Service interface {
	CreateTodo(ctx context.Context, req model.CreateTodoRequest) (*model.Todo, error)
	GetTodo(ctx context.Context, id int64) (*model.Todo, error)
	GetTodos(ctx context.Context) ([]model.Todo, error)
	UpdateTodo(ctx context.Context, id int64, req model.UpdateTodoRequest) (*model.Todo, error)
	DeleteTodo(ctx context.Context, id int64) error
}

type todoService struct {
	repo repository.Repository
}

// NewService creates a new Service instance backed by the provided repository.
func NewService(repo repository.Repository) Service {
	return &todoService{repo: repo}
}

// CreateTodo validates input, sets timestamps, and persists a new todo item.
func (s *todoService) CreateTodo(ctx context.Context, req model.CreateTodoRequest) (*model.Todo, error) {
	title := strings.TrimSpace(req.Title)
	description := strings.TrimRightFunc(req.Description, unicode.IsSpace) // remove whitespace from end
	if title == "" {
		return nil, ErrEmptyTitle
	}

	now := time.Now().UTC()
	todo := &model.Todo{
		Title:       title,
		Description: description,
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return s.repo.Create(ctx, todo)
}

// GetTodo retrieves a todo by its ID.
func (s *todoService) GetTodo(ctx context.Context, id int64) (*model.Todo, error) {
	return s.repo.GetByID(ctx, id)
}

// GetTodos retrieves all todos.
func (s *todoService) GetTodos(ctx context.Context) ([]model.Todo, error) {
	return s.repo.GetAll(ctx)
}

// UpdateTodo validates and updates fields of an existing todo item.
func (s *todoService) UpdateTodo(ctx context.Context, id int64, req model.UpdateTodoRequest) (*model.Todo, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		trimmed := strings.TrimSpace(*req.Title)
		if trimmed == "" {
			return nil, ErrEmptyTitle
		}
		existing.Title = trimmed
	}

	if req.Description != nil {
		trimmed := strings.TrimRightFunc(*req.Description, unicode.IsSpace) // remove whitespace from end
		existing.Description = trimmed
	}

	if req.Completed != nil {
		existing.Completed = *req.Completed
	}

	existing.UpdatedAt = time.Now().UTC()

	return s.repo.Update(ctx, existing)
}

// DeleteTodo deletes a todo item by its ID.
func (s *todoService) DeleteTodo(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
