package repository

import (
	models "github.com/ahmedfargh/server-manager/Database/Models"
	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) CreateUser(user *models.User) error {
	return r.DB.Create(user).Error
}

func (r *UserRepository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	err := r.DB.Preload("Role.Permissions").Preload("Permissions").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUsers() ([]models.User, error) {
	var users []models.User
	err := r.DB.Preload("Role.Permissions").Preload("Permissions").Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) UpdateUser(user *models.User, id uint) error {
	return r.DB.Save(user).Error
}

func (r *UserRepository) GetPaginatedUsers(page, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	r.DB.Model(&models.User{}).Count(&total)

	offset := (page - 1) * limit
	err := r.DB.Preload("Role.Permissions").Preload("Permissions").Offset(offset).Limit(limit).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *UserRepository) DeleteUser(user *models.User) error {
	return r.DB.Delete(user).Error
}

func (r *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.DB.Preload("Role.Permissions").Preload("Permissions").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.DB.Preload("Role.Permissions").Preload("Permissions").Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
