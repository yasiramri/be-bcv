package main

import (
	"log"

	"github.com/be-bcv/ecommerce-backend/internal/handler"
	"github.com/be-bcv/ecommerce-backend/internal/repository"
	"github.com/be-bcv/ecommerce-backend/internal/service"
	"github.com/be-bcv/ecommerce-backend/pkg/config"
	"github.com/be-bcv/ecommerce-backend/pkg/database"
	"github.com/be-bcv/ecommerce-backend/pkg/middleware"
	"github.com/be-bcv/ecommerce-backend/pkg/rabbitmq"
	"github.com/be-bcv/ecommerce-backend/pkg/redis"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	db, err := database.NewDatabase(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName+"_user")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Auto migrate
	if err := db.Migrate(&models.User{}, &models.UserSession{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize Redis
	redisClient, err := redis.NewRedisClient(cfg.RedisHost, cfg.RedisPort, cfg.RedisPassword)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize RabbitMQ
	rabbitmqConn, err := rabbitmq.NewRabbitMQ(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitmqConn.Close()

	// Setup repositories
	userRepo := repository.NewUserRepository(db)

	// Setup services
	userService := service.NewUserService(userRepo, redisClient, rabbitmqConn, cfg)

	// Setup handlers
	userHandler := handler.NewUserHandler(userService)

	// Setup router
	router := gin.Default()
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggerMiddleware())

	// Routes
	api := router.Group("/api/v1")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
			auth.POST("/logout", userHandler.Logout)
			auth.POST("/refresh", userHandler.RefreshToken)
		}

		// User routes (protected)
		users := api.Group("/users")
		users.Use(middleware.JWTAuthMiddleware(cfg.JWTSecret))
		{
			users.GET("/profile", userHandler.GetProfile)
			users.PUT("/profile", userHandler.UpdateProfile)
			users.DELETE("/account", userHandler.DeleteAccount)
		}

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(middleware.JWTAuthMiddleware(cfg.JWTSecret))
		// TODO: Add admin role middleware
		{
			admin.GET("/users", userHandler.GetAllUsers)
			admin.GET("/users/:id", userHandler.GetUserByID)
			admin.PUT("/users/:id/status", userHandler.UpdateUserStatus)
		}
	}

	// Start server
	log.Printf("User service starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}