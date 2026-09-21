package systemd

import (
	"context"
	"testing"
)

func TestParseTabularUnits(t *testing.T) {
	sampleOutput := `
  UNIT                            LOAD   ACTIVE SUB     DESCRIPTION
  ● cron.service                    loaded active running Regular background program processing daemon
  dbus.service                    loaded active running D-Bus System Message Bus
  docker.service                  loaded active running Docker Application Container Engine
  ufw.service                     loaded active exited  Uncomplicated firewall
  snapd.socket                    loaded active listening Socket I/O for snapd
  systemd-tmpfiles-clean.timer    loaded active waiting Daily Cleanup of Temporary Directories

LOAD = Reflects whether the unit definition was properly loaded.
ACTIVE = The high-level unit activation state, i.e. generalization of SUB.
SUB = The low-level unit activation state, values depend on unit type.
6 loaded units listed.
`
	svc := NewService()
	units, err := svc.parseTabularUnits([]byte(sampleOutput), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(units) != 6 {
		t.Fatalf("expected 6 units, got %d", len(units))
	}

	if units[0].Unit != "cron.service" || units[0].Type != "service" || units[0].Active != "active" {
		t.Errorf("unexpected first unit parsed: %+v", units[0])
	}

	if units[4].Unit != "snapd.socket" || units[4].Type != "socket" {
		t.Errorf("unexpected socket unit parsed: %+v", units[4])
	}

	if units[5].Unit != "systemd-tmpfiles-clean.timer" || units[5].Type != "timer" {
		t.Errorf("unexpected timer unit parsed: %+v", units[5])
	}
}

func TestListUnitsIntegration(t *testing.T) {
	svc := NewService()
	if !svc.IsSystemdAvailable() {
		t.Skip("Systemd is not available in this test environment")
	}

	units, err := svc.ListUnits(context.Background(), "service", "")
	if err != nil {
		t.Fatalf("ListUnits error: %v", err)
	}

	t.Logf("Successfully fetched %d units from host systemd", len(units))
}
