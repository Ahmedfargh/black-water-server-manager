package cronjobs

import (
	"time"

	"strconv"

	"fmt"

	config "github.com/ahmedfargh/server-manager/Config"
	Models "github.com/ahmedfargh/server-manager/Database/Models"
)

type ClearAuditLog struct {
}

func (c *ClearAuditLog) Run() (int32, error) {
	periodic_type := config.GetKey("AUDIT_PERIOD_TYPE")
	periodic_counter, err := strconv.ParseInt(config.GetKey("AUDIT_PERIOD_COUNTER"), 10, 0)
	if err != nil {

	}
	var duration time.Duration
	switch periodic_type {
	case "S":
		duration = time.Duration(periodic_counter) * time.Second
	case "m":
		duration = time.Duration(periodic_counter) * time.Minute
	case "H":
		duration = time.Duration(periodic_counter) * time.Hour
	case "D":
		duration = time.Duration(periodic_counter) * time.Hour * 24
	case "M":
		duration = time.Duration(periodic_counter) * time.Hour * 24 * 30
	case "Y":
		duration = time.Duration(periodic_counter) * time.Hour * 24 * 365
	}
	timer := time.NewTimer(duration)

	for {
		select {
		case <-timer.C:
			config.DB.Where("created_at < ?", time.Now().Add(-duration)).Delete(&Models.AuditLog{})
			timer.Reset(duration)
			fmt.Println("Audit logs cleared")

		}
	}
	return 0, nil
}

func (c *ClearAuditLog) HandleError(err error) error {
	return err
}
