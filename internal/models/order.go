package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Cart struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	Quantity  int       `gorm:"not null" json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Product Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

type Order struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	OrderNumber  string     `gorm:"uniqueIndex;not null" json:"order_number"`
	Status       string     `gorm:"default:pending" json:"status"` // pending, confirmed, shipped, delivered, cancelled
	TotalAmount  float64    `gorm:"not null" json:"total_amount"`
	ShippingCost float64    `gorm:"default:0" json:"shipping_cost"`
	Subtotal     float64    `gorm:"not null" json:"subtotal"`
	Address      string     `gorm:"not null" json:"address"`
	City         string     `json:"city"`
	Province     string     `json:"province"`
	PostalCode   string     `json:"postal_code"`
	PaymentID    uuid.UUID  `gorm:"type:uuid" json:"payment_id"`
	PaymentStatus string    `gorm:"default:pending" json:"payment_status"` // pending, paid, failed
	ShippingDate *time.Time `json:"shipping_date"`
	DeliveryDate *time.Time `json:"delivery_date"`
	Notes        string     `json:"notes"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	User     uuid.UUID `gorm:"-" json:"user,omitempty"`
	Payment  *Payment `gorm:"foreignKey:PaymentID" json:"payment,omitempty"`
	Items    []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

type OrderItem struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrderID   uuid.UUID `gorm:"type:uuid;not null" json:"order_id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	Quantity  int       `gorm:"not null" json:"quantity"`
	Price     float64   `gorm:"not null" json:"price"`
	Subtotal  float64   `gorm:"not null" json:"subtotal"`
	CreatedAt time.Time `json:"created_at"`

	Product Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

type OrderStatusHistory struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrderID     uuid.UUID `gorm:"type:uuid;not null" json:"order_id"`
	FromStatus  string    `json:"from_status"`
	ToStatus    string    `gorm:"not null" json:"to_status"`
	Notes       string    `json:"notes"`
	CreatedBy   uuid.UUID `gorm:"type:uuid" json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type Payment struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrderID       uuid.UUID  `gorm:"type:uuid;not null" json:"order_id"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	Amount        float64    `gorm:"not null" json:"amount"`
	Method        string     `gorm:"not null" json:"method"` // credit_card, bank_transfer, e_wallet, etc.
	Status        string     `gorm:"default:pending" json:"status"` // pending, paid, failed, cancelled, refunded
	MidtransID    string     `json:"midtrans_id"`
	MidtransVA    string     `json:"midtrans_va"`
	PaymentURL    string     `json:"payment_url"`
	ExpiredAt     time.Time  `json:"expired_at"`
	PaidAt        *time.Time `json:"paid_at"`
	TransactionID string     `json:"transaction_id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (Cart) TableName() string {
	return "carts"
}

func (Order) TableName() string {
	return "orders"
}

func (OrderItem) TableName() string {
	return "order_items"
}

func (OrderStatusHistory) TableName() string {
	return "order_status_histories"
}

func (Payment) TableName() string {
	return "payments"
}