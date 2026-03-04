package models

import "gorm.io/gorm"

type Process struct {
	gorm.Model
	PID     int32  `json:"pid"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Command string `json:"command"`
	Args    string `json:"args"`
}
