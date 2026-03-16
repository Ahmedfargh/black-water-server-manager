package HardWare

import (
	"github.com/ahmedfargh/server-manager/Processes"

	"github.com/shirou/gopsutil/v3/net"
)

type ConnectionWithProcess struct {
	net.ConnectionStat
	Process processes.ProcessInfo `json:"process"`
}

type NetworkService struct {
}

func NewNetworkService() *NetworkService {
	return &NetworkService{}
}

func (n *NetworkService) GetNetworkInfo() (map[string]interface{}, error) {
	//no grouped
	stat, err := net.IOCounters(false)
	if err != nil {
		return nil, err
	}
	//grouped
	statGrouped, err := net.IOCounters(true)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"network": stat, "networkGrouped": statGrouped}, nil
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
