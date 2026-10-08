package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"todo/internal/model"
)

var (
	// ErrNotFound is returned when a requested todo item does not exist.
	ErrNotFound = errors.New("todo not found")
)

// Repository defines the data access interface for Todo items.
type Repository interface {
	Create(ctx context.Context, todo *model.Todo) (*model.Todo, error)
	GetByID(ctx context.Context, id int64) (*model.Todo, error)
	GetAll(ctx context.Context) ([]model.Todo, error)
	Update(ctx context.Context, todo *model.Todo) (*model.Todo, error)
	Delete(ctx context.Context, id int64) error
}

type sqlRepository struct {
	db *sql.DB
}

// NewRepository creates a new SQLite repository for managing todos.
func NewRepository(db *sql.DB) Repository {
	return &sqlRepository{db: db}
}

// Create inserts a new todo into the database.
func (r *sqlRepository) Create(ctx context.Context, todo *model.Todo) (*model.Todo, error) {
	const query = `
		INSERT INTO todos (title, description, completed, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		todo.Title,
		todo.Description,
		todo.Completed,
		todo.CreatedAt,
		todo.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert todo: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve last insert id: %w", err)
	}

	todo.ID = id
	return todo, nil
}

// GetByID retrieves a single todo by its ID.
func (r *sqlRepository) GetByID(ctx context.Context, id int64) (*model.Todo, error) {
	const query = `
		SELECT id, title, description, completed, created_at, updated_at
		FROM todos
		WHERE id = ?
	`

	var todo model.Todo
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to query todo by id: %w", err)
	}

	return &todo, nil
}

// GetAll retrieves all todos ordered by ID.
func (r *sqlRepository) GetAll(ctx context.Context) ([]model.Todo, error) {
	const query = `
		SELECT id, title, description, completed, created_at, updated_at
		FROM todos
		ORDER BY id ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query todos: %w", err)
	}
	defer rows.Close()

	todos := make([]model.Todo, 0)
	for rows.Next() {
		var todo model.Todo
		if err := rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Description,
			&todo.Completed,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan todo row: %w", err)
		}
		todos = append(todos, todo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading todo rows: %w", err)
	}

	return todos, nil
}

// Update updates an existing todo in the database.
func (r *sqlRepository) Update(ctx context.Context, todo *model.Todo) (*model.Todo, error) {
	const query = `
		UPDATE todos
		SET title = ?, description = ?, completed = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		todo.Title,
		todo.Description,
		todo.Completed,
		todo.UpdatedAt,
		todo.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update todo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, ErrNotFound
	}

	return todo, nil
}

// Delete removes a todo by its ID.
func (r *sqlRepository) Delete(ctx context.Context, id int64) error {
	const query = `
		DELETE FROM todos
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
