package systemd

import "time"

// UnitInfo represents a systemd unit (service, timer, socket, etc.)
type UnitInfo struct {
	Unit        string `json:"unit"`
	Load        string `json:"load"`
	Active      string `json:"active"`
	Sub         string `json:"sub"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Enabled     string `json:"enabled,omitempty"`
}

// UnitActionRequest represents a mutation request on a systemd unit
type UnitActionRequest struct {
	Action string `json:"action" binding:"required"` // start, stop, restart, reload, enable, disable
}

// UnitActionResponse represents the execution result of a unit action
type UnitActionResponse struct {
	Unit      string `json:"unit"`
	Action    string `json:"action"`
	Success   bool   `json:"success"`
	Output    string `json:"output,omitempty"`
	DurationMs int64 `json:"duration_ms"`
}

// SystemdOverview contains system-level service stats
type SystemdOverview struct {
	Available    bool       `json:"available"`
	TotalUnits   int        `json:"total_units"`
	ActiveUnits  int        `json:"active_units"`
	FailedUnits  int        `json:"failed_units"`
	ServiceUnits int        `json:"service_units"`
	TimerUnits   int        `json:"timer_units"`
	SocketUnits  int        `json:"socket_units"`
	Units        []UnitInfo `json:"units"`
	Timestamp    time.Time  `json:"timestamp"`
}
