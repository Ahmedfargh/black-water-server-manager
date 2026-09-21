package nginx

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseAccessLogLine(t *testing.T) {
	service := NewNginxService()

	sampleLine := `192.168.1.50 - - [21/Sep/2026:18:25:23 +0000] "GET /api/v1/health HTTP/1.1" 200 452 "https://example.com" "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"`
	entry := service.parseAccessLine(sampleLine)

	if entry == nil {
		t.Fatalf("Expected entry to be parsed, got nil")
	}

	if entry.RemoteIP != "192.168.1.50" {
		t.Errorf("Expected IP 192.168.1.50, got %s", entry.RemoteIP)
	}
	if entry.Method != "GET" {
		t.Errorf("Expected Method GET, got %s", entry.Method)
	}
	if entry.Path != "/api/v1/health" {
		t.Errorf("Expected Path /api/v1/health, got %s", entry.Path)
	}
	if entry.StatusCode != 200 {
		t.Errorf("Expected StatusCode 200, got %d", entry.StatusCode)
	}
	if entry.BodyBytes != 452 {
		t.Errorf("Expected BodyBytes 452, got %d", entry.BodyBytes)
	}
}

func TestParseErrorLogLine(t *testing.T) {
	service := NewNginxService()

	sampleLine := `2026/09/21 18:25:23 [error] 1234#1234: *1 connect() failed (111: Connection refused) while connecting to upstream, client: 192.168.1.50, server: example.com, request: "GET /api HTTP/1.1"`
	entry := service.parseErrorLine(sampleLine)

	if entry == nil {
		t.Fatalf("Expected error entry to be parsed, got nil")
	}

	if entry.LogLevel != "error" {
		t.Errorf("Expected log level 'error', got '%s'", entry.LogLevel)
	}
	if entry.ClientIP != "192.168.1.50" {
		t.Errorf("Expected client IP 192.168.1.50, got %s", entry.ClientIP)
	}
}

func TestParseNginxConfig(t *testing.T) {
	service := NewNginxService()

	sampleConfig := `
upstream backend_nodes {
    server 127.0.0.1:8080;
    server 127.0.0.1:8081;
}

server {
    listen 80;
    listen 443 ssl;
    server_name example.com www.example.com;

    ssl_certificate /etc/letsencrypt/live/example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/example.com/privkey.pem;

    root /var/www/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://backend_nodes;
    }
}
`
	sc := &SiteConfig{Filename: "example.conf"}
	service.parseNginxContent(sampleConfig, sc)

	if len(sc.Upstreams) != 1 || sc.Upstreams[0] != "backend_nodes" {
		t.Errorf("Expected upstream 'backend_nodes', got %v", sc.Upstreams)
	}

	if sc.ServerCount != 1 || len(sc.Servers) != 1 {
		t.Fatalf("Expected 1 server block, got %d", sc.ServerCount)
	}

	server := sc.Servers[0]
	if !server.IsSSL {
		t.Errorf("Expected IsSSL to be true")
	}
	if len(server.ServerNames) != 2 || server.ServerNames[0] != "example.com" {
		t.Errorf("Expected server names [example.com www.example.com], got %v", server.ServerNames)
	}
	if len(server.Locations) != 2 {
		t.Errorf("Expected 2 location blocks, got %d", len(server.Locations))
	}
}

func TestLogAnalytics(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nginx_log_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "access.log")
	content := `192.168.1.10 - - [21/Sep/2026:18:25:20 +0000] "GET /index.html HTTP/1.1" 200 120 "-" "Mozilla/5.0"
192.168.1.10 - - [21/Sep/2026:18:25:21 +0000] "GET /api/user HTTP/1.1" 200 300 "-" "Mozilla/5.0"
192.168.1.20 - - [21/Sep/2026:18:25:22 +0000] "POST /api/login HTTP/1.1" 401 50 "-" "curl/7.68.0"
192.168.1.30 - - [21/Sep/2026:18:25:23 +0000] "GET /non-existent HTTP/1.1" 404 80 "-" "Mozilla/5.0"
`
	if err := os.WriteFile(logFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	service := NewNginxService()
	analytics, err := service.GetLogAnalytics(logFile)
	if err != nil {
		t.Fatalf("GetLogAnalytics failed: %v", err)
	}

	if analytics.TotalRequests != 4 {
		t.Errorf("Expected TotalRequests 4, got %d", analytics.TotalRequests)
	}
	if analytics.Status2xx != 2 {
		t.Errorf("Expected Status2xx 2, got %d", analytics.Status2xx)
	}
	if analytics.Status4xx != 2 {
		t.Errorf("Expected Status4xx 2, got %d", analytics.Status4xx)
	}
	if analytics.ErrorRate != 50.0 {
		t.Errorf("Expected ErrorRate 50.0, got %f", analytics.ErrorRate)
	}
}
