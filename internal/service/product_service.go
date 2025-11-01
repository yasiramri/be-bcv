package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/be-bcv/ecommerce-backend/internal/models"
	"github.com/be-bcv/ecommerce-backend/internal/repository"
	"github.com/be-bcv/ecommerce-backend/pkg/messages"
	"github.com/be-bcv/ecommerce-backend/pkg/rabbitmq"
	"github.com/be-bcv/ecommerce-backend/pkg/redis"
	"github.com/google/uuid"
)

type ProductService struct {
	productRepo  *repository.ProductRepository
	categoryRepo *repository.CategoryRepository
	redis        *redis.RedisClient
	rabbitmq     *rabbitmq.RabbitMQ
}

func NewProductService(productRepo *repository.ProductRepository, categoryRepo *repository.CategoryRepository, redis *redis.RedisClient, rabbitmq *rabbitmq.RabbitMQ) *ProductService {
	return &ProductService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		redis:        redis,
		rabbitmq:     rabbitmq,
	}
}

type CreateProductRequest struct {
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	Price       float64   `json:"price" binding:"required,min=0"`
	Stock       int       `json:"stock" binding:"required,min=0"`
	CategoryID  uuid.UUID `json:"category_id" binding:"required"`
	SellerID    uuid.UUID `json:"seller_id" binding:"required"`
	Weight      float64   `json:"weight"`
	Dimensions  string    `json:"dimensions"`
	Images      []string  `json:"images"`
}

type UpdateProductRequest struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	CategoryID  uuid.UUID `json:"category_id"`
	Weight      float64   `json:"weight"`
	Dimensions  string    `json:"dimensions"`
	Images      []string  `json:"images"`
}

type UpdateStockRequest struct {
	Stock int `json:"stock" binding:"required,min=0"`
}

type ProductResponse struct {
	*models.Product
	AverageRating float64 `json:"average_rating"`
	ReviewCount   int64   `json:"review_count"`
}

func (s *ProductService) CreateProduct(req *CreateProductRequest) (*models.Product, error) {
	// Check if category exists
	category, err := s.categoryRepo.GetByID(req.CategoryID)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, fmt.Errorf("category not found")
	}

	// Generate SKU
	sku := s.generateSKU(req.Name)

	// Create product
	product := &models.Product{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		SKU:         sku,
		CategoryID:  req.CategoryID,
		SellerID:    req.SellerID,
		Weight:      req.Weight,
		Dimensions:  req.Dimensions,
		Images:      req.Images,
		IsActive:    true,
	}

	if err := s.productRepo.Create(product); err != nil {
		return nil, err
	}

	// Cache product
	s.cacheProduct(product)

	// Publish product created event
	s.publishProductCreatedEvent(product)

	return product, nil
}

func (s *ProductService) GetProductByID(id uuid.UUID) (*ProductResponse, error) {
	// Try to get from cache first
	cachedProduct, err := s.getCachedProduct(id)
	if err == nil && cachedProduct != nil {
		return s.buildProductResponse(cachedProduct)
	}

	// Get from database
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, fmt.Errorf("product not found")
	}

	// Cache product
	s.cacheProduct(product)

	return s.buildProductResponse(product)
}

func (s *ProductService) GetAllProducts(page, limit int, categoryID uuid.UUID, sortBy, sortOrder string) ([]ProductResponse, int64, error) {
	products, total, err := s.productRepo.GetAll(page, limit, categoryID, sortBy, sortOrder)
	if err != nil {
		return nil, 0, err
	}

	var responses []ProductResponse
	for _, product := range products {
		response, err := s.buildProductResponse(&product)
		if err != nil {
			continue
		}
		responses = append(responses, *response)
	}

	return responses, total, nil
}

func (s *ProductService) SearchProducts(query string, page, limit int) ([]ProductResponse, int64, error) {
	products, total, err := s.productRepo.Search(query, page, limit)
	if err != nil {
		return nil, 0, err
	}

	var responses []ProductResponse
	for _, product := range products {
		response, err := s.buildProductResponse(&product)
		if err != nil {
			continue
		}
		responses = append(responses, *response)
	}

	return responses, total, nil
}

