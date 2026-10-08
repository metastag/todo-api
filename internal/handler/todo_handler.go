package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"todo/internal/model"
	"todo/internal/repository"
	"todo/internal/service"
)

// TodoHandler handles HTTP requests for todo operations.
type TodoHandler struct {
	svc service.Service
}

// NewTodoHandler creates a new TodoHandler instance.
func NewTodoHandler(svc service.Service) *TodoHandler {
	return &TodoHandler{svc: svc}
}

// RegisterRoutes registers the todo REST endpoints on an Echo router or group.
func (h *TodoHandler) RegisterRoutes(e *echo.Echo) {
	e.POST("/todos", h.CreateTodo)
	e.GET("/todos", h.GetTodos)
	e.GET("/todos/:id", h.GetTodo)
	e.PUT("/todos/:id", h.UpdateTodo)
	e.DELETE("/todos/:id", h.DeleteTodo)
}

// CreateTodo handles POST /todos
func (h *TodoHandler) CreateTodo(c echo.Context) error {
	var req model.CreateTodoRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "malformed request payload",
		})
	}

	created, err := h.svc.CreateTodo(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrEmptyTitle) {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error: err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: "internal server error",
		})
	}

	return c.JSON(http.StatusCreated, created)
}

// GetTodos handles GET /todos
func (h *TodoHandler) GetTodos(c echo.Context) error {
	todos, err := h.svc.GetTodos(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: "internal server error",
		})
	}

	return c.JSON(http.StatusOK, todos)
}

// GetTodo handles GET /todos/:id
func (h *TodoHandler) GetTodo(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "invalid id parameter",
		})
	}

	todo, err := h.svc.GetTodo(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error: "todo not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: "internal server error",
		})
	}

	return c.JSON(http.StatusOK, todo)
}

// UpdateTodo handles PUT /todos/:id
func (h *TodoHandler) UpdateTodo(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "invalid id parameter",
		})
	}

	var req model.UpdateTodoRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "malformed request payload",
		})
	}

	updated, err := h.svc.UpdateTodo(c.Request().Context(), id, req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error: "todo not found",
			})
		}
		if errors.Is(err, service.ErrEmptyTitle) {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error: err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: "internal server error",
		})
	}

	return c.JSON(http.StatusOK, updated)
}

// DeleteTodo handles DELETE /todos/:id
func (h *TodoHandler) DeleteTodo(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "invalid id parameter",
		})
	}

	if err := h.svc.DeleteTodo(c.Request().Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error: "todo not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: "internal server error",
		})
	}

	return c.NoContent(http.StatusNoContent)
}
