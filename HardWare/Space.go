package HardWare

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
)

type DiskInfo struct {
	Path        string `json:"Path"`
	TotalGB     int32  `json:"TotalGB"`
	UsedGB      int32  `json:"UsedGB"`
	FreeGB      int32  `json:"FreeGB"`
	UsedPercent int32  `json:"UsedPercent"`
	FSType      string `json:"FSType"`
}

type PartitionDetail struct {
	Name        string  `json:"name"`
	Device      string  `json:"device"`
	Mountpoint  string  `json:"mountpoint"`
	FSType      string  `json:"fstype"`
	TotalBytes  uint64  `json:"total_bytes"`
	UsedBytes   uint64  `json:"used_bytes"`
	FreeBytes   uint64  `json:"free_bytes"`
	UsedPercent float64 `json:"used_percent"`
	IsMounted   bool    `json:"is_mounted"`
	Size        uint64  `json:"size"`
}

type PhysicalDisk struct {
	Name         string            `json:"name"`
	Device       string            `json:"device"`
	Model        string            `json:"model"`
	Type         string            `json:"type"` // NVMe, SSD, HDD
	Size         uint64            `json:"size"`
	State        string            `json:"state"`
	IsRotational bool              `json:"is_rotational"`
	IsRemovable  bool              `json:"is_removable"`
	Temperature  *float64          `json:"temperature"` // Celsius, nil if unavailable
	Partitions   []PartitionDetail `json:"partitions"`
}

type DiskUsage struct {
	Disks       []DiskInfo     `json:"Disks"`
	Devices     []PhysicalDisk `json:"devices"`
	TotalBytes  uint64         `json:"total_bytes"`
	UsedBytes   uint64         `json:"used_bytes"`
	FreeBytes   uint64         `json:"free_bytes"`
	UsedPercent float64        `json:"used_percent"`
}

type lsblkChild struct {
	Name       string  `json:"name"`
	Path       string  `json:"path"`
	Type       string  `json:"type"`
	Size       uint64  `json:"size"`
	Mountpoint *string `json:"mountpoint"`
	FSType     *string `json:"fstype"`
	Model      *string `json:"model"`
	Rota       bool    `json:"rota"`
	Rm         bool    `json:"rm"`
	State      *string `json:"state"`
}

type lsblkDevice struct {
	Name       string       `json:"name"`
	Path       string       `json:"path"`
	Type       string       `json:"type"`
	Size       uint64       `json:"size"`
	Mountpoint *string      `json:"mountpoint"`
	FSType     *string      `json:"fstype"`
	Model      *string      `json:"model"`
	Rota       bool         `json:"rota"`
	Rm         bool         `json:"rm"`
	State      *string      `json:"state"`
	Children   []lsblkChild `json:"children"`
}

type lsblkRoot struct {
	BlockDevices []lsblkDevice `json:"blockdevices"`
}

func getDiskTemperatures() map[string]float64 {
	temps := make(map[string]float64)

	// 1. Check gopsutil host sensors
	if sensorTemps, err := host.SensorsTemperatures(); err == nil {
		for _, st := range sensorTemps {
			lower := strings.ToLower(st.SensorKey)
			if strings.Contains(lower, "nvme") && (strings.Contains(lower, "composite") || strings.Contains(lower, "sensor_1")) {
				temps["nvme"] = st.Temperature
			}
		}
	}

	// 2. Scan /sys/class/hwmon for disk and nvme temperatures
	matches, _ := filepath.Glob("/sys/class/hwmon/hwmon*")
	for _, h := range matches {
		nameBytes, err := os.ReadFile(filepath.Join(h, "name"))
		if err != nil {
			continue
		}
		hwName := strings.TrimSpace(string(nameBytes))

		devLink, _ := filepath.EvalSymlinks(filepath.Join(h, "device"))

		tempFiles, _ := filepath.Glob(filepath.Join(h, "temp*_input"))
		for _, tFile := range tempFiles {
			tBytes, err := os.ReadFile(tFile)
			if err != nil {
				continue
			}
			milli, err := strconv.ParseFloat(strings.TrimSpace(string(tBytes)), 64)
			if err != nil {
				continue
			}
			degC := milli / 1000.0

			if strings.Contains(hwName, "nvme") {
				temps["nvme"] = degC
				parts := strings.Split(devLink, "/")
				for _, p := range parts {
					if strings.HasPrefix(p, "nvme") {
						temps[p] = degC
					}
				}
			}

			if strings.Contains(hwName, "drivetemp") {
				parts := strings.Split(devLink, "/")
				for _, p := range parts {
					if strings.HasPrefix(p, "sd") || strings.HasPrefix(p, "hd") {
						temps[p] = degC
					}
				}
			}
		}
	}

	return temps
}

