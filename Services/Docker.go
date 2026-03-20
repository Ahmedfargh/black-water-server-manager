package Services

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	DockerClient "github.com/docker/docker/client"
)

type DockerContainer struct {
	ID      string           `json:"id"`
	Image   string           `json:"image"`
	ImageId string           `json:"imageId"`
	Command string           `json:"command"`
	Created int64            `json:"created"`
	Status  string           `json:"status"`
	Names   []string         `json:"names"`
	Ports   []container.Port `json:"ports"`
	State   string           `json:"state"`
}

type DockerService struct {
}

func NewDockerService() (*DockerService, error) {

	return &DockerService{}, nil
}

func (d *DockerService) GetContainers(ctx context.Context) ([]DockerContainer, error) {
	cli, err := DockerClient.NewClientWithOpts(
		DockerClient.FromEnv,
		DockerClient.WithAPIVersionNegotiation(),
	)
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	dockerContainers := make([]DockerContainer, 0, len(containers))
	for _, c := range containers {
		dockerContainers = append(dockerContainers, DockerContainer{
			ID:      c.ID,
			Image:   c.Image,
			ImageId: c.ImageID,
			Command: c.Command,
			Created: c.Created,
			Status:  c.Status,
			Names:   c.Names,
			Ports:   c.Ports,
			State:   c.State,
		})
	}
	return dockerContainers, nil
}

func (d *DockerService) GetContainerByID(ctx context.Context, id string) (*DockerContainer, error) {
	cli, err := DockerClient.NewClientWithOpts(
		DockerClient.FromEnv,
		DockerClient.WithAPIVersionNegotiation(),
	)
	containerJSON, err := cli.ContainerInspect(ctx, id)
	if err != nil {
		return nil, err
	}

	t, _ := time.Parse(time.RFC3339Nano, containerJSON.Created)
	return &DockerContainer{
		ID:      containerJSON.ID,
		Image:   containerJSON.Config.Image,
		ImageId: containerJSON.Image,
		Command: fmt.Sprintf("%v", containerJSON.Args),
		Created: t.Unix(),
		Status:  containerJSON.State.Status,
		Names:   []string{containerJSON.Name},
		State:   containerJSON.State.Status,
	}, nil
}
