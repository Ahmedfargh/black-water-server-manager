package sshservice

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
)

type GeneratedKeyPair struct {
	KeyType        string `json:"key_type"`
	PublicKey      string `json:"public_key"`
	PrivateKeyPEM  string `json:"private_key_pem"`
	Fingerprint    string `json:"fingerprint"`
	Comment        string `json:"comment"`
}

type SSHKeyService struct {
	authorizedKeysPath string
	mu                 sync.Mutex
}

func NewSSHKeyService(customPath ...string) *SSHKeyService {
	path := ""
	if len(customPath) > 0 && customPath[0] != "" {
		path = customPath[0]
	} else {
		envPath := os.Getenv("SSH_AUTHORIZED_KEYS_PATH")
		if envPath != "" {
			path = envPath
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				home = "/root"
			}
			path = filepath.Join(home, ".ssh", "authorized_keys")
		}
	}
	return &SSHKeyService{
		authorizedKeysPath: path,
	}
}

// GenerateKeyPair generates a new ED25519 or RSA key pair
func (s *SSHKeyService) GenerateKeyPair(keyType string, comment string, rsaBits ...int) (*GeneratedKeyPair, error) {
	keyType = strings.ToLower(strings.TrimSpace(keyType))
	if comment == "" {
		comment = "blackwater-managed-key"
	}

	switch keyType {
	case "ed25519":
		pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ed25519 key: %w", err)
		}

		// Convert to PKCS#8 PEM
		pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privKey)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal ed25519 private key: %w", err)
		}
		privPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "OPENSSH PRIVATE KEY",
			Bytes: pkcs8Bytes,
		})

		// Convert public key to OpenSSH format
		sshPubKey, err := ssh.NewPublicKey(pubKey)
		if err != nil {
			return nil, fmt.Errorf("failed to convert ed25519 public key to ssh format: %w", err)
		}
		pubKeyStr := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshPubKey)))
		if comment != "" {
			pubKeyStr = fmt.Sprintf("%s %s", pubKeyStr, comment)
		}

		fingerprint := s.CalculateFingerprint(sshPubKey)

		return &GeneratedKeyPair{
			KeyType:       "ed25519",
			PublicKey:     pubKeyStr,
			PrivateKeyPEM: string(privPEM),
			Fingerprint:   fingerprint,
			Comment:       comment,
		}, nil

	case "rsa":
		bits := 4096
		if len(rsaBits) > 0 && rsaBits[0] >= 2048 {
			bits = rsaBits[0]
		}
		privKey, err := rsa.GenerateKey(rand.Reader, bits)
		if err != nil {
			return nil, fmt.Errorf("failed to generate rsa key: %w", err)
		}

		// Convert to PKCS#1 PEM
		privBytes := x509.MarshalPKCS1PrivateKey(privKey)
		privPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privBytes,
		})

		// Convert public key to OpenSSH format
		sshPubKey, err := ssh.NewPublicKey(&privKey.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to convert rsa public key to ssh format: %w", err)
		}
		pubKeyStr := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshPubKey)))
		if comment != "" {
			pubKeyStr = fmt.Sprintf("%s %s", pubKeyStr, comment)
		}

		fingerprint := s.CalculateFingerprint(sshPubKey)

		return &GeneratedKeyPair{
			KeyType:       "rsa",
			PublicKey:     pubKeyStr,
			PrivateKeyPEM: string(privPEM),
			Fingerprint:   fingerprint,
			Comment:       comment,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported key type: %s (supported: ed25519, rsa)", keyType)
	}
}

// CalculateFingerprint generates the standard OpenSSH SHA256 fingerprint: SHA256:base64...
func (s *SSHKeyService) CalculateFingerprint(pubKey ssh.PublicKey) string {
	hash := sha256.Sum256(pubKey.Marshal())
	b64 := base64.RawStdEncoding.EncodeToString(hash[:])
	return fmt.Sprintf("SHA256:%s", b64)
}

// ParsePublicKey parses an OpenSSH public key string and returns the parsed key, comment, and fingerprint
func (s *SSHKeyService) ParsePublicKey(rawKey string) (ssh.PublicKey, string, string, error) {
	trimmed := strings.TrimSpace(rawKey)
	if trimmed == "" {
		return nil, "", "", errors.New("empty public key")
	}

	pubKey, comment, _, _, err := ssh.ParseAuthorizedKey([]byte(trimmed))
	if err != nil {
		return nil, "", "", fmt.Errorf("invalid openssh public key format: %w", err)
	}

	fingerprint := s.CalculateFingerprint(pubKey)
	return pubKey, comment, fingerprint, nil
}

