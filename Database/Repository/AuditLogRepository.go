package Repository

import (
	"github.com/ahmedfargh/server-manager/Database/Models"
	"gorm.io/gorm"
)

type AuditRepository struct {
	DB *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{DB: db}
}
func (r *AuditRepository) Paginate(page int, limit int) ([]Models.AuditLog, error) {
	var audits []Models.AuditLog
	offset := (page - 1) * limit
	err := r.DB.Offset(offset).Limit(limit).Find(&audits).Error
	if err != nil {
		return nil, err
	}
	return audits, nil
}
func (r *AuditRepository) CreateAudit(audit *Models.AuditLog) error {
	return r.DB.Create(audit).Error
}
func (r *AuditRepository) GetAudits(page int, limit int, Type string) ([]Models.AuditLog, error) {
	var audits []Models.AuditLog
	offset := (page - 1) * limit
	query := r.DB.Preload("User").Offset(offset).Limit(limit)

	if Type != "" {
		query = query.Where("service_type = ?", Type)
	}

	err := query.Find(&audits).Error
	if err != nil {
		return nil, err
	}
	return audits, nil
}

func (r *AuditRepository) GetAuditByID(id uint) (*Models.AuditLog, error) {
	var audit Models.AuditLog
	err := r.DB.First(&audit, id).Error
	if err != nil {
		return nil, err
	}
	return &audit, nil
}
func (r *AuditRepository) UpdateAudit(audit *Models.AuditLog, id uint) error {
	return r.DB.Save(audit).Error
}
func (r *AuditRepository) DeleteAudit(audit *Models.AuditLog) error {
	return r.DB.Delete(audit).Error
}
func (r *AuditRepository) DeleteAllAudits() error {
	return r.DB.Delete(&Models.AuditLog{}).Error
}
