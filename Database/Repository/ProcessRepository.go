package repository

import (
	models "github.com/ahmedfargh/server-manager/Database/Models"
	"gorm.io/gorm"
)

type ProcessRepository struct {
	DB *gorm.DB
}

func NewProcessRepository(db *gorm.DB) *ProcessRepository {
	return &ProcessRepository{DB: db}
}

func (r *ProcessRepository) CreateProcess(process *models.Process) error {
	return r.DB.Create(process).Error
}

func (r *ProcessRepository) GetProcessByPID(pid int32) (*models.Process, error) {
	var process models.Process
	err := r.DB.First(&process, "pid = ?", pid).Error
	if err != nil {
		return nil, err
	}
	return &process, nil
}

func (r *ProcessRepository) GetProcesses() ([]models.Process, error) {
	var processes []models.Process
	err := r.DB.Find(&processes).Error
	if err != nil {
		return nil, err
	}
	return processes, nil
}

func (r *ProcessRepository) GetPaginatedProcesses(page, pageSize int) ([]models.Process, int64, error) {
	var processes []models.Process
	var total int64

	// Get total count
	if err := r.DB.Model(&models.Process{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := r.DB.Limit(pageSize).Offset(offset).Order("id desc").Find(&processes).Error
	if err != nil {
		return nil, 0, err
	}
	return processes, total, nil
}
