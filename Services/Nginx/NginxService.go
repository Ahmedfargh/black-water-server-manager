package nginx

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ServerBlock represents a parsed Nginx server block
type ServerBlock struct {
	ServerNames []string       `json:"server_names"`
	ListenPorts []string       `json:"listen_ports"`
	IsSSL       bool           `json:"is_ssl"`
	SSLCertPath string         `json:"ssl_cert_path,omitempty"`
	SSLKeyPath  string         `json:"ssl_key_path,omitempty"`
	Root        string         `json:"root,omitempty"`
	Index       string         `json:"index,omitempty"`
	Locations   []LocationRule `json:"locations"`
}

// LocationRule represents a parsed location {} directive
type LocationRule struct {
	Path        string `json:"path"`
	ProxyPass   string `json:"proxy_pass,omitempty"`
	FastCGIPass string `json:"fastcgi_pass,omitempty"`
	Root        string `json:"root,omitempty"`
	TryFiles    string `json:"try_files,omitempty"`
}

// SiteConfig represents an Nginx configuration file
type SiteConfig struct {
	Filename    string        `json:"filename"`
	Path        string        `json:"path"`
	IsEnabled   bool          `json:"is_enabled"`
	IsAvailable bool          `json:"is_available"`
	RawContent  string        `json:"raw_content,omitempty"`
	ServerCount int           `json:"server_count"`
	Servers     []ServerBlock `json:"servers"`
	Upstreams   []string      `json:"upstreams"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// AccessLogEntry represents a parsed Nginx access log row
type AccessLogEntry struct {
	RemoteIP     string    `json:"remote_ip"`
	Timestamp    time.Time `json:"timestamp"`
	RawTimestamp string    `json:"raw_timestamp"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	Protocol     string    `json:"protocol"`
	StatusCode   int       `json:"status_code"`
	BodyBytes    int64     `json:"body_bytes"`
	Referer      string    `json:"referer"`
	UserAgent    string    `json:"user_agent"`
	Raw          string    `json:"raw"`
}

// ErrorLogEntry represents a parsed Nginx error log row
type ErrorLogEntry struct {
	Timestamp    time.Time `json:"timestamp"`
	RawTimestamp string    `json:"raw_timestamp"`
	LogLevel     string    `json:"log_level"`
	PID          string    `json:"pid"`
	Message      string    `json:"message"`
	ClientIP     string    `json:"client_ip,omitempty"`
	Server       string    `json:"server,omitempty"`
	Raw          string    `json:"raw"`
}

// LogAnalytics represents aggregated statistics from access logs
type LogAnalytics struct {
	TotalRequests   int64            `json:"total_requests"`
	TotalBytes      int64            `json:"total_bytes"`
	Status2xx       int64            `json:"status_2xx"`
	Status3xx       int64            `json:"status_3xx"`
	Status4xx       int64            `json:"status_4xx"`
	Status5xx       int64            `json:"status_5xx"`
	ErrorRate       float64          `json:"error_rate"`
	TopIPs          []ItemCount      `json:"top_ips"`
	TopPaths        []ItemCount      `json:"top_paths"`
	TopUserAgents   []ItemCount      `json:"top_user_agents"`
	MethodCounts    map[string]int64 `json:"method_counts"`
	StatusCodeMap   map[int]int64    `json:"status_code_map"`
	HourlyBreakdown []HourlyTraffic  `json:"hourly_breakdown"`
}

type ItemCount struct {
	Item  string `json:"item"`
	Count int64  `json:"count"`
}

type HourlyTraffic struct {
	Hour     string `json:"hour"`
	Requests int64  `json:"requests"`
	Errors   int64  `json:"errors"`
}

