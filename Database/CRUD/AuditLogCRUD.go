package crud

import (
	models "github.com/ahmedfargh/server-manager/Database/Models"
	repository "github.com/ahmedfargh/server-manager/Database/Repository"
)

type AuditLogCRUD struct {
	Repo *repository.AuditRepository
}

func NewAuditLogCRUD(repo *repository.AuditRepository) *AuditLogCRUD {
	return &AuditLogCRUD{Repo: repo}
}

func (c *AuditLogCRUD) CreateAudit(audit *models.AuditLog) error {
	return c.Repo.CreateAudit(audit)
}

func (c *AuditLogCRUD) GetAudits(page int, limit int, Type string) ([]models.AuditLog, error) {
	return c.Repo.GetAudits(page, limit, Type)
}

func (c *AuditLogCRUD) GetAuditByID(id uint) (*models.AuditLog, error) {
	return c.Repo.GetAuditByID(id)
}

func (c *AuditLogCRUD) UpdateAudit(audit *models.AuditLog, id uint) error {
	return c.Repo.UpdateAudit(audit, id)
}

func (c *AuditLogCRUD) DeleteAudit(audit *models.AuditLog) error {
	return c.Repo.DeleteAudit(audit)
}

func (c *AuditLogCRUD) DeleteAllAudits() error {
	return c.Repo.DeleteAllAudits()
}

func (c *AuditLogCRUD) Paginate(page int, limit int) ([]models.AuditLog, error) {
	return c.Repo.Paginate(page, limit)

}
