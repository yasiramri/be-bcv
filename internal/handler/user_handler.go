package handler

import (
	"net/http"
	"strconv"

	"github.com/be-bcv/ecommerce-backend/internal/service"
	"github.com/be-bcv/ecommerce-backend/pkg/middleware"
	"github.com/be-bcv/ecommerce-backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	response, err := h.userService.Register(&req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Registration failed", err.Error())
		return
	}

	utils.SuccessResponse(c, "User registered successfully", response)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	response, err := h.userService.Login(&req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Login failed", err.Error())
		return
	}

	utils.SuccessResponse(c, "Login successful", response)
}

func (h *UserHandler) Logout(c *gin.Context) {
	refreshToken := c.GetHeader("Refresh-Token")
	if refreshToken == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Refresh token required", nil)
		return
	}

	if err := h.userService.Logout(refreshToken); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Logout failed", err.Error())
		return
	}

	utils.SuccessResponse(c, "Logout successful", nil)
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	response, err := h.userService.RefreshToken(req.RefreshToken)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Token refresh failed", err.Error())
		return
	}

	utils.SuccessResponse(c, "Token refreshed successfully", response)
}

func (h *UserHandler) GetProfile(c *gin.Context) {
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

	user, err := h.userService.GetProfile(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User not found", err.Error())
		return
	}

	utils.SuccessResponse(c, "Profile retrieved successfully", user)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
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

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	user, err := h.userService.UpdateProfile(userID, updates)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update profile", err.Error())
		return
	}

	utils.SuccessResponse(c, "Profile updated successfully", user)
}

func (h *UserHandler) DeleteAccount(c *gin.Context) {
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

	if err := h.userService.DeleteAccount(userID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete account", err.Error())
		return
	}

	utils.SuccessResponse(c, "Account deleted successfully", nil)
}

func (h *UserHandler) GetAllUsers(c *gin.Context) {
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

	users, total, err := h.userService.GetAllUsers(page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch users", err.Error())
		return
	}

	pagination := utils.NewPagination(page, limit, int(total))
	utils.PagedResponse(c, "Users retrieved successfully", users, pagination)
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User not found", err.Error())
		return
	}

	utils.SuccessResponse(c, "User retrieved successfully", user)
}

func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	var req struct {
		IsActive bool `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	if err := h.userService.UpdateUserStatus(userID, req.IsActive); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update user status", err.Error())
		return
	}

	utils.SuccessResponse(c, "User status updated successfully", nil)
}