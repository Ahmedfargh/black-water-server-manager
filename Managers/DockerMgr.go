package Managers

import (
	"context"
	"fmt"
	"sync"

	time "time"

	"encoding/json"

	Config "github.com/ahmedfargh/server-manager/Config"
	CRUD "github.com/ahmedfargh/server-manager/Database/CRUD"
	Models "github.com/ahmedfargh/server-manager/Database/Models"
	Repository "github.com/ahmedfargh/server-manager/Database/Repository"
	Docker "github.com/ahmedfargh/server-manager/Services"
	"github.com/docker/docker/api/types"
	DockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container" // Added for correct types
	DockerClient "github.com/docker/docker/client"
	Context "golang.org/x/net/context"
)

type DockerContainerVolums struct {
	Type        string
	Source      string
	Destination string
}
type DockerManager struct {
	mu                    sync.RWMutex
	Containers            map[string]DockerTypes.ContainerJSON
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
	dm.mu.Lock()
	defer dm.mu.Unlock()
	fmt.Println("##################DOCKER IMAGE#############################")
	containers, err := dm.Client.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(len(containers))
	for _, c := range containers {
		fmt.Println(c.ID)
		inspect, err := dm.Client.ContainerInspect(ctx, c.ID)
		if err != nil {
			fmt.Println(err)
			continue
		}

		dm.Containers[c.ID] = inspect
	}
	return nil
}
func (dm *DockerManager) UpdateDockerContainerStatus() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	docker_crud := CRUD.NewDockerCrud(Repository.NewDockerRepository(Config.DB))

	for _, container := range dm.Containers {
		inspect, err := dm.Client.ContainerInspect(Context.Background(), container.ID)
		if err != nil {
			fmt.Println(err)
			return err
		}
		dm.Containers[container.ID] = inspect
		cmd, er := json.Marshal(inspect.Config.Cmd)
		if er != nil {
			fmt.Println(er)
			cmd = []byte("uknown command")
		}
		ports, _ := json.Marshal(inspect.NetworkSettings.Ports)
		fmt.Println("docker image")
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
func (dm *DockerManager) Act(action string, docker *Models.Docker) (string, error) {
	if action == "restart" {
		err := dm.Client.ContainerRestart(Context.Background(), docker.ContainerID, container.StopOptions{})
		message := "Container restarted that has id " + docker.ContainerID + " That Has Name " + docker.Name
		go dm.NotifyDockerStatus(message, nil)
		if err != nil {
			return "", err
		}
	} else if action == "stop" {
		docker_service, err := Docker.NewDockerService()

		if err != nil {
			fmt.Println(err)
			return "", err
		}
		err_stop := docker_service.StopContainer(Context.Background(), docker.ContainerID)
		if err_stop != nil {
			fmt.Println(err_stop)
			return "", err
		}
		message := "Container stopped that has id " + docker.ContainerID + " That Has Name " + docker.Name
		go dm.NotifyDockerStatus(message, nil)
		return "", nil
	} else if action == "remove" {
		err := dm.Client.ContainerRemove(Context.Background(), docker.ContainerID, container.RemoveOptions{})
		if err != nil {
			return "", err
		}
		message := "Container removed that has id " + docker.ContainerID + " That Has Name " + docker.Name
		go dm.NotifyDockerStatus(message, nil)
		return "", nil
	} else if action == "start" {
		err := dm.Client.ContainerStart(Context.Background(), docker.ContainerID, container.StartOptions{})
		if err != nil {
			fmt.Println(err)
			return "", err
		}
		message := "Container started that has id " + docker.ContainerID + " That Has Name " + docker.Name
		go dm.NotifyDockerStatus(message, nil)
		return "", nil
	}
	return "", nil
}

func (dm *DockerManager) ActForAbnormality(docker *Models.Docker, docker_state *Docker.DockerContainerStats, abnormality string, action string) (string, error) {
	//check after 10 minutes
	// time.Sleep(10 * time.Minute)
	if abnormality == "max_cpu_consumation" {
		fmt.Println("checking")
		if docker_state.CPUPercentage > docker.MaxCpuConsumation {

			fmt.Println("test action:-" + action)
			dm.Act(action, docker)
		}
	}
	if abnormality == "max_memory_consumation" {
		if docker_state.MemoryPercentage > docker.MaxMemoryConsumation {
			return "", nil
		} else {
			dm.Act(action, docker)
		}
	}
	if abnormality == "stopped" {
		dm.Act(action, docker)
	}
	return "", nil
}
func (dm *DockerManager) StartMonitoring() {
	// 1. Create a ticker for 1 second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 2. Call the check function
			dm.CheckAll()
		}
	}
}

