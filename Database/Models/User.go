package Models

import (
	config "github.com/ahmedfargh/server-manager/Config"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username           string       `gorm:"unique;not null" json:"username" form:"username" validate:"required,min=3,max=32"`
	Email              string       `gorm:"unique;not null" json:"email" form:"email" validate:"required,email"`
	Password           string       `gorm:"not null" json:"password" form:"password" validate:"required,min=8"`
	RoleID             uint         `gorm:"not null" json:"role_id" form:"role_id"`
	Role               Role         // Belongs to Role
	ImagePath          string       `json:"image_path"`
	Status             bool         `gorm:"default:false" json:"status" form:"status"`
	Permissions        []Permission `gorm:"many2many:user_permissions;"`
	NotificationDriver string       `json:"notification_driver" default:"Telegram"`
	TelegramChatID     string       `json:"telegram_chat_id" default:"null"`
	TelegramBotToken   string       `json:"telegram_bot_token" default:"null"`
	DiscordChannelID   string       `json:"discord_channel_id" default:"null"`
	DiscordBotToken    string       `json:"discord_bot_token" default:"null"`
	WebHookURL         string       `json:"webhook_url" default:"null"`
	WebHookSecret      string       `json:"webhook_secret" default:"null"`
}

func (User) TableName() string {
	return "users"
}

func (u User) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":                  u.ID,
		"username":            u.Username,
		"email":               u.Email,
		"role":                u.Role.Name,
		"image_path":          config.GetKey("APP_URL") + u.ImagePath,
		"status":              u.Status,
		"permissions":         u.Permissions,
		"roles_permssions":    u.Role.Permissions,
		"notification_driver": u.NotificationDriver,
		"telegram_chat_id":    u.TelegramChatID,
		"telegram_bot_token":  u.TelegramBotToken,
		"discord_channel_id":  u.DiscordChannelID,
		"discord_bot_token":   u.DiscordBotToken,
		"webhook_url":         u.WebHookURL,
		"webhook_secret":      u.WebHookSecret,
	}
}
func (u User) HasPermission(permission string) bool {
	for _, p := range u.Permissions {
		if p.Name == permission {
			return true
		}
	}
	for _, p := range u.Role.Permissions {
		if p.Name == permission {
			return true
		}
	}

	return false
}
