package Repository

import (
	"errors"
	"fmt"
	"time"

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
func (r *SiteRepository) GetSiteHealthStatus(site_id uint, page int, limit int) ([]models.SiteHealthStatus, uint, error) {
	var SiteHealthStatus []models.SiteHealthStatus
	var total int64
	if err := Config.DB.Model(&models.Site{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	err := Config.DB.Limit(int(limit)).Offset(int(offset)).Order("id desc").Find(&SiteHealthStatus).Error
	if err != nil {
		return nil, 0, err
	}
	return SiteHealthStatus, uint(total), nil
}
func (r *SiteRepository) GetSiteHealthStatusByDate(site_id uint, start_date string, end_date string) ([]models.SiteHealthStatus, error) {
	var SiteHealthStatus []models.SiteHealthStatus
	start, errStart := time.Parse("2006-1-2", start_date)
	end, errEnd := time.Parse("2006-1-2", end_date)
	fmt.Printf("start:-%s\n end:%s\n", start, end)
	if errStart != nil || errEnd != nil {
		return nil, errors.New("invalid date format, use YYYY-MM-DD")
	}

	end = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	err := Config.DB.Where("site_id = ? AND time BETWEEN ? AND ?", site_id, start, end).
		Order("time ASC").
		Find(&SiteHealthStatus).Error
	if err != nil {
		return nil, err
	}
	return SiteHealthStatus, nil
}
