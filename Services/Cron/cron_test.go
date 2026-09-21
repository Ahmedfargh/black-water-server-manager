package cron

import (
	"testing"
)

func TestHumanizeExpression(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"@reboot", "Run once at system startup"},
		{"0 0 * * *", "Every day at midnight (00:00)"},
		{"@daily", "Every day at midnight (00:00)"},
		{"*/15 * * * *", "Every 15 minutes"},
		{"30 2 * * *", "Daily at 02:30"},
		{"0 * * * *", "Every hour at minute 0"},
		{"15 * * * *", "Every hour at minute 15"},
	}

	for _, tt := range tests {
		got := HumanizeExpression(tt.expr)
		if got != tt.expected {
			t.Errorf("HumanizeExpression(%q) = %q; want %q", tt.expr, got, tt.expected)
		}
	}
}

func TestParseCronLine(t *testing.T) {
	line := "*/5 * * * * /usr/local/bin/backup.sh > /dev/null 2>&1"
	job := parseCronLine(line, true, "Database backup", line)

	if job.Expression != "*/5 * * * *" {
		t.Errorf("expected expression '*/5 * * * *', got '%s'", job.Expression)
	}
	if job.Command != "/usr/local/bin/backup.sh > /dev/null 2>&1" {
		t.Errorf("expected command '/usr/local/bin/backup.sh > /dev/null 2>&1', got '%s'", job.Command)
	}
	if job.Comment != "Database backup" {
		t.Errorf("expected comment 'Database backup', got '%s'", job.Comment)
	}
	if !job.Enabled {
		t.Errorf("expected job to be enabled")
	}
	if job.ID == "" {
		t.Errorf("expected valid non-empty job ID")
	}
}