// NginxOverview provides quick health and setup status
type NginxOverview struct {
	IsInstalled   bool   `json:"is_installed"`
	IsRunning     bool   `json:"is_running"`
	Version       string `json:"version"`
	SitesCount    int    `json:"sites_count"`
	EnabledCount  int    `json:"enabled_count"`
	DisabledCount int    `json:"disabled_count"`
	ConfigValid   bool   `json:"config_valid"`
	ConfigOutput  string `json:"config_output"`
}

type NginxService struct {
	SitesAvailableDir string
	SitesEnabledDir   string
	ConfDDir          string
	MainConfigPath    string
	LogDir            string
}

func NewNginxService() *NginxService {
	return &NginxService{
		SitesAvailableDir: "/etc/nginx/sites-available",
		SitesEnabledDir:   "/etc/nginx/sites-enabled",
		ConfDDir:          "/etc/nginx/conf.d",
		MainConfigPath:    "/etc/nginx/nginx.conf",
		LogDir:            "/var/log/nginx",
	}
}

// --- Configuration Parsing Methods ---

// GetOverview checks Nginx service status and site counts
func (s *NginxService) GetOverview() NginxOverview {
	overview := NginxOverview{
		IsInstalled: false,
		IsRunning:   false,
	}

	// 1. Check version
	out, err := exec.Command("nginx", "-v").CombinedOutput()
	if err == nil || strings.Contains(string(out), "nginx") {
		overview.IsInstalled = true
		rawVer := string(out)
		if idx := strings.Index(rawVer, "nginx/"); idx != -1 {
			overview.Version = strings.TrimSpace(rawVer[idx:])
		} else {
			overview.Version = strings.TrimSpace(rawVer)
		}
	}

	// 2. Check service status
	statusOut, _ := exec.Command("systemctl", "is-active", "nginx").CombinedOutput()
	if strings.TrimSpace(string(statusOut)) == "active" {
		overview.IsRunning = true
	}

	// 3. Test configuration
	testValid, testOut := s.TestConfiguration()
	overview.ConfigValid = testValid
	overview.ConfigOutput = testOut

	// 4. Sites counts
	sites, _ := s.ListSites(false)
	overview.SitesCount = len(sites)
	for _, site := range sites {
		if site.IsEnabled {
			overview.EnabledCount++
		} else {
			overview.DisabledCount++
		}
	}

	return overview
}

// ListSites scans all available and enabled virtual hosts
func (s *NginxService) ListSites(includeContent bool) ([]SiteConfig, error) {
	siteMap := make(map[string]*SiteConfig)

	// Helper to check if file is enabled via symlink in sites-enabled
	checkEnabled := func(filename string) bool {
		enabledPath := filepath.Join(s.SitesEnabledDir, filename)
		if _, err := os.Lstat(enabledPath); err == nil {
			return true
		}
		// Also check without .conf extension
		trimmed := strings.TrimSuffix(filename, ".conf")
		if _, err := os.Lstat(filepath.Join(s.SitesEnabledDir, trimmed)); err == nil {
			return true
		}
		return false
	}

	// 1. Scan sites-available
	if entries, err := os.ReadDir(s.SitesAvailableDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			fullPath := filepath.Join(s.SitesAvailableDir, name)
			info, _ := entry.Info()

			sc := &SiteConfig{
				Filename:    name,
				Path:        fullPath,
				IsAvailable: true,
				IsEnabled:   checkEnabled(name),
				UpdatedAt:   time.Now(),
			}
			if info != nil {
				sc.UpdatedAt = info.ModTime()
			}

			if content, err := os.ReadFile(fullPath); err == nil {
				if includeContent {
					sc.RawContent = string(content)
				}
				s.parseNginxContent(string(content), sc)
			}

			siteMap[name] = sc
		}
	}

	// 2. Scan conf.d
	if entries, err := os.ReadDir(s.ConfDDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".conf") {
				continue
			}
			name := entry.Name()
			if _, exists := siteMap[name]; exists {
				continue
			}
			fullPath := filepath.Join(s.ConfDDir, name)
			info, _ := entry.Info()

			sc := &SiteConfig{
				Filename:    name,
				Path:        fullPath,
				IsAvailable: true,
				IsEnabled:   true, // Files in conf.d are usually active directly
				UpdatedAt:   time.Now(),
			}
			if info != nil {
				sc.UpdatedAt = info.ModTime()
			}

			if content, err := os.ReadFile(fullPath); err == nil {
				if includeContent {
					sc.RawContent = string(content)
				}
				s.parseNginxContent(string(content), sc)
			}

			siteMap[name] = sc
		}
	}

	// Convert map to sorted slice
	var results []SiteConfig
	for _, sc := range siteMap {
		results = append(results, *sc)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Filename < results[j].Filename
	})

	return results, nil
}

