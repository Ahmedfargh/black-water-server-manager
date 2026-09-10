package HardWare

import (
	"sync"
	"time"

	"github.com/ahmedfargh/server-manager/Processes"
	"github.com/shirou/gopsutil/v3/net"
)

type ConnectionWithProcess struct {
	net.ConnectionStat
	Process processes.ProcessInfo `json:"process"`
}

type NetworkService struct {
	mu            sync.RWMutex
	lastTime      time.Time
	lastBytesSent uint64
	lastBytesRecv uint64
	sentRate      float64
	recvRate      float64
	totalSent     uint64
	totalRecv     uint64
}

var globalNetworkService *NetworkService
var once sync.Once

func NewNetworkService() *NetworkService {
	once.Do(func() {
		globalNetworkService = &NetworkService{}
		if stat, err := net.IOCounters(false); err == nil && len(stat) > 0 {
			globalNetworkService.lastTime = time.Now()
			globalNetworkService.lastBytesSent = stat[0].BytesSent
			globalNetworkService.lastBytesRecv = stat[0].BytesRecv
			globalNetworkService.totalSent = stat[0].BytesSent
			globalNetworkService.totalRecv = stat[0].BytesRecv
		}
		go globalNetworkService.startSampler()
	})
	return globalNetworkService
}

func (n *NetworkService) startSampler() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stat, err := net.IOCounters(false)
		if err != nil || len(stat) == 0 {
			continue
		}

		now := time.Now()
		currSent := stat[0].BytesSent
		currRecv := stat[0].BytesRecv

		n.mu.Lock()
		if !n.lastTime.IsZero() {
			dt := now.Sub(n.lastTime).Seconds()
			if dt > 0.1 {
				if currSent >= n.lastBytesSent {
					n.sentRate = float64(currSent-n.lastBytesSent) / dt
				}
				if currRecv >= n.lastBytesRecv {
					n.recvRate = float64(currRecv-n.lastBytesRecv) / dt
				}
			}
		}
		n.lastTime = now
		n.lastBytesSent = currSent
		n.lastBytesRecv = currRecv
		n.totalSent = currSent
		n.totalRecv = currRecv
		n.mu.Unlock()
	}
}

func (n *NetworkService) GetNetworkInfo() (map[string]interface{}, error) {
	// no grouped
	stat, err := net.IOCounters(false)
	if err != nil {
		return nil, err
	}
	// grouped
	statGrouped, err := net.IOCounters(true)
	if err != nil {
		return nil, err
	}

	n.mu.RLock()
	sentRate := n.sentRate
	recvRate := n.recvRate
	totalSent := n.totalSent
	totalRecv := n.totalRecv
	n.mu.RUnlock()

	if len(stat) > 0 {
		totalSent = stat[0].BytesSent
		totalRecv = stat[0].BytesRecv
	}

	return map[string]interface{}{
		"network":        stat,
		"networkGrouped": statGrouped,
		"rates": map[string]interface{}{
			"sentPerSec": sentRate,
			"recvPerSec": recvRate,
			"totalSent":  totalSent,
			"totalRecv":  totalRecv,
		},
	}, nil
}
func (n *NetworkService) GetNetworkConnections() (map[string]interface{}, error) {
	connections, err := net.Connections("all")
	if err != nil {
		return nil, err
	}

	var results []ConnectionWithProcess
	for _, conn := range connections {
		item := ConnectionWithProcess{
			ConnectionStat: conn,
		}
		process, err := processes.GetProcessByPID(conn.Pid)
		if err == nil {
			item.Process = process
		}
		results = append(results, item)
	}
	return map[string]interface{}{"networkConnections": results}, nil
}
