package HardWare

import (
	"github.com/shirou/gopsutil/v3/net"
)

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
