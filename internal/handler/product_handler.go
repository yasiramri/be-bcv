package handler

import (
	"net/http"
	"strconv"

	"github.com/be-bcv/ecommerce-backend/internal/service"
	"github.com/be-bcv/ecommerce-backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProductHandler struct {
	productService *service.ProductService
}

func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req service.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	product, err := h.productService.CreateProduct(&req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to create product", err.Error())
		return
	}

	utils.SuccessResponse(c, "Product created successfully", product)
}

func (h *ProductHandler) GetProductByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID", err.Error())
		return
	}

	product, err := h.productService.GetProductByID(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Product not found", err.Error())
		return
	}

	utils.SuccessResponse(c, "Product retrieved successfully", product)
}

func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	categoryIDStr := c.Query("category_id")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	var categoryID uuid.UUID
	if categoryIDStr != "" {
		categoryID, err = uuid.Parse(categoryIDStr)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid category ID", err.Error())
			return
		}
	}

	products, total, err := h.productService.GetAllProducts(page, limit, categoryID, sortBy, sortOrder)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch products", err.Error())
		return
	}

	pagination := utils.NewPagination(page, limit, int(total))
	utils.PagedResponse(c, "Products retrieved successfully", products, pagination)
}

func (h *ProductHandler) SearchProducts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Search query is required", nil)
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	products, total, err := h.productService.SearchProducts(query, page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to search products", err.Error())
		return
	}

	pagination := utils.NewPagination(page, limit, int(total))
	utils.PagedResponse(c, "Search results", products, pagination)
}

func (h *ProductHandler) GetProductsByCategory(c *gin.Context) {
	categoryIDStr := c.Param("categoryId")
	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid category ID", err.Error())
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	products, total, err := h.productService.GetProductsByCategory(categoryID, page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch products by category", err.Error())
		return
	}

	pagination := utils.NewPagination(page, limit, int(total))
	utils.PagedResponse(c, "Products by category retrieved successfully", products, pagination)
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID", err.Error())
		return
	}

	var req service.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	product, err := h.productService.UpdateProduct(id, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to update product", err.Error())
		return
	}

	utils.SuccessResponse(c, "Product updated successfully", product)
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID", err.Error())
		return
	}

	if err := h.productService.DeleteProduct(id); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to delete product", err.Error())
		return
	}

	utils.SuccessResponse(c, "Product deleted successfully", nil)
}

func (h *ProductHandler) UpdateStock(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID", err.Error())
		return
	}

	var req service.UpdateStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	if err := h.productService.UpdateStock(id, &req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to update stock", err.Error())
		return
	}

	utils.SuccessResponse(c, "Stock updated successfully", nil)
}

// Category Handlers
type CategoryHandler struct {
	categoryService *service.CategoryService
}

func NewCategoryHandler(categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var req service.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	category, err := h.categoryService.CreateCategory(&req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to create category", err.Error())
		return
	}

	utils.SuccessResponse(c, "Category created successfully", category)
}

func (h *CategoryHandler) GetCategoryByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid category ID", err.Error())
		return
	}

	category, err := h.categoryService.GetCategoryByID(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Category not found", err.Error())
		return
	}

	utils.SuccessResponse(c, "Category retrieved successfully", category)
}

func (h *CategoryHandler) GetAllCategories(c *gin.Context) {
	categories, err := h.categoryService.GetAllCategories()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch categories", err.Error())
		return
	}

	utils.SuccessResponse(c, "Categories retrieved successfully", categories)
}

func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid category ID", err.Error())
		return
	}

	var req service.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	category, err := h.categoryService.UpdateCategory(id, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to update category", err.Error())
		return
	}

	utils.SuccessResponse(c, "Category updated successfully", category)
}

func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid category ID", err.Error())
		return
	}

	if err := h.categoryService.DeleteCategory(id); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to delete category", err.Error())
		return
	}

	utils.SuccessResponse(c, "Category deleted successfully", nil)
}

// Product Review Handlers
type ProductReviewHandler struct {
	reviewService *service.ProductReviewService
}

func NewProductReviewHandler(reviewService *service.ProductReviewService) *ProductReviewHandler {
	return &ProductReviewHandler{reviewService: reviewService}
}

func (h *ProductReviewHandler) CreateReview(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	var req service.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	review, err := h.reviewService.CreateReview(userID, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to create review", err.Error())
		return
	}

	utils.SuccessResponse(c, "Review created successfully", review)
}

func (h *ProductReviewHandler) GetProductReviews(c *gin.Context) {
	productIDStr := c.Param("id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID", err.Error())
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	reviews, total, err := h.reviewService.GetProductReviews(productID, page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch product reviews", err.Error())
		return
	}

	pagination := utils.NewPagination(page, limit, int(total))
	utils.PagedResponse(c, "Product reviews retrieved successfully", reviews, pagination)
}

func (h *ProductReviewHandler) UpdateReview(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	reviewIDStr := c.Param("reviewId")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid review ID", err.Error())
		return
	}

	var req service.UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	review, err := h.reviewService.UpdateReview(reviewID, userID, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to update review", err.Error())
		return
	}

	utils.SuccessResponse(c, "Review updated successfully", review)
}

func (h *ProductReviewHandler) DeleteReview(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	reviewIDStr := c.Param("reviewId")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid review ID", err.Error())
		return
	}

	if err := h.reviewService.DeleteReview(reviewID, userID); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to delete review", err.Error())
		return
	}

	utils.SuccessResponse(c, "Review deleted successfully", nil)
}