package Models

import (
	"gorm.io/gorm"
)

type HardWareReport struct {
	gorm.Model
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
}
