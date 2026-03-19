package Services

import (
	"context"

	"github.com/docker/docker/api/types/container"
	DockerClient "github.com/docker/docker/client"
)

type DockerService struct {
}

func NewDockerService() *DockerService {
	return &DockerService{}
}
func (d *DockerService) GetContainers() (interface{}, error) {
	ctx := context.Background()
	cli, err := DockerClient.NewClientWithOpts(DockerClient.FromEnv, DockerClient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	containers, err := cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return nil, err
	}
	return containers, nil
}
