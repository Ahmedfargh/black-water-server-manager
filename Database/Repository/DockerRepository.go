package Repository

import (
	models "github.com/ahmedfargh/server-manager/Database/Models"
	"gorm.io/gorm"
)

type DockerRepository struct {
	DB *gorm.DB
}

func NewDockerRepository(db *gorm.DB) *DockerRepository {
	return &DockerRepository{DB: db}
}

func (r *DockerRepository) CreateDocker(docker *models.Docker) error {
	return r.DB.Create(docker).Error
}

func (r *DockerRepository) GetDockerByID(id uint) (*models.Docker, error) {
	var docker models.Docker
	err := r.DB.First(&docker, id).Error
	if err != nil {
		return nil, err
	}
	return &docker, nil
}

func (r *DockerRepository) GetDockerByContainerID(containerID string) (*models.Docker, error) {
	var docker models.Docker
	err := r.DB.Where("container_id = ?", containerID).First(&docker).Error
	if err != nil {
		return nil, err
	}
	return &docker, nil
}

func (r *DockerRepository) GetDockers(page uint, limit uint) ([]models.Docker, uint, error) {
	var total int64
	var dockers []models.Docker
	if err := r.DB.Model(&models.Docker{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	err := r.DB.Limit(int(limit)).Offset(int(offset)).Order("id desc").Find(&dockers).Error
	if err != nil {
		return nil, 0, err
	}
	return dockers, uint(total), nil
}

func (r *DockerRepository) UpdateDocker(docker *models.Docker) error {
	return r.DB.Save(docker).Error
}

func (r *DockerRepository) DeleteDocker(docker *models.Docker) error {
	return r.DB.Delete(docker).Error
}
