package sslservice

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	Config "github.com/ahmedfargh/server-manager/Config"
	models "github.com/ahmedfargh/server-manager/Database/Models"
)

type SSLCertificateInfo struct {
	SubjectCommonName  string    `json:"subject_common_name"`
	IssuerCommonName   string    `json:"issuer_common_name"`
	IssuerOrganization string    `json:"issuer_organization"`
	SANs               []string  `json:"sans"`
	NotBefore          time.Time `json:"not_before"`
	NotAfter           time.Time `json:"not_after"`
	DaysRemaining      int       `json:"days_remaining"`
	IsExpired          bool      `json:"is_expired"`
	IsExpiringSoon     bool      `json:"is_expiring_soon"` // <= 14 days
	SerialNumber       string    `json:"serial_number"`
	SignatureAlgorithm string    `json:"signature_algorithm"`
	PublicKeyAlgorithm string    `json:"public_key_algorithm"`
	OCSPResponseStatus string    `json:"ocsp_response_status,omitempty"`
}

type SSLService struct{}

func NewSSLService() *SSLService {
	return &SSLService{}
}

// CleanHostAndPort extracts clean hostname and port from URL or raw host string
func (s *SSLService) CleanHostAndPort(rawTarget string) (string, string) {
	rawTarget = strings.TrimSpace(rawTarget)
	if strings.Contains(rawTarget, "://") {
		parsed, err := url.Parse(rawTarget)
		if err == nil {
			rawTarget = parsed.Host
		}
	}

	host, port, err := net.SplitHostPort(rawTarget)
	if err != nil {
		host = rawTarget
		port = "443"
	}
	return host, port
}

// InspectDomainSSL connects via TLS to target host:port and parses the certificate
func (s *SSLService) InspectDomainSSL(rawTarget string, timeoutSec ...int) (*SSLCertificateInfo, error) {
	host, port := s.CleanHostAndPort(rawTarget)
	if host == "" {
		return nil, errors.New("invalid or empty host")
	}

	timeout := 10 * time.Second
	if len(timeoutSec) > 0 && timeoutSec[0] > 0 {
		timeout = time.Duration(timeoutSec[0]) * time.Second
	}

	dialer := &net.Dialer{
		Timeout: timeout,
	}

	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: true, // skip verify so we can inspect expired/self-signed certs too
	}

	targetAddr := net.JoinHostPort(host, port)
	conn, err := tls.DialWithDialer(dialer, "tcp", targetAddr, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("tls connection to %s failed: %w", targetAddr, err)
	}
	defer conn.Close()

	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return nil, fmt.Errorf("no peer certificates returned by %s", targetAddr)
	}

	cert := state.PeerCertificates[0]
	return s.parseX509Certificate(cert), nil
}

// InspectCertFile parses a local certificate file (.crt, .pem, .cer)
func (s *SSLService) InspectCertFile(certPath string) (*SSLCertificateInfo, error) {
	cleanPath := filepath.Clean(certPath)
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to parse PEM block from certificate file")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse x509 certificate: %w", err)
	}

	return s.parseX509Certificate(cert), nil
}

func (s *SSLService) parseX509Certificate(cert *x509.Certificate) *SSLCertificateInfo {
	now := time.Now()
	daysRemaining := int(time.Until(cert.NotAfter).Hours() / 24)
	isExpired := now.After(cert.NotAfter)
	isExpiringSoon := !isExpired && daysRemaining <= 14

	issuerOrg := ""
	if len(cert.Issuer.Organization) > 0 {
		issuerOrg = cert.Issuer.Organization[0]
	}

	return &SSLCertificateInfo{
		SubjectCommonName:  cert.Subject.CommonName,
		IssuerCommonName:   cert.Issuer.CommonName,
		IssuerOrganization: issuerOrg,
		SANs:               cert.DNSNames,
		NotBefore:          cert.NotBefore,
		NotAfter:           cert.NotAfter,
		DaysRemaining:      daysRemaining,
		IsExpired:          isExpired,
		IsExpiringSoon:     isExpiringSoon,
		SerialNumber:       cert.SerialNumber.String(),
		SignatureAlgorithm: cert.SignatureAlgorithm.String(),
		PublicKeyAlgorithm: cert.PublicKeyAlgorithm.String(),
	}
}

// RequestCertbotCert triggers Certbot to issue or renew a certificate
func (s *SSLService) RequestCertbotCert(domain string, email string, webroot string, standalone bool) (string, error) {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return "", errors.New("domain is required")
	}

	// Verify certbot exists
	certbotPath, err := exec.LookPath("certbot")
	if err != nil {
		return "", errors.New("certbot is not installed on this system. Please install certbot first")
	}

	args := []string{
		"certonly",
		"--non-interactive",
		"--agree-tos",
		"-d", domain,
	}

	if email != "" {
		args = append(args, "--email", email)
	} else {
		args = append(args, "--register-unsafely-without-email")
	}

	if standalone {
		args = append(args, "--standalone")
	} else if webroot != "" {
		args = append(args, "--webroot", "-w", webroot)
	} else {
		args = append(args, "--nginx")
	}

	cmd := exec.Command(certbotPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("certbot command failed: %v, output: %s", err, string(output))
	}

	return string(output), nil
}

// CheckSiteSSLAndUpdate inspects a site's SSL status and updates the database record
func (s *SSLService) CheckSiteSSLAndUpdate(siteID uint) (*SSLCertificateInfo, error) {
	var site models.Site
	if err := Config.DB.First(&site, siteID).Error; err != nil {
		return nil, fmt.Errorf("site not found: %w", err)
	}

	target := site.URL
	if target == "" {
		target = site.Health_Route
	}

	info, err := s.InspectDomainSSL(target)
	now := time.Now()
	site.SSLLastCheck = &now

	if err != nil {
		site.SSLEnabled = false
		_ = Config.DB.Save(&site)
		return nil, err
	}

	site.SSLEnabled = !info.IsExpired
	site.SSLExpiryDate = &info.NotAfter
	site.SSLIssuer = info.IssuerCommonName
	if site.SSLIssuer == "" {
		site.SSLIssuer = info.IssuerOrganization
	}
	site.SSLDomains = strings.Join(info.SANs, ", ")

	_ = Config.DB.Save(&site)
	return info, nil
}
