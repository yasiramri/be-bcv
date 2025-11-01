package main

import (
	"log"

	"github.com/be-bcv/ecommerce-backend/internal/handler"
	"github.com/be-bcv/ecommerce-backend/internal/repository"
	"github.com/be-bcv/ecommerce-backend/internal/service"
	"github.com/be-bcv/ecommerce-backend/internal/models"
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
	db, err := database.NewDatabase(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName+"_product")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Auto migrate
	if err := db.Migrate(&models.Category{}, &models.Product{}, &models.ProductReview{}); err != nil {
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
	categoryRepo := repository.NewCategoryRepository(db)
	productRepo := repository.NewProductRepository(db)
	reviewRepo := repository.NewProductReviewRepository(db)

	// Setup services
	categoryService := service.NewCategoryService(categoryRepo)
	productService := service.NewProductService(productRepo, categoryRepo, redisClient, rabbitmqConn)
	reviewService := service.NewProductReviewService(reviewRepo, productRepo)

	// Setup handlers
	categoryHandler := handler.NewCategoryHandler(categoryService)
	productHandler := handler.NewProductHandler(productService)
	reviewHandler := handler.NewProductReviewHandler(reviewService)

	// Setup router
	router := gin.Default()
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggerMiddleware())

	// Routes
	api := router.Group("/api/v1")
	{
		// Public routes
		categories := api.Group("/categories")
		{
			categories.GET("", categoryHandler.GetAllCategories)
			categories.GET("/:id", categoryHandler.GetCategoryByID)
		}

		products := api.Group("/products")
		{
			products.GET("", productHandler.GetAllProducts)
			products.GET("/:id", productHandler.GetProductByID)
			products.GET("/search", productHandler.SearchProducts)
			products.GET("/category/:categoryId", productHandler.GetProductsByCategory)
			products.GET("/:id/reviews", reviewHandler.GetProductReviews)
		}

		// Protected routes (require authentication)
		protected := api.Group("/")
		protected.Use(middleware.JWTAuthMiddleware(cfg.JWTSecret))
		{
			// Product management for sellers
			products := protected.Group("/products")
			{
				products.POST("", productHandler.CreateProduct)
				products.PUT("/:id", productHandler.UpdateProduct)
				products.DELETE("/:id", productHandler.DeleteProduct)
				products.PUT("/:id/stock", productHandler.UpdateStock)

				// Product reviews
				products.POST("/:id/reviews", reviewHandler.CreateReview)
				products.PUT("/reviews/:reviewId", reviewHandler.UpdateReview)
				products.DELETE("/reviews/:reviewId", reviewHandler.DeleteReview)
			}

			// Category management (admin only)
			categories := protected.Group("/categories")
			// TODO: Add admin middleware
			{
				categories.POST("", categoryHandler.CreateCategory)
				categories.PUT("/:id", categoryHandler.UpdateCategory)
				categories.DELETE("/:id", categoryHandler.DeleteCategory)
			}
		}
	}

	// Start server
	log.Printf("Product service starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}