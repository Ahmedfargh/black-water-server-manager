package crud

import (
	Config "github.com/ahmedfargh/server-manager/Config"
	models "github.com/ahmedfargh/server-manager/Database/Models"
	repository "github.com/ahmedfargh/server-manager/Database/Repository"
)

type SiteCrud struct {
	Rep *repository.SiteRepository
}

func NewSiteCrud(rep *repository.SiteRepository) *SiteCrud {
	return &SiteCrud{Rep: rep}
}
func (c *SiteCrud) CreateSite(site models.Site) error {
	return c.Rep.CreateSite(&site)
}
func (c *SiteCrud) GetSiteByID(id uint) (*models.Site, error) {
	return c.Rep.GetSiteByID(id)
}
func (c *SiteCrud) GetSites(page int, limit int) ([]models.Site, uint, error) {

	return repository.NewSiteRepository(Config.DB).GetSites(uint(page), uint(limit))

}

func (c *SiteCrud) AddAnalytics(site *models.SiteHealthStatus, id uint) (*models.SiteHealthStatus, error) {
	return c.Rep.CreateSiteHealthStatus(site)
}