// GetSite retrieves a specific site configuration
func (s *NginxService) GetSite(filename string) (*SiteConfig, error) {
	sites, err := s.ListSites(true)
	if err != nil {
		return nil, err
	}

	for _, site := range sites {
		if site.Filename == filename {
			return &site, nil
		}
	}

	return nil, fmt.Errorf("site config '%s' not found", filename)
}

// parseNginxContent extracts servers, upstreams, and directives using regex & block parsing
func (s *NginxService) parseNginxContent(content string, sc *SiteConfig) {
	// 1. Upstreams
	upstreamRegex := regexp.MustCompile(`(?m)^\s*upstream\s+([A-Za-z0-9_\-\.]+)\s*\{`)
	for _, match := range upstreamRegex.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			sc.Upstreams = append(sc.Upstreams, match[1])
		}
	}

	// 2. Extract server { ... } blocks
	serverBlocks := extractBlocks(content, "server")
	sc.ServerCount = len(serverBlocks)

	for _, blockStr := range serverBlocks {
		sb := ServerBlock{
			ServerNames: []string{},
			ListenPorts: []string{},
			Locations:   []LocationRule{},
		}

		// Server Names
		snRegex := regexp.MustCompile(`(?m)^\s*server_name\s+([^;]+);`)
		if match := snRegex.FindStringSubmatch(blockStr); len(match) > 1 {
			names := strings.Fields(match[1])
			sb.ServerNames = append(sb.ServerNames, names...)
		}

		// Listen Ports & SSL
		listenRegex := regexp.MustCompile(`(?m)^\s*listen\s+([^;]+);`)
		for _, match := range listenRegex.FindAllStringSubmatch(blockStr, -1) {
			if len(match) > 1 {
				listenStr := strings.TrimSpace(match[1])
				sb.ListenPorts = append(sb.ListenPorts, listenStr)
				if strings.Contains(listenStr, "ssl") || strings.Contains(listenStr, "443") {
					sb.IsSSL = true
				}
			}
		}

		// SSL Certificate Paths
		certRegex := regexp.MustCompile(`(?m)^\s*ssl_certificate\s+([^;]+);`)
		if match := certRegex.FindStringSubmatch(blockStr); len(match) > 1 {
			sb.SSLCertPath = strings.TrimSpace(match[1])
			sb.IsSSL = true
		}
		keyRegex := regexp.MustCompile(`(?m)^\s*ssl_certificate_key\s+([^;]+);`)
		if match := keyRegex.FindStringSubmatch(blockStr); len(match) > 1 {
			sb.SSLKeyPath = strings.TrimSpace(match[1])
		}

		// Root & Index
		rootRegex := regexp.MustCompile(`(?m)^\s*root\s+([^;]+);`)
		if match := rootRegex.FindStringSubmatch(blockStr); len(match) > 1 {
			sb.Root = strings.TrimSpace(match[1])
		}
		indexRegex := regexp.MustCompile(`(?m)^\s*index\s+([^;]+);`)
		if match := indexRegex.FindStringSubmatch(blockStr); len(match) > 1 {
			sb.Index = strings.TrimSpace(match[1])
		}

		// Locations
		locRegex := regexp.MustCompile(`(?m)location\s+([^{]+)\s*\{([^}]+)\}`)
		for _, locMatch := range locRegex.FindAllStringSubmatch(blockStr, -1) {
			if len(locMatch) > 2 {
				locPath := strings.TrimSpace(locMatch[1])
				locBody := locMatch[2]

				lr := LocationRule{Path: locPath}
				if pp := regexp.MustCompile(`proxy_pass\s+([^;]+);`).FindStringSubmatch(locBody); len(pp) > 1 {
					lr.ProxyPass = strings.TrimSpace(pp[1])
				}
				if fc := regexp.MustCompile(`fastcgi_pass\s+([^;]+);`).FindStringSubmatch(locBody); len(fc) > 1 {
					lr.FastCGIPass = strings.TrimSpace(fc[1])
				}
				if rt := regexp.MustCompile(`root\s+([^;]+);`).FindStringSubmatch(locBody); len(rt) > 1 {
					lr.Root = strings.TrimSpace(rt[1])
				}
				if tf := regexp.MustCompile(`try_files\s+([^;]+);`).FindStringSubmatch(locBody); len(tf) > 1 {
					lr.TryFiles = strings.TrimSpace(tf[1])
				}
				sb.Locations = append(sb.Locations, lr)
			}
		}

		sc.Servers = append(sc.Servers, sb)
	}
}

