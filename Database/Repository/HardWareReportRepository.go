package Repository

import (
	"time"

	models "github.com/ahmedfargh/server-manager/Database/Models"
	"gorm.io/gorm"
)

type HardWareReportRepository struct {
	DB *gorm.DB
}

func NewHardWareReportRepository(db *gorm.DB) *HardWareReportRepository {
	return &HardWareReportRepository{DB: db}
}

func (r *HardWareReportRepository) Create(report *models.HardWareReport) error {
	return r.DB.Create(report).Error
}

func (r *HardWareReportRepository) GetAll(page int, limit int) ([]models.HardWareReport, error) {
	var reports []models.HardWareReport
	err := r.DB.Offset((page - 1) * limit).Limit(limit).Find(&reports).Error
	return reports, err
}

func (r *HardWareReportRepository) GetByID(id uint) (*models.HardWareReport, error) {
	var report models.HardWareReport
	err := r.DB.First(&report, id).Error
	return &report, err
}

func (r *HardWareReportRepository) Update(report *models.HardWareReport) error {
	return r.DB.Save(report).Error
}

func (r *HardWareReportRepository) Delete(id uint) error {
	return r.DB.Delete(&models.HardWareReport{}, id).Error
}

func (r *HardWareReportRepository) GetLatest() (*models.HardWareReport, error) {
	var report models.HardWareReport
	err := r.DB.Order("created_at desc").First(&report).Error
	return &report, err
}
func (r *HardWareReportRepository) GetReportsByTimeRange(start, end time.Time) ([]models.HardWareReport, error) {
	var reports []models.HardWareReport
	err := r.DB.Where("created_at BETWEEN ? AND ?", start, end).Find(&reports).Error
	return reports, err
}
func (r *HardWareReportRepository) GetAverageUsageByTimeRange(start, end time.Time) (float64, float64, float64, error) {
	var result struct {
		AvgCPU    *float64 `gorm:"column:avg_cpu"`
		AvgMemory *float64 `gorm:"column:avg_memory"`
		AvgDisk   *float64 `gorm:"column:avg_disk"`
	}

	err := r.DB.Model(&models.HardWareReport{}).
		Select("AVG(cpu_usage) as avg_cpu, AVG(memory_usage) as avg_memory, AVG(disk_usage) as avg_disk").
		Where("created_at BETWEEN ? AND ?", start, end).
		Scan(&result).Error

	if err != nil {
		return 0, 0, 0, err
	}

	var cpu, mem, disk float64
	if result.AvgCPU != nil {
		cpu = *result.AvgCPU
	}
	if result.AvgMemory != nil {
		mem = *result.AvgMemory
	}
	if result.AvgDisk != nil {
		disk = *result.AvgDisk
	}

	return cpu, mem, disk, nil
}
