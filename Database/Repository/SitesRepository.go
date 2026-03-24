package Repository

import (
	Config "github.com/ahmedfargh/server-manager/Config"
	models "github.com/ahmedfargh/server-manager/Database/Models"
	"gorm.io/gorm"
)

type SiteRepository struct {
	DB *gorm.DB
}

func NewSiteRepository(db *gorm.DB) *SiteRepository {
	return &SiteRepository{DB: db}
}

func (r *SiteRepository) CreateSite(user *models.Site) error {
	return r.DB.Create(user).Error
}

func (r *SiteRepository) GetSiteByID(id uint) (*models.Site, error) {
	var site models.Site
	err := r.DB.First(&site, id).Error
	if err != nil {
		return nil, err
	}
	return &site, nil
}
func (r *SiteRepository) GetSites(page uint, limit uint) ([]models.Site, uint, error) {
	var total int64
	var sites []models.Site
	if err := Config.DB.Model(&models.Site{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	err := Config.DB.Limit(int(limit)).Offset(int(offset)).Order("id desc").Find(&sites).Error
	if err != nil {
		return nil, 0, err
	}
	return sites, uint(total), nil
}
func (r *SiteRepository) UpdateSite(site *models.Site, id uint) error {
	return r.DB.Save(site).Error
}
func (r *SiteRepository) DeleteSite(site *models.Site) error {
	return r.DB.Delete(site).Error
}
func (r *SiteRepository) CreateSiteHealthStatus(SiteHealthStatus *models.SiteHealthStatus) (*models.SiteHealthStatus, error) {
	err := r.DB.Create(SiteHealthStatus).Error
	if err != nil {
		return nil, err
	}
	return SiteHealthStatus, nil
}
