package HardWare

import (
	"github.com/shirou/gopsutil/v3/disk"
)

type DiskInfo struct {
	Path        string
	TotalGB     int32
	UsedGB      int32
	FreeGB      int32
	UsedPercent int32
	FSType      string
}
type DiskUsage struct {
	Disks []DiskInfo
}

func GetDiskInfo() (DiskUsage, error) {
	var usage DiskUsage
	partitions, err := disk.Partitions(true)
	if err != nil {
		return usage, err
	}

	for _, partition := range partitions {
		usageStat, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			return usage, err
		}
		usage.Disks = append(usage.Disks, DiskInfo{
			Path:        partition.Mountpoint,
			TotalGB:     int32(usageStat.Total / 1024 / 1024 / 1024),
			UsedGB:      int32(usageStat.Used / 1024 / 1024 / 1024),
			FreeGB:      int32(usageStat.Free / 1024 / 1024 / 1024),
			UsedPercent: int32(usageStat.UsedPercent),
			FSType:      partition.Fstype,
		})
	}

	return usage, nil

}
