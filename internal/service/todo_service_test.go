package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "modernc.org/sqlite"

	"todo/internal/model"
	"todo/internal/repository"
	"todo/internal/service"
)

func setupTestService(t *testing.T) service.Service {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
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

	repo := repository.NewRepository(db)
	return service.NewService(repo)
}

func TestService_CreateTodo(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	// Empty title validation
	_, err := svc.CreateTodo(ctx, model.CreateTodoRequest{
		Title: "   ",
	})
	if !errors.Is(err, service.ErrEmptyTitle) {
		t.Fatalf("expected ErrEmptyTitle, got %v", err)
	}

	// Valid creation with trailing whitespace in description
	created, err := svc.CreateTodo(ctx, model.CreateTodoRequest{
		Title:       "  Buy Milk  ",
		Description: "From local store   \t\n",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if created.Title != "Buy Milk" {
		t.Errorf("expected trimmed title %q, got %q", "Buy Milk", created.Title)
	}
	if created.Description != "From local store" {
		t.Errorf("expected trimmed description %q, got %q", "From local store", created.Description)
	}
	if created.Completed {
		t.Errorf("expected completed to be false initially")
	}
	if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Errorf("expected timestamps to be set")
	}
}

func TestService_GetTodo_And_GetTodos(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	created, err := svc.CreateTodo(ctx, model.CreateTodoRequest{Title: "Task 1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	retrieved, err := svc.GetTodo(ctx, created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, retrieved.ID)
	}

	// Nonexistent
	_, err = svc.GetTodo(ctx, 999)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	// Get all
	all, err := svc.GetTodos(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 item, got %d", len(all))
	}
}

func TestService_UpdateTodo(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	created, err := svc.CreateTodo(ctx, model.CreateTodoRequest{
		Title:       "Initial",
		Description: "Initial desc",
	})
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	// Update with empty title error
	emptyTitle := "  "
	_, err = svc.UpdateTodo(ctx, created.ID, model.UpdateTodoRequest{
		Title: &emptyTitle,
	})
	if !errors.Is(err, service.ErrEmptyTitle) {
		t.Fatalf("expected ErrEmptyTitle, got %v", err)
	}

	// Valid partial update
	newTitle := "Updated Title"
	completed := true
	updated, err := svc.UpdateTodo(ctx, created.ID, model.UpdateTodoRequest{
		Title:     &newTitle,
		Completed: &completed,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Title != "Updated Title" {
		t.Errorf("expected title to be updated")
	}
	if !updated.Completed {
		t.Errorf("expected completed to be true")
	}
	if updated.Description != "Initial desc" {
		t.Errorf("expected description to remain unchanged")
	}

	// Update nonexistent
	_, err = svc.UpdateTodo(ctx, 999, model.UpdateTodoRequest{
		Title: &newTitle,
	})
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestService_DeleteTodo(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	created, err := svc.CreateTodo(ctx, model.CreateTodoRequest{Title: "Delete Me"})
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	if err := svc.DeleteTodo(ctx, created.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify not found
	_, err = svc.GetTodo(ctx, created.ID)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	// Delete nonexistent
	if err := svc.DeleteTodo(ctx, 999); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
