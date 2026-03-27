package Services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	Config "github.com/ahmedfargh/server-manager/Config"
	CRUD "github.com/ahmedfargh/server-manager/Database/CRUD"
	Repository "github.com/ahmedfargh/server-manager/Database/Repository"
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
type DockerContainerStats struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	CPUPercentage    float64 `json:"cpu_percentage"`
	MemoryUsage      uint64  `json:"memory_usage"`
	MemoryLimit      uint64  `json:"memory_limit"`
	MemoryPercentage float64 `json:"memory_percentage"`
	NetworkRX        uint64  `json:"network_rx"`
	NetworkTX        uint64  `json:"network_tx"`
	BlockRead        uint64  `json:"block_read"`
	BlockWrite       uint64  `json:"block_write"`
	Pids             uint64  `json:"pids"`
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
	if err != nil {
		return nil, err
	}
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
	if err != nil {
		return nil, err
	}
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

func (d *DockerService) loadContainerStatus(v *container.Stats, id string, name string) (DockerContainerStats, error) {
	var (
		cpuPercent = 0.0
		blkRead    = uint64(0)
		blkWrite   = uint64(0)
		mem        = uint64(0)
		memLimit   = uint64(0)
		memPercent = 0.0
		netRx      = uint64(0)
		netTx      = uint64(0)
	)

	cpuDelta := float64(v.CPUStats.CPUUsage.TotalUsage) - float64(v.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(v.CPUStats.SystemUsage) - float64(v.PreCPUStats.SystemUsage)
	onlineCPUs := float64(v.CPUStats.OnlineCPUs)
	if onlineCPUs == 0.0 {
		onlineCPUs = float64(len(v.CPUStats.CPUUsage.PercpuUsage))
	}
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * onlineCPUs * 100.0
	}

	if v.MemoryStats.Stats["total_inactive_file"] != 0 {
		mem = v.MemoryStats.Usage - uint64(v.MemoryStats.Stats["total_inactive_file"])
	} else {
		mem = v.MemoryStats.Usage - uint64(v.MemoryStats.Stats["inactive_file"])
	}
	memLimit = v.MemoryStats.Limit
	if memLimit != 0 {
		memPercent = float64(mem) / float64(memLimit) * 100.0
	}

	for _, n := range v.Networks {
		netRx += n.RxBytes
		netTx += n.TxBytes
	}

	// Block I/O Calculation
	for _, bioEntry := range v.BlkioStats.IoServiceBytesRecursive {
		switch bioEntry.Op {
		case "Read":
			blkRead += bioEntry.Value
		case "Write":
			blkWrite += bioEntry.Value
		}
	}

	return DockerContainerStats{
		ID:               id,
		Name:             name,
		CPUPercentage:    cpuPercent,
		MemoryUsage:      mem,
		MemoryLimit:      memLimit,
		MemoryPercentage: memPercent,
		NetworkRX:        netRx,
		NetworkTX:        netTx,
		BlockRead:        blkRead,
		BlockWrite:       blkWrite,
		Pids:             v.PidsStats.Current,
	}, nil
}

func (d *DockerService) ContainerStatus(ctx context.Context, id string) (DockerContainerStats, error) {
	cli, err := DockerClient.NewClientWithOpts(
		DockerClient.FromEnv,
		DockerClient.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return DockerContainerStats{}, err
	}
	stats, err := cli.ContainerStats(ctx, id, false)
	if err != nil {
		return DockerContainerStats{}, err
	}
	defer stats.Body.Close()

	var v struct {
		container.Stats
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(stats.Body).Decode(&v); err != nil {
		return DockerContainerStats{}, err
	}

	return d.loadContainerStatus(&v.Stats, v.ID, v.Name)
}
func (d *DockerService) ContainerLogs(ctx context.Context, id string) (io.ReadCloser, error) {
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: true,
	}
	cli, err := DockerClient.NewClientWithOpts(
		DockerClient.FromEnv,
		DockerClient.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}
	logs, err := cli.ContainerLogs(ctx, id, options)
	if err != nil {
		return nil, err
	}
	return logs, nil
}
func (d *DockerService) StartContainer(ctx context.Context, id string) error {
	cli, err := DockerClient.NewClientWithOpts(
		DockerClient.FromEnv,
		DockerClient.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return err
	}
	err = cli.ContainerStart(ctx, id, container.StartOptions{})
	if err != nil {
		return err
	}
	return nil
}
func (d *DockerService) StopContainer(ctx context.Context, id string) error {
	cli, err := DockerClient.NewClientWithOpts(
		DockerClient.FromEnv,
		DockerClient.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return err
	}
	err = cli.ContainerStop(ctx, id, container.StopOptions{})
	if err != nil {
		return err
	}
	return nil
}
func (d *DockerService) RestartContainer(ctx context.Context, id string) error {
	cli, err := DockerClient.NewClientWithOpts(
		DockerClient.FromEnv,
		DockerClient.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return err
	}
	err = cli.ContainerRestart(ctx, id, container.StopOptions{})
	if err != nil {
		return err
	}
	return nil
}
func (d *DockerService) AddEventAction(ctx context.Context, id string, event string, action string, value float64) (bool, error) {
	docker_crud_service := CRUD.NewDockerCrud(Repository.NewDockerRepository(Config.DB))
	docker, err := docker_crud_service.GetDockerByContainerID(id)
	if err != nil {
		return false, err
	}
	return docker_crud_service.AddEventAction(docker, event, action, value)
}
