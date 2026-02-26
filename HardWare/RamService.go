package HardWare

import (
	"github.com/shirou/gopsutil/v3/mem"
)

type VertiualInfo struct {
	Total_memory int32
	Used_memory  int32
	Free_memory  int32
	Used_percent int32
}
type SwapInfo struct {
	Total_memory int32
	Used_memory  int32
	Free_memory  int32
	Used_percent int32
}
type RamInfo struct {
	Vertiual_info VertiualInfo
	SwapInfo      SwapInfo
}

func GetRamInfo() (RamInfo, error) {
	info := RamInfo{}
	//load Virtual Memeory
	v, _ := mem.VirtualMemory()
	info.Vertiual_info.Total_memory = int32(v.Total / 1024 / 1024)
	info.Vertiual_info.Used_memory = int32(v.Used / 1024 / 1024)
	info.Vertiual_info.Free_memory = int32(v.Free / 1024 / 1024)
	info.Vertiual_info.Used_percent = int32(v.UsedPercent)
	//load Swap Memory
	s, _ := mem.SwapMemory()
	info.SwapInfo.Total_memory = int32(s.Total / 1024 / 1024)
	info.SwapInfo.Used_memory = int32(s.Used / 1024 / 1024)
	info.SwapInfo.Free_memory = int32(s.Free / 1024 / 1024)
	info.SwapInfo.Used_percent = int32(s.UsedPercent)
	return info, nil
}