// extractBlocks extracts nested blocks like server { ... } handling braces
func extractBlocks(content string, blockKeyword string) []string {
	var blocks []string
	idx := 0
	target := blockKeyword

	for {
		start := strings.Index(content[idx:], target)
		if start == -1 {
			break
		}
		absStart := idx + start
		braceStart := strings.Index(content[absStart:], "{")
		if braceStart == -1 {
			break
		}
		absBraceStart := absStart + braceStart

		// Track braces
		depth := 1
		curr := absBraceStart + 1
		for curr < len(content) && depth > 0 {
			if content[curr] == '{' {
				depth++
			} else if content[curr] == '}' {
				depth--
			}
			curr++
		}

		if depth == 0 {
			blocks = append(blocks, content[absStart:curr])
		}
		idx = curr
	}

	return blocks
}

// SaveSite creates or updates an Nginx site config file safely with rollback
func (s *NginxService) SaveSite(filename string, content string) error {
	cleanName := filepath.Base(filename)
	targetPath := filepath.Join(s.SitesAvailableDir, cleanName)

	_ = os.MkdirAll(s.SitesAvailableDir, 0755)

	// Backup existing if present
	var backup []byte
	if exists, _ := os.ReadFile(targetPath); exists != nil {
		backup = exists
	}

	if err := os.WriteFile(targetPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write site file: %w", err)
	}

	// Test syntax
	valid, testOut := s.TestConfiguration()
	if !valid {
		// Rollback
		if backup != nil {
			_ = os.WriteFile(targetPath, backup, 0644)
		} else {
			_ = os.Remove(targetPath)
		}
		return fmt.Errorf("nginx syntax test failed: %s", testOut)
	}

	return nil
}

// ToggleSite enables or disables an Nginx site
func (s *NginxService) ToggleSite(filename string, enable bool) error {
	cleanName := filepath.Base(filename)
	availablePath := filepath.Join(s.SitesAvailableDir, cleanName)
	enabledPath := filepath.Join(s.SitesEnabledDir, cleanName)

	if enable {
		if _, err := os.Stat(availablePath); os.IsNotExist(err) {
			return fmt.Errorf("site '%s' not found in sites-available", cleanName)
		}
		_ = os.MkdirAll(s.SitesEnabledDir, 0755)
		_ = os.Remove(enabledPath) // remove broken link if any
		if err := os.Symlink(availablePath, enabledPath); err != nil {
			return fmt.Errorf("failed to enable site symlink: %w", err)
		}

		// Test after enabling
		if valid, out := s.TestConfiguration(); !valid {
			_ = os.Remove(enabledPath)
			return fmt.Errorf("nginx configuration error when enabling site: %s", out)
		}
	} else {
		_ = os.Remove(enabledPath)
		// Check variant without .conf
		_ = os.Remove(filepath.Join(s.SitesEnabledDir, strings.TrimSuffix(cleanName, ".conf")))
	}

	return nil
}

