package messages

import "time"

type EventMessage struct {
	EventID   string      `json:"event_id"`
	EventName string      `json:"event_name"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
	Service   string      `json:"service"`
}

// User Events
type UserRegisteredEvent struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

type UserUpdatedEvent struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

// Product Events
type ProductCreatedEvent struct {
	ProductID   string  `json:"product_id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	CategoryID  string  `json:"category_id"`
	SellerID    string  `json:"seller_id"`
}

type ProductUpdatedEvent struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Stock     int     `json:"stock"`
}

type ProductDeletedEvent struct {
	ProductID string `json:"product_id"`
}

type StockUpdatedEvent struct {
	ProductID string `json:"product_id"`
	OldStock  int    `json:"old_stock"`
	NewStock  int    `json:"new_stock"`
}

// Order Events
type OrderCreatedEvent struct {
	OrderID   string    `json:"order_id"`
	UserID    string    `json:"user_id"`
	Total     float64   `json:"total"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type OrderUpdatedEvent struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

type OrderCancelledEvent struct {
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}

// Payment Events
type PaymentCreatedEvent struct {
	PaymentID    string  `json:"payment_id"`
	OrderID      string  `json:"order_id"`
	UserID       string  `json:"user_id"`
	Amount       float64 `json:"amount"`
	Method       string  `json:"method"`
	Status       string  `json:"status"`
	MidtransID   string  `json:"midtrans_id"`
}

type PaymentSuccessEvent struct {
	PaymentID string `json:"payment_id"`
	OrderID   string `json:"order_id"`
	Amount    float64 `json:"amount"`
}

type PaymentFailedEvent struct {
	PaymentID string `json:"payment_id"`
	OrderID   string `json:"order_id"`
	Amount    float64 `json:"amount"`
	Reason    string  `json:"reason"`
}