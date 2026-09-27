package router

import (
	"github.com/gin-gonic/gin"
	"github.com/irzan-aldi/task-manager/internal/handler"
	"github.com/irzan-aldi/task-manager/internal/middleware"
	"github.com/irzan-aldi/task-manager/internal/model"
	"github.com/irzan-aldi/task-manager/internal/service"
	"github.com/redis/go-redis/v9"
	"time"
)

type Handlers struct {
	Auth  *handler.AuthHandler
	User  *handler.UserHandler
	Board *handler.BoardHandler
	Task  *handler.TaskHandler
}

func Setup(r *gin.Engine, h *Handlers, authService *service.AuthService, redisClient *redis.Client) {
	// Health check (public)
	r.GET("/health", handler.HealthCheck)

	// Rate limiter: 100 requests per minute per IP
	r.Use(middleware.RateLimiter(redisClient, 100, time.Minute))

	v1 := r.Group("/api/v1")
	{
		// Public routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", h.Auth.Register)
			auth.POST("/login", h.Auth.Login)
			auth.POST("/refresh", h.Auth.RefreshToken)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			// Auth
			protected.GET("/auth/me", h.Auth.Me)

			// Users
			protected.GET("/users/profile", h.User.GetProfile)
			protected.PUT("/users/profile", h.User.Update)

			// Users — admin only
			admin := protected.Group("")
			admin.Use(middleware.RoleMiddleware(model.RoleAdmin))
			{
				admin.GET("/users", h.User.GetAll)
				admin.DELETE("/users/:id", h.User.Delete)
			}

			// Boards
			protected.POST("/boards", h.Board.Create)
			protected.GET("/boards", h.Board.GetMyBoards)
			protected.GET("/boards/:id", h.Board.GetByID)
			protected.PUT("/boards/:id", h.Board.Update)
			protected.DELETE("/boards/:id", h.Board.Delete)

			// Tasks
			protected.POST("/boards/:boardId/tasks", h.Task.Create)
			protected.GET("/boards/:boardId/tasks", h.Task.GetByBoardID)
			protected.GET("/tasks/my", h.Task.GetMyTasks)
			protected.GET("/tasks/:id", h.Task.GetByID)
			protected.PUT("/tasks/:id", h.Task.Update)
			protected.DELETE("/tasks/:id", h.Task.Delete)
		}
	}
}
