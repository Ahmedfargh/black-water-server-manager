package Services

import (
	"testing"
)

func TestFirewallStatusAndRules(t *testing.T) {
	fw := NewFirewall()
	t.Logf("Platform: %s", fw.Platform)

	status, err := fw.Status(0)
	t.Logf("Status: %s (err: %v)", status, err)
	if err != nil {
		t.Errorf("Unexpected error in Status: %v", err)
	}

	rules, err := fw.Rules()
	t.Logf("Rules: %s (err: %v)", rules, err)
	if err != nil {
		t.Errorf("Unexpected error in Rules: %v", err)
	}
}
