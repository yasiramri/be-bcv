package repository

import (
	"errors"

	"github.com/be-bcv/ecommerce-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(category *models.Category) error {
	return r.db.Create(category).Error
}

func (r *CategoryRepository) GetByID(id uuid.UUID) (*models.Category, error) {
	var category models.Category
	err := r.db.Where("id = ?", id).First(&category).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *CategoryRepository) GetAll() ([]models.Category, error) {
	var categories []models.Category
	err := r.db.Find(&categories).Error
	return categories, err
}

func (r *CategoryRepository) Update(category *models.Category) error {
	return r.db.Save(category).Error
}

func (r *CategoryRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Category{}, id).Error
}

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(product *models.Product) error {
	return r.db.Create(product).Error
}

func (r *ProductRepository) GetByID(id uuid.UUID) (*models.Product, error) {
	var product models.Product
	err := r.db.Preload("Category").Where("id = ? AND is_active = ?", id, true).First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) GetAll(page, limit int, categoryID uuid.UUID, sortBy string, sortOrder string) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	query := r.db.Model(&models.Product{}).Preload("Category").Where("is_active = ?", true)

	// Filter by category
	if categoryID != uuid.Nil {
		query = query.Where("category_id = ?", categoryID)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sorting
	switch sortBy {
	case "price":
		if sortOrder == "desc" {
			query = query.Order("price desc")
		} else {
			query = query.Order("price asc")
		}
	case "created_at":
		if sortOrder == "desc" {
			query = query.Order("created_at desc")
		} else {
			query = query.Order("created_at asc")
		}
	case "name":
		if sortOrder == "desc" {
			query = query.Order("name desc")
		} else {
			query = query.Order("name asc")
		}
	default:
		query = query.Order("created_at desc")
	}

	// Pagination
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Find(&products).Error

	return products, total, err
}

func (r *ProductRepository) Search(query string, page, limit int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	searchQuery := "%" + query + "%"

	dbQuery := r.db.Model(&models.Product{}).
		Preload("Category").
		Where("is_active = ? AND (name ILIKE ? OR description ILIKE ?)", true, searchQuery, searchQuery)

	// Count total
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	offset := (page - 1) * limit
	err := dbQuery.Offset(offset).Limit(limit).Find(&products).Error

	return products, total, err
}

func (r *ProductRepository) GetByCategory(categoryID uuid.UUID, page, limit int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	query := r.db.Model(&models.Product{}).
		Preload("Category").
		Where("category_id = ? AND is_active = ?", categoryID, true)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Find(&products).Error

	return products, total, err
}

func (r *ProductRepository) GetBySeller(sellerID uuid.UUID, page, limit int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	query := r.db.Model(&models.Product{}).
		Preload("Category").
		Where("seller_id = ?", sellerID)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Find(&products).Error

	return products, total, err
}

func (r *ProductRepository) Update(product *models.Product) error {
	return r.db.Save(product).Error
}

func (r *ProductRepository) UpdateStock(productID uuid.UUID, newStock int) error {
	return r.db.Model(&models.Product{}).
		Where("id = ?", productID).
		Update("stock", newStock).Error
}

func (r *ProductRepository) Delete(id uuid.UUID) error {
	// Soft delete
	return r.db.Model(&models.Product{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

func (r *ProductRepository) GetBySKU(sku string) (*models.Product, error) {
	var product models.Product
	err := r.db.Preload("Category").Where("sku = ? AND is_active = ?", sku, true).First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

type ProductReviewRepository struct {
	db *gorm.DB
}

func NewProductReviewRepository(db *gorm.DB) *ProductReviewRepository {
	return &ProductReviewRepository{db: db}
}

func (r *ProductReviewRepository) Create(review *models.ProductReview) error {
	return r.db.Create(review).Error
}

func (r *ProductReviewRepository) GetByID(id uuid.UUID) (*models.ProductReview, error) {
	var review models.ProductReview
	err := r.db.Preload("Product").Where("id = ?", id).First(&review).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &review, nil
}

func (r *ProductReviewRepository) GetByProduct(productID uuid.UUID, page, limit int) ([]models.ProductReview, int64, error) {
	var reviews []models.ProductReview
	var total int64

	query := r.db.Model(&models.ProductReview{}).
		Preload("Product").
		Where("product_id = ?", productID)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at desc").Find(&reviews).Error

	return reviews, total, err
}

func (r *ProductReviewRepository) GetByUser(userID uuid.UUID, page, limit int) ([]models.ProductReview, int64, error) {
	var reviews []models.ProductReview
	var total int64

	query := r.db.Model(&models.ProductReview{}).
		Preload("Product").
		Where("user_id = ?", userID)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at desc").Find(&reviews).Error

	return reviews, total, err
}

func (r *ProductReviewRepository) Update(review *models.ProductReview) error {
	return r.db.Save(review).Error
}

func (r *ProductReviewRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.ProductReview{}, id).Error
}

func (r *ProductReviewRepository) GetAverageRating(productID uuid.UUID) (float64, error) {
	var avgRating float64
	err := r.db.Model(&models.ProductReview{}).
		Where("product_id = ?", productID).
		Select("AVG(rating)").
		Scan(&avgRating).Error
	return avgRating, err
}

func (r *ProductReviewRepository) HasUserReviewed(userID, productID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.ProductReview{}).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Count(&count).Error
	return count > 0, err
}