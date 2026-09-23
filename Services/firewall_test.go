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

func TestValidateIPOrCIDR(t *testing.T) {
	validCases := []struct {
		input  string
		isIPv6 bool
	}{
		{"192.168.1.1", false},
		{"10.0.0.1", false},
		{"127.0.0.1", false},
		{"172.16.0.0/16", false},
		{"192.168.0.0/24", false},
		{"2001:db8::1", true},
		{"::1", true},
		{"2001:db8::/32", true},
	}

	for _, tc := range validCases {
		cleaned, isIPv6, err := ValidateIPOrCIDR(tc.input)
		if err != nil {
			t.Errorf("Expected valid IP/CIDR for %s, got error: %v", tc.input, err)
		}
		if isIPv6 != tc.isIPv6 {
			t.Errorf("Expected isIPv6=%v for %s, got %v", tc.isIPv6, tc.input, isIPv6)
		}
		if cleaned != tc.input {
			t.Errorf("Expected cleaned=%s, got %s", tc.input, cleaned)
		}
	}

	invalidCases := []string{
		"",
		"   ",
		"999.999.999.999",
		"192.168.1.1; rm -rf /",
		"192.168.1.1 | cat /etc/passwd",
		"not-an-ip",
		"10.0.0.1/35",
		"192.168.1.1/abc",
	}

	for _, tc := range invalidCases {
		_, _, err := ValidateIPOrCIDR(tc)
		if err == nil {
			t.Errorf("Expected error for invalid input %q, but got nil", tc)
		}
	}
}

func TestFirewallBlockInvalidIP(t *testing.T) {
	fw := NewFirewall()

	// Should reject invalid IP before attempting command execution
	_, err := fw.BlockIP("invalid-ip-address; whoami", 0)
	if err == nil {
		t.Error("Expected error when blocking invalid IP, got nil")
	}

	_, err = fw.UnblockIP("invalid-ip-address; whoami", 0)
	if err == nil {
		t.Error("Expected error when unblocking invalid IP, got nil")
	}
}

