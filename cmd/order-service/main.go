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
	db, err := database.NewDatabase(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName+"_order")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Auto migrate
	if err := db.Migrate(&models.Cart{}, &models.Order{}, &models.OrderItem{}, &models.OrderStatusHistory{}, &models.Payment{}); err != nil {
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
	cartRepo := repository.NewCartRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)

	// Setup services
	cartService := service.NewCartService(cartRepo, redisClient)
	orderService := service.NewOrderService(orderRepo, cartRepo, redisClient, rabbitmqConn, cfg)
	paymentService := service.NewPaymentService(paymentRepo, orderRepo, redisClient, rabbitmqConn, cfg)

	// Setup handlers
	cartHandler := handler.NewCartHandler(cartService)
	orderHandler := handler.NewOrderHandler(orderService)
	paymentHandler := handler.NewPaymentHandler(paymentService)

	// Setup router
	router := gin.Default()
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggerMiddleware())

	// Routes
	api := router.Group("/api/v1")
	{
		// Protected routes (require authentication)
		protected := api.Group("/")
		protected.Use(middleware.JWTAuthMiddleware(cfg.JWTSecret))
		{
			// Cart routes
			cart := protected.Group("/cart")
			{
				cart.GET("", cartHandler.GetCart)
				cart.POST("/items", cartHandler.AddToCart)
				cart.PUT("/items/:id", cartHandler.UpdateCartItem)
				cart.DELETE("/items/:id", cartHandler.RemoveFromCart)
				cart.DELETE("", cartHandler.ClearCart)
			}

			// Order routes
			orders := protected.Group("/orders")
			{
				orders.GET("", orderHandler.GetUserOrders)
				orders.GET("/:id", orderHandler.GetOrderByID)
				orders.POST("", orderHandler.CreateOrder)
				orders.PUT("/:id/cancel", orderHandler.CancelOrder)
				orders.GET("/:id/status", orderHandler.GetOrderStatus)
			}

			// Payment routes
			payments := protected.Group("/payments")
			{
				payments.POST("", paymentHandler.CreatePayment)
				payments.GET("/:id", paymentHandler.GetPaymentByID)
				payments.POST("/:id/callback", paymentHandler.PaymentCallback)
			}

			// Checkout
			protected.POST("/checkout", orderHandler.Checkout)
		}
	}

	// Start server
	log.Printf("Order service starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}