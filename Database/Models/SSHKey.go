package Models

import (
	"time"

	"gorm.io/gorm"
)

type SSHKey struct {
	ID                    uint           `gorm:"primaryKey" json:"id"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`
	Name                  string         `gorm:"not null" json:"name"`
	KeyType               string         `gorm:"not null" json:"key_type"` // ed25519, rsa
	PublicKey             string         `gorm:"type:text;not null" json:"public_key"`
	Fingerprint           string         `gorm:"index;not null" json:"fingerprint"`
	Comment               string         `json:"comment"`
	AddedToAuthorizedKeys bool           `gorm:"default:false" json:"added_to_authorized_keys"`
	UserID                uint           `json:"user_id"`
}

func (SSHKey) TableName() string {
	return "ssh_keys"
}