// DeleteSite removes site from sites-available and sites-enabled
func (s *NginxService) DeleteSite(filename string) error {
	cleanName := filepath.Base(filename)
	_ = s.ToggleSite(cleanName, false)
	availablePath := filepath.Join(s.SitesAvailableDir, cleanName)
	return os.Remove(availablePath)
}

// TestConfiguration runs `nginx -t`
func (s *NginxService) TestConfiguration() (bool, string) {
	cmd := exec.Command("nginx", "-t")
	out, err := cmd.CombinedOutput()
	outputStr := string(out)
	if err == nil && (strings.Contains(outputStr, "successful") || strings.Contains(outputStr, "syntax is ok")) {
		return true, strings.TrimSpace(outputStr)
	}
	return false, strings.TrimSpace(outputStr)
}

// Reload reloads the Nginx daemon
func (s *NginxService) Reload() error {
	if valid, out := s.TestConfiguration(); !valid {
		return fmt.Errorf("cannot reload nginx, config is invalid: %s", out)
	}
	cmd := exec.Command("systemctl", "reload", "nginx")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("nginx reload failed: %s (%w)", string(out), err)
	}
	return nil
}

// Restart restarts the Nginx service
func (s *NginxService) Restart() error {
	cmd := exec.Command("systemctl", "restart", "nginx")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("nginx restart failed: %s (%w)", string(out), err)
	}
	return nil
}

// --- Log Parsing & Analytics Methods ---

// DiscoverLogFiles finds all access and error log files in /var/log/nginx
func (s *NginxService) DiscoverLogFiles() (accessLogs []string, errorLogs []string) {
	if entries, err := os.ReadDir(s.LogDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			fullPath := filepath.Join(s.LogDir, name)

			if strings.Contains(name, "error") {
				errorLogs = append(errorLogs, fullPath)
			} else if strings.Contains(name, "access") || strings.HasSuffix(name, ".log") {
				accessLogs = append(accessLogs, fullPath)
			}
		}
	}

	if len(accessLogs) == 0 {
		accessLogs = append(accessLogs, "/var/log/nginx/access.log")
	}
	if len(errorLogs) == 0 {
		errorLogs = append(errorLogs, "/var/log/nginx/error.log")
	}

	return accessLogs, errorLogs
}

// Combined Log Format Regex: 127.0.0.1 - - [21/Sep/2026:18:25:23 +0000] "GET /api/status HTTP/1.1" 200 1234 "referer" "user-agent"
var combinedLogRegex = regexp.MustCompile(`^(\S+)\s+\S+\s+\S+\s+\[([^\]]+)\]\s+"([A-Z]+)\s+([^"]*?)(?:\s+HTTP\/[0-9\.]+)?\"\s+(\d{3})\s+(\d+|-)(?:\s+"([^"]*)"\s+"([^"]*)")?`)

// ParseAccessLogs reads and filters access logs from a file
func (s *NginxService) ParseAccessLogs(filePath string, limit int, statusFilter int, search string) ([]AccessLogEntry, error) {
	if filePath == "" {
		filePath = filepath.Join(s.LogDir, "access.log")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open access log: %w", err)
	}
	defer file.Close()

	var entries []AccessLogEntry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		entry := s.parseAccessLine(line)
		if entry == nil {
			continue
		}

		// Filters
		if statusFilter > 0 && entry.StatusCode != statusFilter {
			continue
		}
		if search != "" {
			searchLower := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(entry.Path), searchLower) &&
				!strings.Contains(strings.ToLower(entry.RemoteIP), searchLower) &&
				!strings.Contains(strings.ToLower(entry.UserAgent), searchLower) {
				continue
			}
		}

		entries = append(entries, *entry)
	}

	// Reverse to get latest entries first
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	return entries, nil
}

