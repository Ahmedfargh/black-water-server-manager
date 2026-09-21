package sshservice

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateED25519Key(t *testing.T) {
	svc := NewSSHKeyService()
	keyPair, err := svc.GenerateKeyPair("ed25519", "test-key-ed25519")
	if err != nil {
		t.Fatalf("Failed to generate ED25519 key: %v", err)
	}

	if keyPair.KeyType != "ed25519" {
		t.Errorf("Expected key_type ed25519, got %s", keyPair.KeyType)
	}
	if !strings.HasPrefix(keyPair.PublicKey, "ssh-ed25519 ") {
		t.Errorf("Expected public key to start with ssh-ed25519, got %s", keyPair.PublicKey)
	}
	if !strings.Contains(keyPair.PublicKey, "test-key-ed25519") {
		t.Errorf("Expected comment in public key, got %s", keyPair.PublicKey)
	}
	if !strings.HasPrefix(keyPair.Fingerprint, "SHA256:") {
		t.Errorf("Expected fingerprint to start with SHA256:, got %s", keyPair.Fingerprint)
	}
	if !strings.Contains(keyPair.PrivateKeyPEM, "PRIVATE KEY") {
		t.Errorf("Expected private key PEM, got %s", keyPair.PrivateKeyPEM)
	}
}

func TestGenerateRSAKey(t *testing.T) {
	svc := NewSSHKeyService()
	keyPair, err := svc.GenerateKeyPair("rsa", "test-key-rsa", 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	if keyPair.KeyType != "rsa" {
		t.Errorf("Expected key_type rsa, got %s", keyPair.KeyType)
	}
	if !strings.HasPrefix(keyPair.PublicKey, "ssh-rsa ") {
		t.Errorf("Expected public key to start with ssh-rsa, got %s", keyPair.PublicKey)
	}
	if !strings.HasPrefix(keyPair.Fingerprint, "SHA256:") {
		t.Errorf("Expected fingerprint to start with SHA256:, got %s", keyPair.Fingerprint)
	}
}

func TestAuthorizedKeysFileManagement(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bw_ssh_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	authKeysPath := filepath.Join(tmpDir, "authorized_keys")
	svc := NewSSHKeyService(authKeysPath)

	key1, err := svc.GenerateKeyPair("ed25519", "user1@host")
	if err != nil {
		t.Fatalf("Failed to generate key1: %v", err)
	}
	key2, err := svc.GenerateKeyPair("ed25519", "user2@host")
	if err != nil {
		t.Fatalf("Failed to generate key2: %v", err)
	}

	// 1. Initially should not be in authorized_keys
	inAuth, err := svc.IsInAuthorizedKeys(key1.PublicKey)
	if err != nil {
		t.Fatalf("Error checking IsInAuthorizedKeys: %v", err)
	}
	if inAuth {
		t.Errorf("Expected key1 not to be in authorized_keys")
	}

	// 2. Add key1
	if err := svc.AddToAuthorizedKeys(key1.PublicKey); err != nil {
		t.Fatalf("Failed to add key1: %v", err)
	}

	inAuth, err = svc.IsInAuthorizedKeys(key1.PublicKey)
	if err != nil || !inAuth {
		t.Errorf("Expected key1 to be in authorized_keys")
	}

	// Verify file permissions (0600)
	info, err := os.Stat(authKeysPath)
	if err != nil {
		t.Fatalf("Failed to stat authorized_keys: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected permissions 0600, got %o", info.Mode().Perm())
	}

	// 3. Add key2
	if err := svc.AddToAuthorizedKeys(key2.PublicKey); err != nil {
		t.Fatalf("Failed to add key2: %v", err)
	}

	inAuth1, _ := svc.IsInAuthorizedKeys(key1.PublicKey)
	inAuth2, _ := svc.IsInAuthorizedKeys(key2.PublicKey)
	if !inAuth1 || !inAuth2 {
		t.Errorf("Expected both keys in authorized_keys")
	}

	// 4. Remove key1
	if err := svc.RemoveFromAuthorizedKeys(key1.PublicKey); err != nil {
		t.Fatalf("Failed to remove key1: %v", err)
	}

	inAuth1, _ = svc.IsInAuthorizedKeys(key1.PublicKey)
	inAuth2, _ = svc.IsInAuthorizedKeys(key2.PublicKey)
	if inAuth1 {
		t.Errorf("Expected key1 to be removed from authorized_keys")
	}
	if !inAuth2 {
		t.Errorf("Expected key2 to remain in authorized_keys")
	}
}
