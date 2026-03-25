package Models

import "gorm.io/gorm"

// Define the custom type
type DockerAllowedAction string

// Correct constant declaration
const (
	Restart = "restart"
	Start   = "start"
	Stop    = "stop"
	Remove  = "remove"
	NoThing = "nothing"
)

type Docker struct {
	gorm.Model
	ContainerID            string              `gorm:"unique;not null" json:"container_id"`
	Name                   string              `gorm:"not null" json:"name"`
	Image                  string              `gorm:"not null" json:"image"`
	Status                 string              `gorm:"not null" json:"status"`
	Command                string              `gorm:"not null" json:"command"`
	Created                string              `gorm:"not null" json:"created"`
	Ports                  string              `gorm:"not null" json:"ports"`
	MaxCpuConsumation      float32             `gorm:"not null;check:max_cpu_consumation >= 0 AND max_cpu_consumation <= 100" json:"max_cpu_consumation"`
	MaxMemoryConsumation   float32             `gorm:"not null;check:max_memory_consumation >= 0 AND max_memory_consumation <= 100" json:"max_memory_consumation"`
	OnMaxCpuConsumation    DockerAllowedAction `gorm:"type:varchar(20);default:'nothing'" json:"on_max_cpu_consumation"`
	OnMaxMemoryConsumation DockerAllowedAction `gorm:"type:varchar(20);default:'nothing'" json:"on_max_memory_consumation"`
	OnStopped              DockerAllowedAction `gorm:"type:varchar(20);default:'nothing'" json:"on_stopped"`
}

func (Docker) TableName() string {
	return "dockers"
}
