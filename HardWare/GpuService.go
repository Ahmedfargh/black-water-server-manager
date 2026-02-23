package HardWare

import (
	"os/exec"
	"runtime"
	"strings"
)

type GpuInfo struct {
	Logical_core       uint
	Max_Cpu_run        uint
	Arch               string
	Os                 string
	Gpu_Hard_Ware_Info map[int]map[string]string
}

func GetGpuInfo() (GpuInfo, error) {
	info := GpuInfo{
		Logical_core:       uint(runtime.NumCPU()),
		Max_Cpu_run:        uint(runtime.GOMAXPROCS(0)),
		Arch:               runtime.GOARCH,
		Os:                 runtime.GOOS,
		Gpu_Hard_Ware_Info: make(map[int]map[string]string),
	}

	switch runtime.GOOS {
	case "linux":
		gpuInfo := detectLinuxGPU()
		if len(gpuInfo) > 0 {
			info.Gpu_Hard_Ware_Info = gpuInfo
		}
	case "windows":
		gpuInfo := detectWindowsGPU()
		if len(gpuInfo) > 0 {
			info.Gpu_Hard_Ware_Info = gpuInfo
		}
	case "darwin":
		gpuInfo := detectMacGPU()
		if len(gpuInfo) > 0 {
			info.Gpu_Hard_Ware_Info = gpuInfo
		}
	}

	return info, nil
}
func detectLinuxGPU() map[int]map[string]string {
	gpuInfo := make(map[int]map[string]string)
	out, err := exec.Command("sh", "-c", "lspci -nn | grep -E -i 'vga|3d|display'").Output()
	if err != nil {
		return gpuInfo
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for i, line := range lines {
		gpuInfo[i] = map[string]string{
			"Raw":   line,
			"Brand": identifyBrand(line),
		}
	}
	return gpuInfo
}

func identifyBrand(line string) string {
	line = strings.ToLower(line)
	if strings.Contains(line, "nvidia") {
		return "NVIDIA"
	}
	if strings.Contains(line, "intel") {
		return "Intel"
	}
	if strings.Contains(line, "amd") || strings.Contains(line, "ati") {
		return "AMD"
	}
	return "Unknown"
}
func detectWindowsGPU() map[int]map[string]string {
	gpuInfo := make(map[int]map[string]string)
	// Querying Name and AdapterRAM
	out, err := exec.Command("wmic", "path", "win32_VideoController", "get", "name").Output()
	if err == nil {
		lines := strings.Split(string(out), "\r\n")
		index := 0
		for _, line := range lines[1:] { // Skip header
			name := strings.TrimSpace(line)
			if name != "" {
				gpuInfo[index] = map[string]string{"Name": name}
				index++
			}
		}
	}
	return gpuInfo
}
func detectMacGPU() map[int]map[string]string {
	gpuInfo := make(map[int]map[string]string)
	out, err := exec.Command("system_profiler", "SPDisplaysDataType").Output()
	if err == nil {
		// Simple parsing: looking for "Chipset Model"
		lines := strings.Split(string(out), "\n")
		index := 0
		for _, line := range lines {
			if strings.Contains(line, "Chipset Model") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					gpuInfo[index] = map[string]string{"Model": strings.TrimSpace(parts[1])}
					index++
				}
			}
		}
	}
	return gpuInfo
}
