package reports

import (
	"time"

	fmt "fmt"

	config "github.com/ahmedfargh/server-manager/Config"
	model "github.com/ahmedfargh/server-manager/Database/Models"
	NotificationManager "github.com/ahmedfargh/server-manager/Managers"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type HardwareReport struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
}
type BackgroundHardwareReport struct {
	Report         HardwareReport `json:"report"`
	RunEachSeconds int            `json:"run_each_seconds"`
	timer          *time.Timer
}

func (b *BackgroundHardwareReport) Run() (int32, error) {
	fmt.Println("start hard ware reports")
	b.StartMonitoring()
	return 0, nil
}
func (b *BackgroundHardwareReport) HandleError(err error) error {
	return err
}
func (b *BackgroundHardwareReport) StartMonitoring() {
	b.timer = time.NewTimer(time.Duration(b.RunEachSeconds) * time.Second)
	for {
		select {
		case <-b.timer.C:
			cpuUsage, _ := cpu.Percent(0, false)
			memUsage, _ := mem.VirtualMemory()
			diskUsage, _ := disk.Usage("/")
			// Collect hardware report data
			report := HardwareReport{
				CPUUsage:    cpuUsage[0],
				MemoryUsage: memUsage.UsedPercent,
				DiskUsage:   diskUsage.UsedPercent,
			}
			b.Report = report
			b.SendReport()
		}
	}
}
func (b *BackgroundHardwareReport) StopMonitoring() {
	if b.timer != nil {
		b.timer.Stop()
	}
}
func (b *BackgroundHardwareReport) SendReport() {
	if b.timer != nil {
		b.timer.Reset(time.Duration(b.RunEachSeconds) * time.Second)
	}
	var users []model.User
	config.DB.Find(&users)
	message := "Hardware Report:\n Cpu Usage: %.2f%%\n Memory Usage: %.2f%%\n Disk Usage: %.2f%%\n"
	message = fmt.Sprintf(message, b.Report.CPUUsage, b.Report.MemoryUsage, b.Report.DiskUsage)
	NotificationManager := NotificationManager.NewNotificationManager(nil)
	NotificationManager.NotifyUsers(users, message, map[string]string{"ReportType": "Hardware"})
}