func (s *ProductService) GetProductsByCategory(categoryID uuid.UUID, page, limit int) ([]ProductResponse, int64, error) {
	products, total, err := s.productRepo.GetByCategory(categoryID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	var responses []ProductResponse
	for _, product := range products {
		response, err := s.buildProductResponse(&product)
		if err != nil {
			continue
		}
		responses = append(responses, *response)
	}

	return responses, total, nil
}

func (s *ProductService) GetProductsBySeller(sellerID uuid.UUID, page, limit int) ([]ProductResponse, int64, error) {
	products, total, err := s.productRepo.GetBySeller(sellerID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	var responses []ProductResponse
	for _, product := range products {
		response, err := s.buildProductResponse(&product)
		if err != nil {
			continue
		}
		responses = append(responses, *response)
	}

	return responses, total, nil
}

func (s *ProductService) UpdateProduct(id uuid.UUID, req *UpdateProductRequest) (*models.Product, error) {
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, fmt.Errorf("product not found")
	}

	// Update fields
	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.Price > 0 {
		product.Price = req.Price
	}
	if req.Stock >= 0 {
		product.Stock = req.Stock
	}
	if req.CategoryID != uuid.Nil {
		// Check if category exists
		category, err := s.categoryRepo.GetByID(req.CategoryID)
		if err != nil {
			return nil, err
		}
		if category == nil {
			return nil, fmt.Errorf("category not found")
		}
		product.CategoryID = req.CategoryID
	}
	if req.Weight > 0 {
		product.Weight = req.Weight
	}
	if req.Dimensions != "" {
		product.Dimensions = req.Dimensions
	}
	if req.Images != nil {
		product.Images = req.Images
	}

	if err := s.productRepo.Update(product); err != nil {
		return nil, err
	}

	// Update cache
	s.cacheProduct(product)

	// Publish product updated event
	s.publishProductUpdatedEvent(product)

	return product, nil
}

func (s *ProductService) UpdateStock(id uuid.UUID, req *UpdateStockRequest) error {
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return err
	}
	if product == nil {
		return fmt.Errorf("product not found")
	}

	oldStock := product.Stock

	if err := s.productRepo.UpdateStock(id, req.Stock); err != nil {
		return err
	}

	// Update product in cache
	product.Stock = req.Stock
	s.cacheProduct(product)

	// Publish stock updated event
	s.publishStockUpdatedEvent(id, oldStock, req.Stock)

	return nil
}

func (s *ProductService) DeleteProduct(id uuid.UUID) error {
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return err
	}
	if product == nil {
		return fmt.Errorf("product not found")
	}

	if err := s.productRepo.Delete(id); err != nil {
		return err
	}

	// Remove from cache
	ctx := context.Background()
	key := fmt.Sprintf("product:%s", id.String())
	s.redis.Del(ctx, key)

	// Publish product deleted event
	s.publishProductDeletedEvent(id)

	return nil
}

func (s *ProductService) buildProductResponse(product *models.Product) (*ProductResponse, error) {
	// TODO: Get average rating and review count from review service
	return &ProductResponse{
		Product:      product,
		AverageRating: 0,
		ReviewCount:  0,
	}, nil
}

func (s *ProductService) generateSKU(name string) string {
	// Simple SKU generation - in production, you might want a more sophisticated approach
	timestamp := time.Now().Unix()
	return fmt.Sprintf("PRD-%d", timestamp)
}

func (s *ProductService) cacheProduct(product *models.Product) {
	ctx := context.Background()
	key := fmt.Sprintf("product:%s", product.ID.String())
	// Cache for 1 hour
	s.redis.Set(ctx, key, product, time.Hour)
}

func (s *ProductService) getCachedProduct(id uuid.UUID) (*models.Product, error) {
	ctx := context.Background()
	key := fmt.Sprintf("product:%s", id.String())

	// This is a simplified version - in production, you'd want proper deserialization
	_, err := s.redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	// For now, return nil to always fetch from DB
	// TODO: Implement proper caching with serialization
	return nil, nil
}

func (s *ProductService) publishProductCreatedEvent(product *models.Product) {
	event := messages.EventMessage{
		EventID:   uuid.New().String(),
		EventName: "product.created",
		Timestamp: time.Now(),
		Data: messages.ProductCreatedEvent{
			ProductID:  product.ID.String(),
			Name:       product.Name,
			Price:      product.Price,
			Stock:      product.Stock,
			CategoryID: product.CategoryID.String(),
			SellerID:   product.SellerID.String(),
		},
		Service: "product-service",
	}

	// TODO: Publish to RabbitMQ
	// s.rabbitmq.Publish("product_events", "product.created", event)
}

func (s *ProductService) publishProductUpdatedEvent(product *models.Product) {
	event := messages.EventMessage{
		EventID:   uuid.New().String(),
		EventName: "product.updated",
		Timestamp: time.Now(),
		Data: messages.ProductUpdatedEvent{
			ProductID: product.ID.String(),
			Name:      product.Name,
			Price:     product.Price,
			Stock:     product.Stock,
		},
		Service: "product-service",
	}

	// TODO: Publish to RabbitMQ
	// s.rabbitmq.Publish("product_events", "product.updated", event)
}

