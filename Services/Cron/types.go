package cron

// CronJob represents a scheduled task in the system crontab
type CronJob struct {
	ID            string `json:"id"`
	Expression    string `json:"expression"`
	ScheduleHuman string `json:"schedule_human"`
	Command       string `json:"command"`
	Comment       string `json:"comment"`
	Enabled       bool   `json:"enabled"`
	Raw           string `json:"raw,omitempty"`
}

// CronMutationRequest payload for creating or updating a cron job
type CronMutationRequest struct {
	ID         string `json:"id,omitempty"`
	Expression string `json:"expression" binding:"required"`
	Command    string `json:"command" binding:"required"`
	Comment    string `json:"comment"`
	Enabled    bool   `json:"enabled"`
}

// CronRunResponse returns the execution result of manually triggering a cron command
type CronRunResponse struct {
	ID         string `json:"id"`
	Command    string `json:"command"`
	Output     string `json:"output"`
	Success    bool   `json:"success"`
	DurationMs int64  `json:"duration_ms"`
}
