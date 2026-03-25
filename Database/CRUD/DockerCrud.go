package crud

import (
	Config "github.com/ahmedfargh/server-manager/Config"
	models "github.com/ahmedfargh/server-manager/Database/Models"
	repository "github.com/ahmedfargh/server-manager/Database/Repository"
)

type DockerCrud struct {
	Rep *repository.DockerRepository
}

func NewDockerCrud(rep *repository.DockerRepository) *DockerCrud {
	return &DockerCrud{Rep: rep}
}

func (c *DockerCrud) CreateDocker(docker models.Docker) error {
	return c.Rep.CreateDocker(&docker)
}

func (c *DockerCrud) GetDockerByID(id uint) (*models.Docker, error) {
	return c.Rep.GetDockerByID(id)
}

func (c *DockerCrud) GetDockers(page int, limit int) ([]models.Docker, uint, error) {
	return repository.NewDockerRepository(Config.DB).GetDockers(uint(page), uint(limit))
}

func (c *DockerCrud) UpdateDocker(docker *models.Docker) error {
	return c.Rep.UpdateDocker(docker)
}

func (c *DockerCrud) DeleteDocker(docker *models.Docker) error {
	return c.Rep.DeleteDocker(docker)
}

func (c *DockerCrud) GetDockerByContainerID(containerID string) (*models.Docker, error) {
	return c.Rep.GetDockerByContainerID(containerID)
}