func (s *ProductService) publishStockUpdatedEvent(productID uuid.UUID, oldStock, newStock int) {
	event := messages.EventMessage{
		EventID:   uuid.New().String(),
		EventName: "product.stock_updated",
		Timestamp: time.Now(),
		Data: messages.StockUpdatedEvent{
			ProductID: productID.String(),
			OldStock:  oldStock,
			NewStock:  newStock,
		},
		Service: "product-service",
	}

	// TODO: Publish to RabbitMQ
	// s.rabbitmq.Publish("product_events", "product.stock_updated", event)
}

func (s *ProductService) publishProductDeletedEvent(productID uuid.UUID) {
	event := messages.EventMessage{
		EventID:   uuid.New().String(),
		EventName: "product.deleted",
		Timestamp: time.Now(),
		Data: messages.ProductDeletedEvent{
			ProductID: productID.String(),
		},
		Service: "product-service",
	}

	// TODO: Publish to RabbitMQ
	// s.rabbitmq.Publish("product_events", "product.deleted", event)
}

// Category Service
type CategoryService struct {
	categoryRepo *repository.CategoryRepository
}

func NewCategoryService(categoryRepo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{categoryRepo: categoryRepo}
}

type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateCategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (s *CategoryService) CreateCategory(req *CreateCategoryRequest) (*models.Category, error) {
	category := &models.Category{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.categoryRepo.Create(category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *CategoryService) GetCategoryByID(id uuid.UUID) (*models.Category, error) {
	return s.categoryRepo.GetByID(id)
}

func (s *CategoryService) GetAllCategories() ([]models.Category, error) {
	return s.categoryRepo.GetAll()
}

func (s *CategoryService) UpdateCategory(id uuid.UUID, req *UpdateCategoryRequest) (*models.Category, error) {
	category, err := s.categoryRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, fmt.Errorf("category not found")
	}

	if req.Name != "" {
		category.Name = req.Name
	}
	if req.Description != "" {
		category.Description = req.Description
	}

	if err := s.categoryRepo.Update(category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *CategoryService) DeleteCategory(id uuid.UUID) error {
	return s.categoryRepo.Delete(id)
}

// Product Review Service
type ProductReviewService struct {
	reviewRepo   *repository.ProductReviewRepository
	productRepo  *repository.ProductRepository
}

func NewProductReviewService(reviewRepo *repository.ProductReviewRepository, productRepo *repository.ProductRepository) *ProductReviewService {
	return &ProductReviewService{
		reviewRepo:  reviewRepo,
		productRepo: productRepo,
	}
}

type CreateReviewRequest struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Rating    int       `json:"rating" binding:"required,min=1,max=5"`
	Comment   string    `json:"comment"`
}

type UpdateReviewRequest struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment"`
}

func (s *ProductReviewService) CreateReview(userID uuid.UUID, req *CreateReviewRequest) (*models.ProductReview, error) {
	// Check if product exists
	product, err := s.productRepo.GetByID(req.ProductID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, fmt.Errorf("product not found")
	}

	// Check if user already reviewed this product
	hasReviewed, err := s.reviewRepo.HasUserReviewed(userID, req.ProductID)
	if err != nil {
		return nil, err
	}
	if hasReviewed {
		return nil, fmt.Errorf("user already reviewed this product")
	}

	review := &models.ProductReview{
		ID:        uuid.New(),
		ProductID: req.ProductID,
		UserID:    userID,
		Rating:    req.Rating,
		Comment:   req.Comment,
	}

	if err := s.reviewRepo.Create(review); err != nil {
		return nil, err
	}

	return review, nil
}

func (s *ProductReviewService) GetProductReviews(productID uuid.UUID, page, limit int) ([]models.ProductReview, int64, error) {
	return s.reviewRepo.GetByProduct(productID, page, limit)
}

func (s *ProductReviewService) GetUserReviews(userID uuid.UUID, page, limit int) ([]models.ProductReview, int64, error) {
	return s.reviewRepo.GetByUser(userID, page, limit)
}

func (s *ProductReviewService) UpdateReview(reviewID uuid.UUID, userID uuid.UUID, req *UpdateReviewRequest) (*models.ProductReview, error) {
	review, err := s.reviewRepo.GetByID(reviewID)
	if err != nil {
		return nil, err
	}
	if review == nil {
		return nil, fmt.Errorf("review not found")
	}

	// Check if user owns this review
	if review.UserID != userID {
		return nil, fmt.Errorf("unauthorized")
	}

	review.Rating = req.Rating
	review.Comment = req.Comment

	if err := s.reviewRepo.Update(review); err != nil {
		return nil, err
	}

	return review, nil
}

func (s *ProductReviewService) DeleteReview(reviewID uuid.UUID, userID uuid.UUID) error {
	review, err := s.reviewRepo.GetByID(reviewID)
	if err != nil {
		return err
	}
	if review == nil {
		return fmt.Errorf("review not found")
	}

	// Check if user owns this review
	if review.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	return s.reviewRepo.Delete(reviewID)
}