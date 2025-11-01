package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Redis
	RedisHost     string
	RedisPort     string
	RedisPassword string

	// RabbitMQ
	RabbitMQURL string

	// JWT
	JWTSecret    string
	JWTExpiredIn string

	// Midtrans
	MidtransServerKey  string
	MidtransClientKey  string
	MidtransEnvironment string
	MidtransMerchantID string

	// Service URLs
	ProductServiceURL string
	UserServiceURL    string
	OrderServiceURL   string
	PaymentServiceURL string

	// Server Port
	Port string
}

func LoadConfig() *Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	return &Config{
		DBHost:    getEnv("DB_HOST", "localhost"),
		DBPort:    getEnv("DB_PORT", "5432"),
		DBUser:    getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:    getEnv("DB_NAME", "ecommerce_db"),

		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://admin:password@localhost:5672/"),

		JWTSecret:    getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
		JWTExpiredIn: getEnv("JWT_EXPIRED_IN", "24h"),

		MidtransServerKey:   getEnv("MIDTRANS_SERVER_KEY", ""),
		MidtransClientKey:   getEnv("MIDTRANS_CLIENT_KEY", ""),
		MidtransEnvironment: getEnv("MIDTRANS_ENVIRONMENT", "sandbox"),
		MidtransMerchantID:  getEnv("MIDTRANS_MERCHANT_ID", ""),

		ProductServiceURL: getEnv("PRODUCT_SERVICE_URL", "http://localhost:8001"),
		UserServiceURL:    getEnv("USER_SERVICE_URL", "http://localhost:8002"),
		OrderServiceURL:   getEnv("ORDER_SERVICE_URL", "http://localhost:8003"),
		PaymentServiceURL: getEnv("PAYMENT_SERVICE_URL", "http://localhost:8004"),

		Port: getEnv("PORT", "8000"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}