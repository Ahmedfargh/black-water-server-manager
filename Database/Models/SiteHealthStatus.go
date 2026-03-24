package Models

type SiteHealthStatus struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	SiteID uint   `gorm:"not null" json:"site_id"`
	Site   Site   `gorm:"foreignKey:SiteID" json:"site"`
	Status string `gorm:"not null" json:"status"`
	Time   string `gorm:"not null" json:"time"`
}

func (SiteHealthStatus) TableName() string {
	return "site_health_statuses"
}
