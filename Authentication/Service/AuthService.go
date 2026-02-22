package service

import (
	"errors"
	"fmt"
	"time"

	config "github.com/ahmedfargh/server-manager/Config"
	crud "github.com/ahmedfargh/server-manager/Database/CRUD"
	models "github.com/ahmedfargh/server-manager/Database/Models"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Claims defines the structure of the JWT claims
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type AuthService struct {
	UserCRUD *crud.UserCRUD
	RoleCRUD *crud.RoleCRUD
}

func NewAuthService(userCRUD *crud.UserCRUD, roleCRUD *crud.RoleCRUD) *AuthService {
	return &AuthService{UserCRUD: userCRUD, RoleCRUD: roleCRUD}
}

func (s *AuthService) Register(user *models.User) error {
	validate := validator.New()
	if err := validate.Struct(user); err != nil {
		return err
	}

	role, err := s.RoleCRUD.FindOrCreateRole(user.Role.Name)
	if err != nil {
		return fmt.Errorf("failed to get or create role: %w", err)
	}
	user.RoleID = role.ID

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	return s.UserCRUD.CreateUser(user)
}

func (s *AuthService) Login(email, password string) (map[string]interface{}, error) {
	user, err := s.UserCRUD.GetUserByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials") // User not found
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid credentials") // Passwords do not match
	}

	// Generate JWT token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.JwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Prepare response data
	response := user.ToMap()
	response["token"] = tokenString

	return response, nil
}
