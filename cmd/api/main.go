package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/irzan-aldi/task-manager/internal/config"
	"github.com/irzan-aldi/task-manager/internal/handler"
	"github.com/irzan-aldi/task-manager/internal/model"
	"github.com/irzan-aldi/task-manager/internal/repository"
	"github.com/irzan-aldi/task-manager/internal/router"
	"github.com/irzan-aldi/task-manager/internal/service"
)

func main() {
	// Load config
	cfg := config.Load()

	// Connect database
	db := config.ConnectDB(cfg)

	// Auto-migrate models
	if err := db.AutoMigrate(&model.User{}, &model.Board{}, &model.Task{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migrated successfully")

	// Connect Redis
	redisClient := config.ConnectRedis(cfg)

	// Setup repositories
	userRepo := repository.NewUserRepository(db)
	boardRepo := repository.NewBoardRepository(db)
	taskRepo := repository.NewTaskRepository(db)

	// Setup services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	userService := service.NewUserService(userRepo)
	boardService := service.NewBoardService(boardRepo)
	taskService := service.NewTaskService(taskRepo, boardRepo)

	// Setup handlers
	handlers := &router.Handlers{
		Auth:  handler.NewAuthHandler(authService),
		User:  handler.NewUserHandler(userService),
		Board: handler.NewBoardHandler(boardService),
		Task:  handler.NewTaskHandler(taskService),
	}

	// Setup Gin
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	router.Setup(r, handlers, authService, redisClient)

	// Start server
	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s (env: %s)", port, cfg.Env)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
		os.Exit(1)
	}
}
