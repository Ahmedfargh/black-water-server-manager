package service

import (
	"errors"
	"fmt"
	"math/rand/v2"
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

	// Create user first
	if err := s.UserCRUD.CreateUser(user); err != nil {
		return err
	}

	// Send verification email after successful registration
	if err := s.sendVerificationEmail(user); err != nil {
		// Log error but don't fail registration
		fmt.Printf("Failed to send verification email: %v\n", err)
	}

	return nil
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

	// Check if email is verified
	if !user.EmailVerified {
		go s.GenerateVerificationCode(user.ID) // Generate OTP code in the background

		return map[string]interface{}{
			"requires_verification": true,
			"user_id":               user.ID,
			"email":                 user.Email,
			"username":              user.Username,
		}, nil
	}

	// Generate JWT token
	if user.OTPEnabled {
		requiresOTP := true
		if user.EmailVerifiedAt != nil {
			if time.Since(*user.EmailVerifiedAt) < 5*time.Minute {
				requiresOTP = false
			}
		}

		if requiresOTP {
			go s.GenerateVerificationCode(user.ID) // Generate OTP code in the background
			return map[string]interface{}{
				"requires_otp": true,
				"user_id":      user.ID,
				"email":        user.Email,
				"username":     user.Username,
			}, nil
		}
	}
	return s.GenerateTokenResponse(user)
}

func (s *AuthService) GenerateTokenResponse(user *models.User) (map[string]interface{}, error) {
	expirationTime := time.Now().Add(3 * time.Hour)
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
	response["otp_enabled"] = user.OTPEnabled
	return response, nil
}
func (s *AuthService) GenerateVerificationCode(userID uint) (bool, error) {

	user, err := s.UserCRUD.GetUserByID(userID)
	if err != nil {
		fmt.Println("error loading user")
		return false, err
	}

	code := randString()
	otpSecret, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error hashing OTP code:", err)
		return false, err
	}
	user.OTPSecret = string(otpSecret) // Invalidate the code after use
	err = s.UserCRUD.UpdateUser(user, userID)
	if err != nil {
		fmt.Println("Error updating user with OTP secret:", err)
		return false, err
	}
	MailService := NewMailService()
	MailService.From = config.GetKey("MAIL_FROM")
	MailService.password = config.GetKey("MAIL_PASSWORD")
	MailService.smtpHost = config.GetKey("MAIL_SMTP_HOST")
	MailService.smtpPort = config.GetKey("MAIL_SMTP_PORT")
	MailService.To = user.Email
	err = MailService.SendOTPEmail(user, code, "Login")
	if err != nil {
		fmt.Println("Error sending verification email:", err)
		return false, err
	}
	return true, nil // Valid code
}
func randString() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.IntN(len(charset))]
	}
	return string(b)
}
func (s *AuthService) UpdateNotificationSettings(userID uint, driver, telegramToken, telegramChatID, discordToken, discordChannelID, webhookURL, webhookSecret string) error {
	user, err := s.UserCRUD.GetUserByID(userID)
	if err != nil {
		return err
	}

	user.NotificationDriver = driver
	user.TelegramBotToken = telegramToken
	user.TelegramChatID = telegramChatID
	user.DiscordBotToken = discordToken
	user.DiscordChannelID = discordChannelID
	user.WebHookURL = webhookURL
	user.WebHookSecret = webhookSecret

	return s.UserCRUD.UpdateUser(user, userID)
}

func (s *AuthService) VerifyEmail(userID uint, code string) error {
	user, err := s.UserCRUD.GetUserByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if user.EmailVerified {
		return errors.New("email already verified")
	}

	// Check if OTP secret exists and matches
	if user.OTPSecret == "" {
		return errors.New("no verification code found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.OTPSecret), []byte(code))
	if err != nil {
		return errors.New("invalid verification code")
	}

	// Mark email as verified and clear OTP secret
	now := time.Now()
	user.EmailVerified = true
	user.EmailVerifiedAt = &now
	user.OTPSecret = ""
	return s.UserCRUD.UpdateUser(user, userID)
}

func (s *AuthService) VerifyOTP(userID uint, code string) (map[string]interface{}, error) {
	user, err := s.UserCRUD.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if user.OTPSecret == "" {
		return nil, errors.New("no verification code found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.OTPSecret), []byte(code))
	if err != nil {
		return nil, errors.New("invalid verification code")
	}

	// Code is valid, now clear it and return token
	user.OTPSecret = ""
	err = s.UserCRUD.UpdateUser(user, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return s.GenerateTokenResponse(user)
}

func (s *AuthService) ResendVerificationCode(userID uint) error {
	user, err := s.UserCRUD.GetUserByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if user.EmailVerified {
		return errors.New("email already verified")
	}

	// Generate new verification code
	return s.sendVerificationEmail(user)
}

func (s *AuthService) sendVerificationEmail(user *models.User) error {
	code := randString()
	otpSecret, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash verification code: %w", err)
	}

	user.OTPSecret = string(otpSecret)
	err = s.UserCRUD.UpdateUser(user, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	MailService := NewMailService()
	MailService.From = config.GetKey("MAIL_FROM")
	MailService.password = config.GetKey("MAIL_PASSWORD")
	MailService.smtpHost = config.GetKey("MAIL_SMTP_HOST")
	MailService.smtpPort = config.GetKey("MAIL_SMTP_PORT")
	MailService.To = user.Email

	err = MailService.SendVerificationEmail(user, code)
	if err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}
