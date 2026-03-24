package Models

import "gorm.io/gorm"

type Site struct {
	gorm.Model
	Name         string `gorm:"unique;not null;length:255" json:"name"`
	URL          string `gorm:"unique;not null" json:"url"`
	Health_Route string `gorm:"unique;not null" json:"health_route"`
	Description  string `gorm:"not null" json:"description"`
}

func (Site) TableName() string {
	return "sites"
}
func NewSite() Site {
	return Site{}
}
