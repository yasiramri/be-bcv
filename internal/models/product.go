package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Category struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type Product struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string     `gorm:"not null" json:"name"`
	Description string     `json:"description"`
	Price       float64    `gorm:"not null" json:"price"`
	Stock       int        `gorm:"not null;default:0" json:"stock"`
	SKU         string     `gorm:"uniqueIndex;not null" json:"sku"`
	Images      []string   `gorm:"type:text[]" json:"images"`
	CategoryID  uuid.UUID  `gorm:"type:uuid;not null" json:"category_id"`
	SellerID    uuid.UUID  `gorm:"type:uuid;not null" json:"seller_id"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	Weight      float64    `json:"weight"` // in kg
	Dimensions  string     `json:"dimensions"` // format: "lengthxwidthxheight"
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Category Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

type ProductReview struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Rating    int       `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Product Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (Category) TableName() string {
	return "categories"
}

func (Product) TableName() string {
	return "products"
}

func (ProductReview) TableName() string {
	return "product_reviews"
}