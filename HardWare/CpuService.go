package HardWare

import (
	"fmt"
	"runtime"
	"strconv"

	"github.com/shirou/gopsutil/v3/cpu"
)

type CpuInfo struct {
	Logical_core       uint
	Max_Cpu_run        uint
	Arch               string
	Os                 string
	Cpu_Hard_Ware_Info map[int]map[string]string 
}

func GetCpuInfo() (CpuInfo, error) {
	info := CpuInfo{
		Logical_core:       uint(runtime.NumCPU()),
		Max_Cpu_run:        uint(runtime.GOMAXPROCS(0)),
		Arch:               runtime.GOARCH,
		Os:                 runtime.GOOS,
		Cpu_Hard_Ware_Info: make(map[int]map[string]string),
	}
	cpuInfo, err := cpu.Info()
	if err != nil {
		return info, fmt.Errorf("failed to get CPU info: %w", err)
	}

	for i, val := range cpuInfo {
		cpuHardwareInfo := make(map[string]string)
		cpuHardwareInfo["model"] = val.ModelName
		cpuHardwareInfo["vendor"] = val.VendorID
		cpuHardwareInfo["family"] = val.Family
		cpuHardwareInfo["model_name"] = val.ModelName
		cpuHardwareInfo["stepping"] = strconv.Itoa(int(val.Stepping))
		cpuHardwareInfo["physical_cores"] = strconv.Itoa(int(val.Cores))
		cpuHardwareInfo["mhz"] = fmt.Sprintf("%.2f", val.Mhz)
		cpuHardwareInfo["cache_size"] = strconv.Itoa(int(val.CacheSize))
		cpuHardwareInfo["microcode"] = val.Microcode
		if len(val.Flags) > 0 {
			flagsStr := ""
			for _, flag := range val.Flags {
				flagsStr += flag + " "
			}
			cpuHardwareInfo["flags"] = flagsStr
		}

		cpuHardwareInfo["cpu_index"] = strconv.Itoa(i)
		cpuHardwareInfo["cpu_id"] = strconv.Itoa(int(val.CPU))
		info.Cpu_Hard_Ware_Info[i] = cpuHardwareInfo
	}

	return info, nil
}
