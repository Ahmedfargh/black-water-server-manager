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
	Expected_Status int    `gorm:"default:200" json:"expected_status"`
	Status       string `gorm:"-" json:"status"`
	LastChecked  string `gorm:"-" json:"last_checked"`
}

func (Site) TableName() string {
	return "sites"
}
func NewSite() Site {
	return Site{}
}
