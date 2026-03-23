package Models

import "gorm.io/gorm"

type AuditLog struct {
	gorm.Model
	UserID      *uint  `json:"user_id"`
	ServiceType string `json:"service_type"`
	ServiceID   string `json:"service_id"`
	Action      string `json:"action"`
	User        User   `gorm:"foreignKey:UserID"`
}

func (AuditLog) TableName() string {
	return "audit_logs"

}
func NewAuditLog() AuditLog {
	return AuditLog{}

}
