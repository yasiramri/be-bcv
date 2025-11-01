package repository

import (
	"errors"

	"github.com/be-bcv/ecommerce-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CartRepository struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) *CartRepository {
	return &CartRepository{db: db}
}

func (r *CartRepository) AddToCart(cart *models.Cart) error {
	// Check if item already exists in cart
	var existingCart models.Cart
	err := r.db.Where("user_id = ? AND product_id = ?", cart.UserID, cart.ProductID).First(&existingCart).Error
	if err == nil {
		// Update quantity if item exists
		existingCart.Quantity += cart.Quantity
		return r.db.Save(&existingCart).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Add new item to cart
	return r.db.Create(cart).Error
}

func (r *CartRepository) GetCart(userID uuid.UUID) ([]models.Cart, error) {
	var cartItems []models.Cart
	err := r.db.Preload("Product").Where("user_id = ?", userID).Find(&cartItems).Error
	return cartItems, err
}

func (r *CartRepository) UpdateCartItem(cartID uuid.UUID, userID uuid.UUID, quantity int) error {
	return r.db.Model(&models.Cart{}).
		Where("id = ? AND user_id = ?", cartID, userID).
		Update("quantity", quantity).Error
}

func (r *CartRepository) RemoveFromCart(cartID uuid.UUID, userID uuid.UUID) error {
	return r.db.Where("id = ? AND user_id = ?", cartID, userID).Delete(&models.Cart{}).Error
}

func (r *CartRepository) ClearCart(userID uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.Cart{}).Error
}

func (r *CartRepository) GetCartItemByID(cartID uuid.UUID) (*models.Cart, error) {
	var cart models.Cart
	err := r.db.Preload("Product").Where("id = ?", cartID).First(&cart).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cart, nil
}

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) CreateOrder(order *models.Order) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create order
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		// Create order status history
		history := &models.OrderStatusHistory{
			ID:        uuid.New(),
			OrderID:   order.ID,
			ToStatus:  order.Status,
			Notes:     "Order created",
			CreatedBy: order.UserID,
		}
		return tx.Create(history).Error
	})
}

func (r *OrderRepository) CreateOrderItem(item *models.OrderItem) error {
	return r.db.Create(item).Error
}

func (r *OrderRepository) GetOrderByID(orderID uuid.UUID, userID uuid.UUID) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("Items.Product").
		Preload("Payment").
		Where("id = ? AND user_id = ?", orderID, userID).
		First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) GetOrderByIDForAdmin(orderID uuid.UUID) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("Items.Product").
		Preload("Payment").
		Where("id = ?", orderID).
		First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) GetUserOrders(userID uuid.UUID, page, limit int) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := r.db.Model(&models.Order{}).
		Preload("Items.Product").
		Preload("Payment").
		Where("user_id = ?", userID)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at desc").Find(&orders).Error

	return orders, total, err
}

func (r *OrderRepository) GetAllOrders(page, limit int, status string) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := r.db.Model(&models.Order{}).
		Preload("Items.Product").
		Preload("Payment")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at desc").Find(&orders).Error

	return orders, total, err
}

func (r *OrderRepository) UpdateOrderStatus(orderID uuid.UUID, status string, notes string, updatedBy uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Get current order
		var order models.Order
		if err := tx.Where("id = ?", orderID).First(&order).Error; err != nil {
			return err
		}

		// Update order status
		if err := tx.Model(&order).Update("status", status).Error; err != nil {
			return err
		}

		// Create status history
		history := &models.OrderStatusHistory{
			ID:        uuid.New(),
			OrderID:   orderID,
			FromStatus: order.Status,
			ToStatus:  status,
			Notes:     notes,
			CreatedBy: updatedBy,
		}
		return tx.Create(history).Error
	})
}

func (r *OrderRepository) UpdatePaymentStatus(orderID uuid.UUID, paymentID uuid.UUID, paymentStatus string) error {
	return r.db.Model(&models.Order{}).
		Where("id = ?", orderID).
		Updates(map[string]interface{}{
			"payment_id":     paymentID,
			"payment_status": paymentStatus,
		}).Error
}

func (r *OrderRepository) GetOrderStatusHistories(orderID uuid.UUID) ([]models.OrderStatusHistory, error) {
	var histories []models.OrderStatusHistory
	err := r.db.Where("order_id = ?", orderID).Order("created_at desc").Find(&histories).Error
	return histories, err
}

func (r *OrderRepository) UpdateOrder(order *models.Order) error {
	return r.db.Save(order).Error
}

type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) CreatePayment(payment *models.Payment) error {
	return r.db.Create(payment).Error
}

func (r *PaymentRepository) GetPaymentByID(paymentID uuid.UUID) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Where("id = ?", paymentID).First(&payment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

func (r *PaymentRepository) GetPaymentByOrderID(orderID uuid.UUID) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Where("order_id = ?", orderID).First(&payment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

func (r *PaymentRepository) UpdatePaymentStatus(paymentID uuid.UUID, status string, transactionID string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if transactionID != "" {
		updates["transaction_id"] = transactionID
	}

	if status == "paid" {
		now := time.Now()
		updates["paid_at"] = &now
	}

	return r.db.Model(&models.Payment{}).
		Where("id = ?", paymentID).
		Updates(updates).Error
}

func (r *PaymentRepository) UpdatePayment(payment *models.Payment) error {
	return r.db.Save(payment).Error
}

func (r *PaymentRepository) GetPaymentsByUserID(userID uuid.UUID, page, limit int) ([]models.Payment, int64, error) {
	var payments []models.Payment
	var total int64

	query := r.db.Model(&models.Payment{}).Where("user_id = ?", userID)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at desc").Find(&payments).Error

	return payments, total, err
}