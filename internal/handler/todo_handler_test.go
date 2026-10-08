package handler_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	_ "modernc.org/sqlite"

	"todo/internal/handler"
	"todo/internal/model"
	"todo/internal/repository"
	"todo/internal/service"
)

func setupTestApp(t *testing.T) *echo.Echo {
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
		t.Fatalf("failed to execute schema: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	repo := repository.NewRepository(db)
	svc := service.NewService(repo)
	h := handler.NewTodoHandler(svc)

	e := echo.New()
	h.RegisterRoutes(e)
	return e
}

func TestHandler_CreateTodo(t *testing.T) {
	e := setupTestApp(t)

	// 1. Success
	body := `{"title": "Test Task", "description": "Details"}`
	req := httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d. Body: %s", rec.Code, rec.Body.String())
	}
	var created model.Todo
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if created.Title != "Test Task" || created.ID <= 0 {
		t.Errorf("unexpected created todo: %+v", created)
	}

	// 2. Empty title validation -> 400
	badBody := `{"title": "   "}`
	req = httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(badBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", rec.Code)
	}
	var errResp model.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil || errResp.Error == "" {
		t.Errorf("expected error JSON response, got: %s", rec.Body.String())
	}

	// 3. Malformed JSON -> 400
	req = httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader("{invalid-json"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for malformed JSON, got %d", rec.Code)
	}
}

func TestHandler_GetTodos(t *testing.T) {
	e := setupTestApp(t)

	// Create one todo first
	createReq := httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(`{"title": "Item 1"}`))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)

	// GET /todos
	req := httptest.NewRequest(http.MethodGet, "/todos", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}
	var list []model.Todo
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("failed to decode json: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 item, got %d", len(list))
	}
}

func TestHandler_GetTodo_ByID(t *testing.T) {
	e := setupTestApp(t)

	// Create a todo
	createReq := httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(`{"title": "Item To Get"}`))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)

	var created model.Todo
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)

	// 1. Success
	req := httptest.NewRequest(http.MethodGet, "/todos/1", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}

	// 2. Nonexistent -> 404
	req = httptest.NewRequest(http.MethodGet, "/todos/999", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found, got %d", rec.Code)
	}

	// 3. Invalid ID parameter -> 400
	req = httptest.NewRequest(http.MethodGet, "/todos/abc", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", rec.Code)
	}
}

func TestHandler_UpdateTodo(t *testing.T) {
	e := setupTestApp(t)

	// Create a todo
	createReq := httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(`{"title": "To Update"}`))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)

	// 1. Success -> 200
	updateBody := `{"title": "Updated Name", "completed": true}`
	req := httptest.NewRequest(http.MethodPut, "/todos/1", strings.NewReader(updateBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}
	var updated model.Todo
	_ = json.Unmarshal(rec.Body.Bytes(), &updated)
	if updated.Title != "Updated Name" || !updated.Completed {
		t.Errorf("update did not apply properly: %+v", updated)
	}

	// 2. Not found -> 404
	req = httptest.NewRequest(http.MethodPut, "/todos/999", strings.NewReader(updateBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found, got %d", rec.Code)
	}

	// 3. Invalid ID -> 400
	req = httptest.NewRequest(http.MethodPut, "/todos/invalid", strings.NewReader(updateBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", rec.Code)
	}

	// 4. Invalid input (empty title) -> 400
	req = httptest.NewRequest(http.MethodPut, "/todos/1", strings.NewReader(`{"title": "   "}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", rec.Code)
	}
}

func TestHandler_DeleteTodo(t *testing.T) {
	e := setupTestApp(t)

	// Create a todo
	createReq := httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(`{"title": "To Delete"}`))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)

	// 1. Success -> 204
	req := httptest.NewRequest(http.MethodDelete, "/todos/1", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content, got %d", rec.Code)
	}

	// 2. Already deleted / nonexistent -> 404
	req = httptest.NewRequest(http.MethodDelete, "/todos/1", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found, got %d", rec.Code)
	}

	// 3. Invalid ID -> 400
	req = httptest.NewRequest(http.MethodDelete, "/todos/invalid", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", rec.Code)
	}
}