func (s *NginxService) parseAccessLine(line string) *AccessLogEntry {
	match := combinedLogRegex.FindStringSubmatch(line)
	if len(match) < 6 {
		// Fallback simple scanner
		parts := strings.Fields(line)
		if len(parts) >= 9 {
			status, _ := strconv.Atoi(parts[8])
			return &AccessLogEntry{
				RemoteIP:   parts[0],
				Method:     strings.Trim(parts[5], `"`),
				Path:       parts[6],
				StatusCode: status,
				Raw:        line,
			}
		}
		return nil
	}

	ip := match[1]
	timeStr := match[2]
	method := match[3]
	path := match[4]
	status, _ := strconv.Atoi(match[5])
	bytesVal := int64(0)
	if match[6] != "-" {
		bytesVal, _ = strconv.ParseInt(match[6], 10, 64)
	}
	referer := ""
	ua := ""
	if len(match) > 7 {
		referer = match[7]
	}
	if len(match) > 8 {
		ua = match[8]
	}

	parsedTime, _ := time.Parse("02/Jan/2006:15:04:05 -0700", timeStr)

	return &AccessLogEntry{
		RemoteIP:     ip,
		Timestamp:    parsedTime,
		RawTimestamp: timeStr,
		Method:       method,
		Path:         path,
		StatusCode:   status,
		BodyBytes:    bytesVal,
		Referer:      referer,
		UserAgent:    ua,
		Raw:          line,
	}
}

// Error Log Format Regex: 2026/09/21 18:25:23 [error] 1234#1234: *1 message, client: 127.0.0.1, server: domain.com
var errorLogRegex = regexp.MustCompile(`^(\d{4}\/\d{2}\/\d{2}\s+\d{2}:\d{2}:\d{2})\s+\[([a-z]+)\]\s+(\d+#\d+):\s+(.*)`)

// ParseErrorLogs reads and parses Nginx error log file
func (s *NginxService) ParseErrorLogs(filePath string, limit int, levelFilter string, search string) ([]ErrorLogEntry, error) {
	if filePath == "" {
		filePath = filepath.Join(s.LogDir, "error.log")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open error log: %w", err)
	}
	defer file.Close()

	var entries []ErrorLogEntry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		entry := s.parseErrorLine(line)
		if entry == nil {
			continue
		}

		if levelFilter != "" && !strings.EqualFold(entry.LogLevel, levelFilter) {
			continue
		}
		if search != "" {
			searchLower := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(entry.Message), searchLower) &&
				!strings.Contains(strings.ToLower(entry.ClientIP), searchLower) &&
				!strings.Contains(strings.ToLower(entry.Server), searchLower) {
				continue
			}
		}

		entries = append(entries, *entry)
	}

	// Reverse to get latest entries first
	for i, j := 0, len(entries)-1; i < j; i, j = i-1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	return entries, nil
}

func (s *NginxService) parseErrorLine(line string) *ErrorLogEntry {
	match := errorLogRegex.FindStringSubmatch(line)
	if len(match) < 5 {
		return &ErrorLogEntry{
			LogLevel: "unknown",
			Message:  line,
			Raw:      line,
		}
	}

	timeStr := match[1]
	level := match[2]
	pid := match[3]
	msg := match[4]

	parsedTime, _ := time.Parse("2006/01/02 15:04:05", timeStr)

	// Extract client IP and server if present
	clientIP := ""
	if cMatch := regexp.MustCompile(`client:\s*([^,]+)`).FindStringSubmatch(msg); len(cMatch) > 1 {
		clientIP = strings.TrimSpace(cMatch[1])
	}
	server := ""
	if sMatch := regexp.MustCompile(`server:\s*([^,]+)`).FindStringSubmatch(msg); len(sMatch) > 1 {
		server = strings.TrimSpace(sMatch[1])
	}

	return &ErrorLogEntry{
		Timestamp:    parsedTime,
		RawTimestamp: timeStr,
		LogLevel:     level,
		PID:          pid,
		Message:      msg,
		ClientIP:     clientIP,
		Server:       server,
		Raw:          line,
	}
}

