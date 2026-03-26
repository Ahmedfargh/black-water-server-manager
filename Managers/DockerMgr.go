package Managers

import (
	"context"
	"sync"

	time "time"

	"encoding/json"

	Config "github.com/ahmedfargh/server-manager/Config"
	CRUD "github.com/ahmedfargh/server-manager/Database/CRUD"
	Models "github.com/ahmedfargh/server-manager/Database/Models"
	Repository "github.com/ahmedfargh/server-manager/Database/Repository"
	Docker "github.com/ahmedfargh/server-manager/Services"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container" // Added for correct types
	DockerClient "github.com/docker/docker/client"
	Context "golang.org/x/net/context"
)

type DockerManager struct {
	mu                    sync.RWMutex
	Containers            map[string]types.ContainerJSON
	Client                *DockerClient.Client
	DockerContainerStates map[string]Docker.DockerContainerStats
}

var (
	instance *DockerManager
	once     sync.Once
)

func GetDockerManager() *DockerManager {
	once.Do(func() {
		cli, _ := DockerClient.NewClientWithOpts(DockerClient.FromEnv, DockerClient.WithAPIVersionNegotiation())

		instance = &DockerManager{
			Containers: make(map[string]types.ContainerJSON),
			Client:     cli,
		}
		instance.DiscoverContainers(context.Background())
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					instance.UpdateDockerContainerStatus()
				}
			}
		}()
	})
	return instance
}

func (dm *DockerManager) DiscoverContainers(ctx context.Context) error {
	containers, err := dm.Client.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return err
	}

	dm.mu.Lock()
	defer dm.mu.Unlock()

	for _, c := range containers {
		inspect, err := dm.Client.ContainerInspect(ctx, c.ID)
		if err != nil {
			continue
		}
		dm.Containers[c.ID] = inspect
	}
	return nil
}
func (dm *DockerManager) UpdateDockerContainerStatus() error {
	docker_crud := CRUD.NewDockerCrud(Repository.NewDockerRepository(Config.DB))

	dm.mu.Lock()
	defer dm.mu.Unlock()
	for _, container := range dm.Containers {
		inspect, err := dm.Client.ContainerInspect(Context.Background(), container.ID)
		if err != nil {
			return err
		}
		dm.Containers[container.ID] = inspect
		cmd, er := json.Marshal(inspect.Config.Cmd)
		if er != nil {
			cmd = []byte("uknown command")
		}
		ports, _ := json.Marshal(inspect.NetworkSettings.Ports)

		docker_model := Models.Docker{
			ContainerID: inspect.ID,
			Name:        inspect.Name,
			Image:       inspect.Config.Image,
			Status:      inspect.State.Status,
			Command:     string(cmd),
			Created:     inspect.Created,
			Ports:       string(ports),
		}
		docker_crud.Rep.CreateDocker(&docker_model)
		return nil
	}
	return nil
}
func (dm *DockerManager) GetDockerState(containerID string) (Docker.DockerContainerStats, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if state, ok := dm.DockerContainerStates[containerID]; ok {
		return state, nil
	}
	return Docker.DockerContainerStats{}, nil
}