func isVirtualFS(fstype string) bool {
	lower := strings.ToLower(fstype)
	switch lower {
	case "tmpfs", "devtmpfs", "squashfs", "overlay", "nsfs", "proc", "sysfs",
		"efivarfs", "securityfs", "devpts", "cgroup2", "pstore", "bpf",
		"autofs", "mqueue", "debugfs", "tracefs", "hugetlbfs", "configfs", "fusectl":
		return true
	}
	return false
}

func GetDiskInfo() (DiskUsage, error) {
	var usage DiskUsage
	temps := getDiskTemperatures()

	out, err := exec.Command("lsblk", "-J", "-b", "-o", "NAME,PATH,TYPE,SIZE,MOUNTPOINT,FSTYPE,MODEL,ROTA,RM,STATE").Output()
	if err == nil {
		var parsed lsblkRoot
		if err := json.Unmarshal(out, &parsed); err == nil {
			var totalMounted, usedMounted, freeMounted uint64

			for _, dev := range parsed.BlockDevices {
				// Skip pseudo/loop/zram disks from physical listing
				if dev.Type != "disk" || strings.HasPrefix(dev.Name, "loop") || strings.HasPrefix(dev.Name, "zram") {
					continue
				}

				diskType := "SSD"
				if strings.HasPrefix(dev.Name, "nvme") {
					diskType = "NVMe"
				} else if dev.Rota {
					diskType = "HDD"
				}

				model := "Disk Storage"
				if dev.Model != nil && *dev.Model != "" {
					model = strings.TrimSpace(*dev.Model)
				}

				state := "connected"
				if dev.State != nil && *dev.State != "" {
					state = strings.TrimSpace(*dev.State)
				}

				var diskTemp *float64
				if t, ok := temps[dev.Name]; ok {
					diskTemp = &t
				} else if strings.HasPrefix(dev.Name, "nvme") {
					for k, v := range temps {
						if strings.HasPrefix(dev.Name, k) || strings.HasPrefix(k, "nvme") {
							val := v
							diskTemp = &val
							break
						}
					}
				}

				pDisk := PhysicalDisk{
					Name:         dev.Name,
					Device:       dev.Path,
					Model:        model,
					Type:         diskType,
					Size:         dev.Size,
					State:        state,
					IsRotational: dev.Rota,
					IsRemovable:  dev.Rm,
					Temperature:  diskTemp,
					Partitions:   []PartitionDetail{},
				}

				// If the disk has no children (e.g. single raw unpartitioned filesystem), check disk itself
				if len(dev.Children) == 0 {
					mount := ""
					isMounted := false
					if dev.Mountpoint != nil && *dev.Mountpoint != "" {
						mount = *dev.Mountpoint
						isMounted = true
					}
					fstype := "raw"
					if dev.FSType != nil && *dev.FSType != "" {
						fstype = *dev.FSType
					}

					var totalBytes, usedBytes, freeBytes uint64
					var usedPercent float64

					if isMounted {
						if u, err := disk.Usage(mount); err == nil {
							totalBytes = u.Total
							usedBytes = u.Used
							freeBytes = u.Free
							usedPercent = u.UsedPercent
							totalMounted += totalBytes
							usedMounted += usedBytes
							freeMounted += freeBytes

							usage.Disks = append(usage.Disks, DiskInfo{
								Path:        mount,
								TotalGB:     int32(u.Total / 1024 / 1024 / 1024),
								UsedGB:      int32(u.Used / 1024 / 1024 / 1024),
								FreeGB:      int32(u.Free / 1024 / 1024 / 1024),
								UsedPercent: int32(u.UsedPercent),
								FSType:      fstype,
							})
						}
					} else {
						totalBytes = dev.Size
					}

					pDisk.Partitions = append(pDisk.Partitions, PartitionDetail{
						Name:        dev.Name,
						Device:      dev.Path,
						Mountpoint:  mount,
						FSType:      fstype,
						TotalBytes:  totalBytes,
						UsedBytes:   usedBytes,
						FreeBytes:   freeBytes,
						UsedPercent: usedPercent,
						IsMounted:   isMounted,
						Size:        dev.Size,
					})
				}

				for _, child := range dev.Children {
					mount := ""
					isMounted := false
					if child.Mountpoint != nil && *child.Mountpoint != "" {
						mount = *child.Mountpoint
						isMounted = true
					}

					fstype := "raw"
					if child.FSType != nil && *child.FSType != "" {
						fstype = *child.FSType
					}

					var totalBytes, usedBytes, freeBytes uint64
					var usedPercent float64

					if isMounted {
						if u, err := disk.Usage(mount); err == nil {
							totalBytes = u.Total
							usedBytes = u.Used
							freeBytes = u.Free
							usedPercent = u.UsedPercent
							totalMounted += totalBytes
							usedMounted += usedBytes
							freeMounted += freeBytes

							usage.Disks = append(usage.Disks, DiskInfo{
								Path:        mount,
								TotalGB:     int32(u.Total / 1024 / 1024 / 1024),
								UsedGB:      int32(u.Used / 1024 / 1024 / 1024),
								FreeGB:      int32(u.Free / 1024 / 1024 / 1024),
								UsedPercent: int32(u.UsedPercent),
								FSType:      fstype,
							})
						} else {
							totalBytes = child.Size
						}
					} else {
						totalBytes = child.Size
					}

					pDisk.Partitions = append(pDisk.Partitions, PartitionDetail{
						Name:        child.Name,
						Device:      child.Path,
						Mountpoint:  mount,
						FSType:      fstype,
						TotalBytes:  totalBytes,
						UsedBytes:   usedBytes,
						FreeBytes:   freeBytes,
						UsedPercent: usedPercent,
						IsMounted:   isMounted,
						Size:        child.Size,
					})
				}

				usage.Devices = append(usage.Devices, pDisk)
			}

			usage.TotalBytes = totalMounted
			usage.UsedBytes = usedMounted
			usage.FreeBytes = freeMounted
			if totalMounted > 0 {
				usage.UsedPercent = float64(usedMounted) / float64(totalMounted) * 100.0
			}

			return usage, nil
		}
	}

	// Fallback using gopsutil partitions with virtual FS filtering
	partitions, err := disk.Partitions(true)
	if err != nil {
		return usage, err
	}

	var totalMounted, usedMounted, freeMounted uint64
	for _, partition := range partitions {
		if isVirtualFS(partition.Fstype) {
			continue
		}

		usageStat, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue
		}

		totalMounted += usageStat.Total
		usedMounted += usageStat.Used
		freeMounted += usageStat.Free

		usage.Disks = append(usage.Disks, DiskInfo{
			Path:        partition.Mountpoint,
			TotalGB:     int32(usageStat.Total / 1024 / 1024 / 1024),
			UsedGB:      int32(usageStat.Used / 1024 / 1024 / 1024),
			FreeGB:      int32(usageStat.Free / 1024 / 1024 / 1024),
			UsedPercent: int32(usageStat.UsedPercent),
			FSType:      partition.Fstype,
		})
	}

	usage.TotalBytes = totalMounted
	usage.UsedBytes = usedMounted
	usage.FreeBytes = freeMounted
	if totalMounted > 0 {
		usage.UsedPercent = float64(usedMounted) / float64(totalMounted) * 100.0
	}

	return usage, nil
}
