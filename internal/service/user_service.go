package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/be-bcv/ecommerce-backend/internal/models"
	"github.com/be-bcv/ecommerce-backend/internal/repository"
	"github.com/be-bcv/ecommerce-backend/pkg/config"
	"github.com/be-bcv/ecommerce-backend/pkg/messages"
	"github.com/be-bcv/ecommerce-backend/pkg/rabbitmq"
	"github.com/be-bcv/ecommerce-backend/pkg/redis"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo  *repository.UserRepository
	redis     *redis.RedisClient
	rabbitmq  *rabbitmq.RabbitMQ
	config    *config.Config
}

func NewUserService(userRepo *repository.UserRepository, redis *redis.RedisClient, rabbitmq *rabbitmq.RabbitMQ, config *config.Config) *UserService {
	return &UserService{
		userRepo: userRepo,
		redis:    redis,
		rabbitmq: rabbitmq,
		config:   config,
	}
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	User         models.User `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
}

func (s *UserService) Register(req *RegisterRequest) (*AuthResponse, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("user already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		ID:       uuid.New(),
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Phone:    req.Phone,
		Address:  req.Address,
		Role:     "user",
		IsActive: true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Generate tokens
	accessToken, refreshToken, err := s.generateTokens(user)
	if err != nil {
		return nil, err
	}

	// Store refresh token
	if err := s.storeRefreshToken(user.ID, refreshToken); err != nil {
		return nil, err
	}

	// Publish user registered event
	s.publishUserRegisteredEvent(user)

	// Clear password for response
	user.Password = ""

	return &AuthResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserService) Login(req *LoginRequest) (*AuthResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("user account is deactivated")
	}

	// Generate tokens
	accessToken, refreshToken, err := s.generateTokens(user)
	if err != nil {
		return nil, err
	}

	// Store refresh token
	if err := s.storeRefreshToken(user.ID, refreshToken); err != nil {
		return nil, err
	}

	// Clear password for response
	user.Password = ""

	return &AuthResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserService) Logout(refreshToken string) error {
	// Remove refresh token from Redis
	ctx := context.Background()
	key := fmt.Sprintf("refresh_token:%s", refreshToken)
	return s.redis.Del(ctx, key)
}

func (s *UserService) RefreshToken(refreshToken string) (*AuthResponse, error) {
	// Validate refresh token
	token, err := jwt.ParseWithClaims(refreshToken, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	userIDStr, ok := (*claims)["user_id"].(string)
	if !ok {
		return nil, errors.New("invalid user ID in token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Check if refresh token exists in Redis
	ctx := context.Background()
	key := fmt.Sprintf("refresh_token:%s", refreshToken)
	exists, err := s.redis.Exists(ctx, key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("refresh token not found")
	}

	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Generate new tokens
	accessToken, newRefreshToken, err := s.generateTokens(user)
	if err != nil {
		return nil, err
	}

	// Remove old refresh token
	s.redis.Del(ctx, key)

	// Store new refresh token
	if err := s.storeRefreshToken(user.ID, newRefreshToken); err != nil {
		return nil, err
	}

	// Clear password for response
	user.Password = ""

	return &AuthResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *UserService) GetProfile(userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Clear password
	user.Password = ""
	return user, nil
}

func (s *UserService) UpdateProfile(userID uuid.UUID, updates map[string]interface{}) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Update fields
	if name, ok := updates["name"].(string); ok {
		user.Name = name
	}
	if phone, ok := updates["phone"].(string); ok {
		user.Phone = phone
	}
	if address, ok := updates["address"].(string); ok {
		user.Address = address
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	// Clear password
	user.Password = ""
	return user, nil
}

func (s *UserService) DeleteAccount(userID uuid.UUID) error {
	return s.userRepo.Delete(userID)
}

func (s *UserService) GetAllUsers(page, limit int) ([]models.User, int64, error) {
	users, total, err := s.userRepo.GetAll(page, limit)
	if err != nil {
		return nil, 0, err
	}

	// Clear passwords
	for i := range users {
		users[i].Password = ""
	}

	return users, total, nil
}

func (s *UserService) GetUserByID(userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Clear password
	user.Password = ""
	return user, nil
}

func (s *UserService) UpdateUserStatus(userID uuid.UUID, isActive bool) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	user.IsActive = isActive
	return s.userRepo.Update(user)
}

func (s *UserService) generateTokens(user *models.User) (string, string, error) {
	// Generate access token
	accessClaims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // 24 hours
		"iat":     time.Now().Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", "", err
	}

	// Generate refresh token
	refreshClaims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
		"iat":     time.Now().Unix(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

func (s *UserService) storeRefreshToken(userID uuid.UUID, refreshToken string) error {
	ctx := context.Background()
	key := fmt.Sprintf("refresh_token:%s", refreshToken)
	return s.redis.Set(ctx, key, userID.String(), time.Hour*24*7)
}

func (s *UserService) publishUserRegisteredEvent(user *models.User) {
	event := messages.EventMessage{
		EventID:   uuid.New().String(),
		EventName: "user.registered",
		Timestamp: time.Now(),
		Data: messages.UserRegisteredEvent{
			UserID: user.ID.String(),
			Email:  user.Email,
			Name:   user.Name,
		},
		Service: "user-service",
	}

	// Publish to RabbitMQ
	// s.rabbitmq.Publish("user_events", "user.registered", event)
}