// EnsureSSHDirectory creates the ~/.ssh directory if missing and enforces 0700 permissions
func (s *SSHKeyService) EnsureSSHDirectory() error {
	dir := filepath.Dir(s.authorizedKeysPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create ssh directory %s: %w", dir, err)
	}
	// Hardening: enforce 0700
	_ = os.Chmod(dir, 0700)
	return nil
}

// AddToAuthorizedKeys appends a public key to the host's authorized_keys file safely
func (s *SSHKeyService) AddToAuthorizedKeys(publicKeyStr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.EnsureSSHDirectory(); err != nil {
		return err
	}

	pubKey, _, _, err := s.ParsePublicKey(publicKeyStr)
	if err != nil {
		return err
	}

	targetMarshaled := string(pubKey.Marshal())

	// Read existing keys
	existingBytes, err := os.ReadFile(s.authorizedKeysPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read authorized_keys: %w", err)
	}

	lines := strings.Split(string(existingBytes), "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}
		parsedExisting, _, _, _, pErr := ssh.ParseAuthorizedKey([]byte(trimmedLine))
		if pErr == nil && string(parsedExisting.Marshal()) == targetMarshaled {
			// Key is already in authorized_keys
			return nil
		}
	}

	// Append key
	normalizedKey := strings.TrimSpace(publicKeyStr) + "\n"
	f, err := os.OpenFile(s.authorizedKeysPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to open authorized_keys for appending: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(normalizedKey); err != nil {
		return fmt.Errorf("failed to write to authorized_keys: %w", err)
	}

	// Defensive hardening: Ensure 0600
	_ = os.Chmod(s.authorizedKeysPath, 0600)
	return nil
}

// RemoveFromAuthorizedKeys removes a public key from the host's authorized_keys file
func (s *SSHKeyService) RemoveFromAuthorizedKeys(publicKeyStr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := os.Stat(s.authorizedKeysPath); os.IsNotExist(err) {
		return nil
	}

	pubKey, _, _, err := s.ParsePublicKey(publicKeyStr)
	if err != nil {
		return err
	}

	targetMarshaled := string(pubKey.Marshal())

	existingBytes, err := os.ReadFile(s.authorizedKeysPath)
	if err != nil {
		return fmt.Errorf("failed to read authorized_keys: %w", err)
	}

	lines := strings.Split(string(existingBytes), "\n")
	var remainingLines []string

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}
		if strings.HasPrefix(trimmedLine, "#") {
			remainingLines = append(remainingLines, line)
			continue
		}

		parsedExisting, _, _, _, pErr := ssh.ParseAuthorizedKey([]byte(trimmedLine))
		if pErr == nil && string(parsedExisting.Marshal()) == targetMarshaled {
			// Skip this key (removing it)
			continue
		}
		remainingLines = append(remainingLines, line)
	}

	output := strings.Join(remainingLines, "\n")
	if len(remainingLines) > 0 {
		output += "\n"
	}

	if err := os.WriteFile(s.authorizedKeysPath, []byte(output), 0600); err != nil {
		return fmt.Errorf("failed to write updated authorized_keys: %w", err)
	}

	return nil
}

// IsInAuthorizedKeys checks if a public key exists in authorized_keys
func (s *SSHKeyService) IsInAuthorizedKeys(publicKeyStr string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := os.Stat(s.authorizedKeysPath); os.IsNotExist(err) {
		return false, nil
	}

	pubKey, _, _, err := s.ParsePublicKey(publicKeyStr)
	if err != nil {
		return false, err
	}

	targetMarshaled := string(pubKey.Marshal())

	existingBytes, err := os.ReadFile(s.authorizedKeysPath)
	if err != nil {
		return false, err
	}

	lines := strings.Split(string(existingBytes), "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}
		parsedExisting, _, _, _, pErr := ssh.ParseAuthorizedKey([]byte(trimmedLine))
		if pErr == nil && string(parsedExisting.Marshal()) == targetMarshaled {
			return true, nil
		}
	}

	return false, nil
}