func (dm *DockerManager) CheckAll() {
	// Only lock when accessing or modifying the SHARED state (dm)
	// Don't lock during the DB find or Docker API calls if possible

	var containers []Models.Docker
	err := Config.DB.Find(&containers).Error
	if err != nil {
		return
	}

	doc, err := Docker.NewDockerService()
	if err != nil {
		return
	}

	for _, container := range containers {
		// Create a local copy to avoid pointer issues with goroutines
		c := container

		if c.MaxCpuConsumation > 0 {
			stats, err := doc.ContainerStatus(context.Background(), c.ContainerID)
			if err != nil {
				continue
			}

			// Logic Check: You likely want to act if stats > max, not stats <= max
			if stats.CPUPercentage >= c.MaxCpuConsumation {
				fmt.Printf("abnormality :\n max_cpu_consumation %f %s \n", c.MaxCpuConsumation, c.OnMaxCpuConsumation)
				go dm.ActForAbnormality(&c, &stats, "max_cpu_consumation", c.OnMaxCpuConsumation)
			}

		}

		if c.MaxMemoryConsumation > 0 {
			stat, err := doc.ContainerStatus(context.Background(), c.ContainerID)
			if err != nil {
				continue
			}
			if stat.MemoryPercentage >= c.MaxMemoryConsumation {
				fmt.Printf("abnormality :\n max_memory_consumation %f %s \n", c.MaxMemoryConsumation, c.OnMaxMemoryConsumation)
				go dm.ActForAbnormality(&c, &stat, "max_memory_consumation", c.OnMaxMemoryConsumation)
			}
		}
		if c.OnStopped != "nothing" {
			go dm.ActForAbnormality(&c, &Docker.DockerContainerStats{}, "stopped", c.OnStopped)
		}
	}
}
func (dm *DockerManager) NotifyDockerStatus(messsage string, metadata map[string]string) (any, error) {

	users := []Models.User{}
	Config.DB.Find(&users)
	fmt.Println("User Count:-", len(users))
	NotificationManager := NewNotificationManager(nil)
	NotificationManager.NotifyUsers(users, messsage, metadata)
	return "", nil
}

func (dm *DockerManager) DockerContainerVolumns(ctx context.Context, id string) ([]DockerContainerVolums, error) {

	docker_inspect, err := dm.Client.ContainerInspect(ctx, id)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var DockerContainerVolumns_slice []DockerContainerVolums
	volumns := docker_inspect.Mounts

	for i, volum := range volumns {
		volumn_obj := DockerContainerVolums{
			Type:        string(volum.Type),
			Source:      volum.Source,
			Destination: volum.Destination,
		}
		fmt.Println(i)
		DockerContainerVolumns_slice = append(DockerContainerVolumns_slice, volumn_obj)
	}
	return DockerContainerVolumns_slice, nil
}
func (dm *DockerManager) PruneDockerImage(ctx context.Context, id string) (bool, error) {
	option := container.RemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	}

	err := dm.Client.ContainerRemove(ctx, id, option)
	if err != nil {
		return false, err
	}

	repo := Repository.NewDockerRepository(Config.DB)
	repo.DeleteByContainerId(id)
	return true, nil
}
