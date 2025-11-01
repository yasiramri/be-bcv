package repository

import (
	"errors"

	"github.com/be-bcv/ecommerce-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, id).Error
}

func (r *UserRepository) GetAll(page, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	offset := (page - 1) * limit

	if err := r.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Offset(offset).Limit(limit).Find(&users).Error
	return users, total, err
}

func (r *UserRepository) CreateSession(session *models.UserSession) error {
	return r.db.Create(session).Error
}

func (r *UserRepository) GetSessionByToken(token string) (*models.UserSession, error) {
	var session models.UserSession
	err := r.db.Preload("User").Where("token = ? AND expires_at > NOW()", token).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func (r *UserRepository) DeleteSessionByToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&models.UserSession{}).Error
}

func (r *UserRepository) DeleteAllUserSessions(userID uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.UserSession{}).Error
}