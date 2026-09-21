package Models

import (
	"time"

	"gorm.io/gorm"
)

type Site struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Name         string `gorm:"unique;not null;length:255" json:"name"`
	URL          string `gorm:"unique;not null" json:"url"`
	Health_Route    string `gorm:"unique;not null" json:"health_route"`
	Description     string `gorm:"not null" json:"description"`
	Method          string `gorm:"default:GET" json:"method"`
	Expected_Status int        `gorm:"default:200" json:"expected_status"`
	Status          string     `gorm:"-" json:"status"`
	LastChecked     string     `gorm:"-" json:"last_checked"`
	SSLEnabled      bool       `gorm:"default:false" json:"ssl_enabled"`
	SSLExpiryDate   *time.Time `json:"ssl_expiry_date"`
	SSLIssuer       string     `json:"ssl_issuer"`
	SSLDomains      string     `json:"ssl_domains"`
	SSLAutoRenew    bool       `gorm:"default:false" json:"ssl_auto_renew"`
	SSLCertPath     string     `json:"ssl_cert_path"`
	SSLKeyPath      string     `json:"ssl_key_path"`
	SSLLastCheck    *time.Time `json:"ssl_last_check"`
}

func (Site) TableName() string {
	return "sites"
}
func NewSite() Site {
	return Site{}
}