// GetLogAnalytics computes full aggregated metrics across an access log
func (s *NginxService) GetLogAnalytics(filePath string) (*LogAnalytics, error) {
	if filePath == "" {
		filePath = filepath.Join(s.LogDir, "access.log")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open access log: %w", err)
	}
	defer file.Close()

	analytics := &LogAnalytics{
		MethodCounts:    make(map[string]int64),
		StatusCodeMap:   make(map[int]int64),
		HourlyBreakdown: []HourlyTraffic{},
	}

	ipMap := make(map[string]int64)
	pathMap := make(map[string]int64)
	uaMap := make(map[string]int64)
	hourlyMap := make(map[string]*HourlyTraffic)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		entry := s.parseAccessLine(line)
		if entry == nil {
			continue
		}

		analytics.TotalRequests++
		analytics.TotalBytes += entry.BodyBytes

		// Status codes
		analytics.StatusCodeMap[entry.StatusCode]++
		if entry.StatusCode >= 200 && entry.StatusCode < 300 {
			analytics.Status2xx++
		} else if entry.StatusCode >= 300 && entry.StatusCode < 400 {
			analytics.Status3xx++
		} else if entry.StatusCode >= 400 && entry.StatusCode < 500 {
			analytics.Status4xx++
		} else if entry.StatusCode >= 500 {
			analytics.Status5xx++
		}

		// Method
		if entry.Method != "" {
			analytics.MethodCounts[entry.Method]++
		}

		// Aggregates
		if entry.RemoteIP != "" {
			ipMap[entry.RemoteIP]++
		}
		if entry.Path != "" {
			pathMap[entry.Path]++
		}
		if entry.UserAgent != "" && entry.UserAgent != "-" {
			uaMap[entry.UserAgent]++
		}

		// Hourly breakdown
		if !entry.Timestamp.IsZero() {
			hourKey := entry.Timestamp.Format("2006-01-02 15:00")
			if _, exists := hourlyMap[hourKey]; !exists {
				hourlyMap[hourKey] = &HourlyTraffic{Hour: hourKey}
			}
			hourlyMap[hourKey].Requests++
			if entry.StatusCode >= 400 {
				hourlyMap[hourKey].Errors++
			}
		}
	}

	// Calculate Error Rate
	if analytics.TotalRequests > 0 {
		errorCount := analytics.Status4xx + analytics.Status5xx
		analytics.ErrorRate = float64(errorCount) / float64(analytics.TotalRequests) * 100.0
	}

	// Sort Top Items
	analytics.TopIPs = getTopItems(ipMap, 10)
	analytics.TopPaths = getTopItems(pathMap, 10)
	analytics.TopUserAgents = getTopItems(uaMap, 10)

	// Hourly breakdown sorted
	var hours []string
	for h := range hourlyMap {
		hours = append(hours, h)
	}
	sort.Strings(hours)
	for _, h := range hours {
		analytics.HourlyBreakdown = append(analytics.HourlyBreakdown, *hourlyMap[h])
	}

	return analytics, nil
}

func getTopItems(m map[string]int64, topN int) []ItemCount {
	var list []ItemCount
	for k, v := range m {
		list = append(list, ItemCount{Item: k, Count: v})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Count > list[j].Count
	})
	if len(list) > topN {
		list = list[:topN]
	}
	return list
}
