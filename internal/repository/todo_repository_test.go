package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"todo/internal/model"
	"todo/internal/repository"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	db.SetMaxOpenConns(1)

	schema := `
	CREATE TABLE todos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		completed BOOLEAN NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func TestRepository_Create_And_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRepository(db)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Second)
	todo := &model.Todo{
		Title:       "Test Todo",
		Description: "A description",
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	created, err := repo.Create(ctx, todo)
	if err != nil {
		t.Fatalf("unexpected error creating todo: %v", err)
	}

	if created.ID <= 0 {
		t.Errorf("expected positive ID, got %d", created.ID)
	}

	retrieved, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("unexpected error retrieving todo: %v", err)
	}

	if retrieved.Title != todo.Title {
		t.Errorf("expected title %q, got %q", todo.Title, retrieved.Title)
	}
	if retrieved.Description != todo.Description {
		t.Errorf("expected description %q, got %q", todo.Description, retrieved.Description)
	}
	if retrieved.Completed != todo.Completed {
		t.Errorf("expected completed %v, got %v", todo.Completed, retrieved.Completed)
	}
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 9999)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRepository(db)
	ctx := context.Background()

	// Initial check: empty slice
	todos, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("unexpected error getting all todos: %v", err)
	}
	if len(todos) != 0 {
		t.Fatalf("expected 0 todos, got %d", len(todos))
	}

	now := time.Now().UTC()
	_, _ = repo.Create(ctx, &model.Todo{Title: "Todo 1", CreatedAt: now, UpdatedAt: now})
	_, _ = repo.Create(ctx, &model.Todo{Title: "Todo 2", CreatedAt: now, UpdatedAt: now})

	todos, err = repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(todos) != 2 {
		t.Fatalf("expected 2 todos, got %d", len(todos))
	}
}

func TestRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRepository(db)
	ctx := context.Background()

	now := time.Now().UTC()
	created, err := repo.Create(ctx, &model.Todo{
		Title:       "Original",
		Description: "Old",
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	created.Title = "Updated Title"
	created.Description = "New description"
	created.Completed = true
	created.UpdatedAt = time.Now().UTC()

	updated, err := repo.Update(ctx, created)
	if err != nil {
		t.Fatalf("unexpected error updating: %v", err)
	}

	if updated.Title != created.Title || updated.Description != created.Description || !updated.Completed {
		t.Errorf("fields not updated properly: %+v", updated)
	}

	// Update nonexistent
	nonExistent := &model.Todo{ID: 8888, Title: "Missing"}
	_, err = repo.Update(ctx, nonExistent)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRepository(db)
	ctx := context.Background()

	now := time.Now().UTC()
	created, err := repo.Create(ctx, &model.Todo{
		Title:     "To Delete",
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("unexpected error deleting: %v", err)
	}

	// Verify it's gone
	_, err = repo.GetByID(ctx, created.ID)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after deletion, got %v", err)
	}

	// Delete again should return ErrNotFound
	if err := repo.Delete(ctx, created.ID); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound deleting nonexistent, got %v", err)
	}
}
