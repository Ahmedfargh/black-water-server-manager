package crud

import (
	"time"

	models "github.com/ahmedfargh/server-manager/Database/Models"
	Repository "github.com/ahmedfargh/server-manager/Database/Repository"
	"gorm.io/gorm"
)

type HardWareReportCRUD struct {
	DB   *gorm.DB
	repo *Repository.HardWareReportRepository
}

func NewHardWareReportCRUD(db *gorm.DB) *HardWareReportCRUD {
	return &HardWareReportCRUD{DB: db, repo: Repository.NewHardWareReportRepository(db)}
}

func (c *HardWareReportCRUD) Create(report *models.HardWareReport) error {
	return c.repo.Create(report)
}

func (c *HardWareReportCRUD) GetAll(page int, limit int) ([]models.HardWareReport, error) {
	return c.repo.GetAll(page, limit)
}

func (c *HardWareReportCRUD) GetByID(id uint) (*models.HardWareReport, error) {
	return c.repo.GetByID(id)
}
func (c *HardWareReportCRUD) Update(report *models.HardWareReport) error {
	return c.repo.Update(report)
}

func (c *HardWareReportCRUD) Delete(id uint) error {
	return c.repo.Delete(id)
}

func (c *HardWareReportCRUD) GetLatest() (*models.HardWareReport, error) {
	return c.repo.GetLatest()
}

func (c *HardWareReportCRUD) GetReportsByTimeRange(start, end time.Time) ([]models.HardWareReport, error) {
	return c.repo.GetReportsByTimeRange(start, end)
}

func (c *HardWareReportCRUD) GetAverageUsageByTimeRange(start, end time.Time) (float64, float64, float64, error) {
	return c.repo.GetAverageUsageByTimeRange(start, end)
}
