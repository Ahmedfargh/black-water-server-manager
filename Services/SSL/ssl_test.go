package sslservice

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCleanHostAndPort(t *testing.T) {
	svc := NewSSLService()

	tests := []struct {
		input        string
		expectedHost string
		expectedPort string
	}{
		{"example.com", "example.com", "443"},
		{"https://example.com", "example.com", "443"},
		{"http://test.org:8080/api", "test.org", "8080"},
		{"api.server.io:8443", "api.server.io", "8443"},
	}

	for _, tc := range tests {
		host, port := svc.CleanHostAndPort(tc.input)
		if host != tc.expectedHost || port != tc.expectedPort {
			t.Errorf("CleanHostAndPort(%s) = (%s, %s); want (%s, %s)", tc.input, host, port, tc.expectedHost, tc.expectedPort)
		}
	}
}

func TestEnsureSelfSignedCertificateAndInspect(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bw_ssl_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	certPath := filepath.Join(tmpDir, "server.crt")
	keyPath := filepath.Join(tmpDir, "server.key")

	err = EnsureSelfSignedCertificate(certPath, keyPath, "localhost", "127.0.0.1", "blackwater.local")
	if err != nil {
		t.Fatalf("Failed to generate self-signed certificate: %v", err)
	}

	// Verify files exist
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		t.Fatalf("Cert file was not created")
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Fatalf("Key file was not created")
	}

	// Inspect generated certificate
	svc := NewSSLService()
	info, err := svc.InspectCertFile(certPath)
	if err != nil {
		t.Fatalf("Failed to inspect generated certificate: %v", err)
	}

	if info.SubjectCommonName != "Blackwater Internal TLS" {
		t.Errorf("Expected SubjectCommonName 'Blackwater Internal TLS', got %s", info.SubjectCommonName)
	}
	if info.IssuerOrganization != "Blackwater Server Manager" {
		t.Errorf("Expected IssuerOrganization 'Blackwater Server Manager', got %s", info.IssuerOrganization)
	}
	if info.DaysRemaining < 360 {
		t.Errorf("Expected ~365 days remaining, got %d", info.DaysRemaining)
	}
	if info.IsExpired {
		t.Errorf("Newly generated certificate should not be expired")
	}

	// Verify SANs
	foundLocalhost := false
	foundBlackwater := false
	for _, san := range info.SANs {
		if san == "localhost" {
			foundLocalhost = true
		}
		if san == "blackwater.local" {
			foundBlackwater = true
		}
	}
	if !foundLocalhost || !foundBlackwater {
		t.Errorf("Expected SANs to contain localhost and blackwater.local, got %v", info.SANs)
	}
}
