package main

import (
	"log"
	"os"

	"github.com/labstack/echo/v4"

	"todo/internal/database"
	"todo/internal/handler"
	"todo/internal/repository"
	"todo/internal/service"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "todos.db"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 1. Initialize SQLite database
	db, err := database.InitDB(dbPath)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// 2. Instantiate repository layer
	repo := repository.NewRepository(db)

	// 3. Instantiate service layer
	svc := service.NewService(repo)

	// 4. Instantiate handler layer
	todoHandler := handler.NewTodoHandler(svc)

	// 5. Create Echo instance and register routes
	e := echo.New()
	todoHandler.RegisterRoutes(e)

	// 6. Start HTTP server
	log.Printf("Starting server on port %s (db: %s)", port, dbPath)
	if err := e.Start(":" + port); err != nil {
		e.Logger.Fatal(err)
	}
}
