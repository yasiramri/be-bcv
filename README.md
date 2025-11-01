# BCV E-Commerce Backend

Modular Go backend for the BCV marketplace, composed of independent microservices for products, users, orders, and payments. Each service exposes REST APIs, shares infrastructure via PostgreSQL, Redis, and RabbitMQ, and is orchestrated through an API gateway.

## Features

- Segmented microservices with dedicated data stores and business logic
- JWT-based authentication and session handling backed by Redis
- Product catalogue with categories, search, reviews, and seller tooling
- Shopping cart, order lifecycle, and payment processing abstractions
- Asynchronous event hooks prepared via RabbitMQ for cross-service workflows
- Containerized deployment with Docker and shared infrastructure defined in `docker-compose.yml`

## Architecture Overview

- `api-gateway`: Routes external requests to downstream services and centralizes authentication
- `product-service`: Manages categories, products, stock levels, and user-generated reviews
- `user-service`: Handles registration, login, profile management, and admin operations
- `order-service`: Owns carts, orders, order items, and payment state transitions
- `payment-service`: Integrates with Midtrans and manages payment records (implementation placeholder)

Shared packages live under `pkg/` (configuration, database, redis, rabbitmq, middleware, utilities), while domain-specific logic sits under `internal/` (handlers, services, repositories, and models).

Primary infrastructure dependencies:

- PostgreSQL for persistent storage (one logical database per service)
- Redis for caching and token/session management
- RabbitMQ for asynchronous messaging and eventual consistency

## Tech Stack

- Go 1.21 with Gin, GORM, go-redis, and JWT libraries
- PostgreSQL 15
- Redis 7
- RabbitMQ 3 (with management UI)
- Midtrans Go SDK for payment gateway integration
- Docker & Docker Compose for local orchestration

## Getting Started

### Prerequisites

- Go 1.21+
- Docker & Docker Compose (recommended for full stack)
- Make sure the `scripts/init-db.sql` script is compatible with your Postgres image (uses `IF NOT EXISTS` syntax)

### Clone the Repository

```
git clone https://github.com/<your-org>/be-bcv.git
cd be-bcv
```

### Environment Variables

Create a `.env` file at the project root. Default values exist in `pkg/config/config.go`, but a full example is below:

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=ecommerce_db

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

RABBITMQ_URL=amqp://admin:password@localhost:5672/

JWT_SECRET=your-super-secret-jwt-key
JWT_EXPIRED_IN=24h

MIDTRANS_SERVER_KEY=your-midtrans-server-key
MIDTRANS_CLIENT_KEY=your-midtrans-client-key
MIDTRANS_ENVIRONMENT=sandbox
MIDTRANS_MERCHANT_ID=

PRODUCT_SERVICE_URL=http://localhost:8001
USER_SERVICE_URL=http://localhost:8002
ORDER_SERVICE_URL=http://localhost:8003
PAYMENT_SERVICE_URL=http://localhost:8004

PORT=8000
```

### Run with Docker Compose (Recommended)

```
docker-compose up --build
```

Exposed ports:

- API Gateway: 8000
- Product Service: 8001
- User Service: 8002
- Order Service: 8003
- Payment Service: 8004
- PostgreSQL: 5432
- Redis: 6379
- RabbitMQ AMQP: 5672 (Management UI on 15672)

The compose file builds each service from source and provisions the infrastructure containers. Database init scripts under `scripts/` run automatically on first boot.

### Run Services Locally (without Docker)

1. Ensure PostgreSQL, Redis, and RabbitMQ are running and your `.env` file is populated.
2. Start individual services:
   - `go run ./cmd/product-service`
   - `go run ./cmd/user-service`
   - `go run ./cmd/order-service`
   - `go run ./cmd/payment-service`
   - `go run ./cmd/api-gateway`
3. Each service reads configuration from environment variables at startup via `pkg/config`.

## Key Directories

- `cmd/`: Entry points for each microservice and the API gateway
- `internal/`: Core business logic (handlers, services, repositories, models)
- `pkg/`: Shared utilities and infrastructure helpers
- `scripts/`: Database initialization scripts and operational tooling
- `docs/`: Supplemental documentation (if any)

## Messaging & Events

Event payload definitions reside in `pkg/messages`. Publishing hooks are scaffolded in service layer code (`internal/service`) and can be wired to RabbitMQ exchanges to broadcast domain events such as `product.created`, `user.registered`, or `order.created`.

## Testing

Unit and integration tests are not yet implemented. Recommended next steps:

- Introduce service-layer tests using dependency injection/mocks
- Add contract tests for API handlers
- Incorporate end-to-end tests hitting the API gateway

## Contributing

1. Fork the repository and create a feature branch.
2. Make your changes and ensure services build: `go build ./...`
3. Submit a pull request with clear description and testing notes.

## License

Project licensing is not specified. Add a `LICENSE` file if you intend to open-source or share this code.
