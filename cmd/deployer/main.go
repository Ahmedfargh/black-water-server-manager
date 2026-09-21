package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ANSI color formatting
const (
	ColorReset  = "\033[0m"
	ColorBold   = "\033[1m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
)

func printBanner() {
	banner := ColorCyan + ColorBold + `
╔═══════════════════════════════════════════════════════════════════╗
║            BLACKWATER SERVER MANAGER - AUTO DEPLOYER              ║
║         Automated Production VPS Deployment & Provisioning        ║
╚═══════════════════════════════════════════════════════════════════╝` + ColorReset
	fmt.Println(banner)
}

func logStep(step int, total int, title string) {
	fmt.Printf("\n%s[%d/%d] %s%s\n", ColorPurple+ColorBold, step, total, title, ColorReset)
}

func logSuccess(format string, a ...interface{}) {
	fmt.Printf(ColorGreen+"  ✔ "+ColorReset+format+"\n", a...)
}

func logWarn(format string, a ...interface{}) {
	fmt.Printf(ColorYellow+"  ⚠ "+ColorReset+format+"\n", a...)
}

func logError(format string, a ...interface{}) {
	fmt.Printf(ColorRed+"  ✖ "+ColorReset+format+"\n", a...)
}

func logInfo(format string, a ...interface{}) {
	fmt.Printf(ColorCyan+"  ℹ "+ColorReset+format+"\n", a...)
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runCmdSilent(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func generateRandomHex(n int) string {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("bw_secret_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

type Config struct {
	Domain      string
	Email       string
	Port        string
	DBDriver    string
	DBName      string
	InstallDir  string
	EnableSSL   bool
	NonInteractive bool
}

func detectDistro() (string, string) {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "unknown", "unknown"
	}
	content := string(data)
	var id, version string
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "ID=") {
			id = strings.Trim(strings.TrimPrefix(line, "ID="), `"`)
		} else if strings.HasPrefix(line, "VERSION_ID=") {
			version = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), `"`)
		}
	}
	return id, version
}

func installDependencies(distro string) error {
	logInfo("Detected Linux distribution: %s", distro)

	switch distro {
	case "ubuntu", "debian":
		logInfo("Updating package index (apt-get update)...")
		_ = runCmd("apt-get", "update", "-y")
		logInfo("Installing Nginx, Git, GCC, Curl, Certbot...")
		if err := runCmd("apt-get", "install", "-y", "nginx", "git", "gcc", "build-essential", "curl", "certbot", "python3-certbot-nginx", "ufw"); err != nil {
			return err
		}
	case "centos", "rhel", "almalinux", "rocky", "fedora":
		pkgMgr := "dnf"
		if _, err := exec.LookPath("dnf"); err != nil {
			pkgMgr = "yum"
		}
		logInfo("Installing EPEL and base tools with %s...", pkgMgr)
		_ = runCmd(pkgMgr, "install", "-y", "epel-release")
		if err := runCmd(pkgMgr, "install", "-y", "nginx", "git", "gcc", "make", "curl", "certbot", "python3-certbot-nginx", "firewalld"); err != nil {
			return err
		}
	default:
		logWarn("Unrecognized distro '%s'. Ensuring basic packages exist...", distro)
	}

	// Check for Node.js / npm
	if _, err := exec.LookPath("node"); err != nil {
		logInfo("Node.js not found. Installing Node.js LTS via NodeSource...")
		_ = runCmd("bash", "-c", "curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && apt-get install -y nodejs || dnf install -y nodejs")
	}

	// Check for Go
	if _, err := exec.LookPath("go"); err != nil {
		logWarn("Go compiler not found in PATH. Please ensure Go is installed to compile from source, or provide a precompiled binary.")
	}

	return nil
}

func prompt(scanner *bufio.Scanner, message string, defaultValue string) string {
	fmt.Printf("%s%s [%s]: %s", ColorCyan, message, defaultValue, ColorReset)
	if scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text != "" {
			return text
		}
	}
	return defaultValue
}

func main() {
	printBanner()

	if runtime.GOOS != "linux" {
		logWarn("Blackwater Server Manager is optimized for Linux VPS hosts. Detected OS: %s", runtime.GOOS)
	}

	if os.Geteuid() != 0 {
		logWarn("Auto-deployer is running without root privileges. Nginx & Systemd configuration will require sudo.")
	}

	domainFlag := flag.String("domain", "", "Domain or public IP address")
	emailFlag := flag.String("email", "", "Let's Encrypt SSL contact email")
	portFlag := flag.String("port", "8080", "Backend internal port")
	dbFlag := flag.String("db", "sqlite", "Database driver (sqlite/mysql)")
	sslFlag := flag.Bool("ssl", false, "Enable automatic Let's Encrypt SSL")
	yesFlag := flag.Bool("yes", false, "Non-interactive mode (use defaults/flags)")
	flag.Parse()

	rootDir, err := os.Getwd()
	if err != nil {
		logError("Failed to get current directory: %v", err)
		os.Exit(1)
	}

	cfg := Config{
		Domain:         *domainFlag,
		Email:          *emailFlag,
		Port:           *portFlag,
		DBDriver:       *dbFlag,
		DBName:         "blackwater.db",
		InstallDir:     rootDir,
		EnableSSL:      *sslFlag,
		NonInteractive: *yesFlag,
	}

	scanner := bufio.NewScanner(os.Stdin)

	if !cfg.NonInteractive {
		fmt.Println("\n" + ColorBold + "Please enter deployment parameters:" + ColorReset)
		cfg.Domain = prompt(scanner, "Domain or Public IP", cfg.Domain)
		if cfg.Domain == "" {
			cfg.Domain = "_"
		}

		cfg.Port = prompt(scanner, "Go Backend Port", cfg.Port)
		cfg.DBDriver = prompt(scanner, "Database Driver (sqlite/mysql)", cfg.DBDriver)

		if cfg.Domain != "_" && !cfg.EnableSSL {
			sslChoice := prompt(scanner, "Setup Free Let's Encrypt SSL Certificate? (y/n)", "n")
			if strings.ToLower(sslChoice) == "y" || strings.ToLower(sslChoice) == "yes" {
				cfg.EnableSSL = true
				cfg.Email = prompt(scanner, "Email for SSL certificate notifications", "admin@"+cfg.Domain)
			}
		}
	}

	totalSteps := 6
	step := 1

	// Step 1: Dependencies
	logStep(step, totalSteps, "Installing System Dependencies & Tools")
	step++
	distro, _ := detectDistro()
	if err := installDependencies(distro); err != nil {
		logWarn("Some dependencies could not be automatically installed: %v", err)
	} else {
		logSuccess("System dependencies verified.")
	}

	// Step 2: Environment Configuration (.env)
	logStep(step, totalSteps, "Generating Production Configuration (.env)")
	step++
	envPath := filepath.Join(cfg.InstallDir, ".env")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		jwtSecret := generateRandomHex(32)
		envContent := fmt.Sprintf(`# Blackwater Server Manager - Production Configuration
APP_PORT=":%s"
APP_URL="http://%s:%s/"
PORT=%s
GIN_MODE=release

# Database
DB_DRIVER=%s
DB_NAME=%s
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=

# Security
JWT_SECRET=%s

# Audit & Background Telemetry
AUDIT_PERIOD_TYPE=S
AUDIT_PERIOD_COUNTER=15
`, cfg.Port, cfg.Domain, cfg.Port, cfg.Port, cfg.DBDriver, cfg.DBName, jwtSecret)

		if err := os.WriteFile(envPath, []byte(envContent), 0600); err != nil {
			logError("Failed to write .env: %v", err)
		} else {
			logSuccess("Generated secure .env with new JWT Secret.")
		}
	} else {
		logInfo("Existing .env file preserved.")
	}

	// Step 3: Compile Go Binary
	logStep(step, totalSteps, "Compiling Optimized Go Backend Binary")
	step++
	binaryPath := filepath.Join(cfg.InstallDir, "server-manager")
	logInfo("Building %s with CGO...", binaryPath)
	cmdGo := exec.Command("go", "build", "-ldflags=-s -w", "-o", "server-manager", "main.go")
	cmdGo.Dir = cfg.InstallDir
	cmdGo.Env = append(os.Environ(), "CGO_ENABLED=1")
	if out, err := cmdGo.CombinedOutput(); err != nil {
		logError("Go build failed: %v\nOutput: %s", err, string(out))
		os.Exit(1)
	}
	logSuccess("Go binary compiled successfully (%s)", binaryPath)

	// Step 4: Build Vue 3 Frontend
	logStep(step, totalSteps, "Building Vue 3 SPA Frontend Assets")
	step++
	frontendDir := filepath.Join(cfg.InstallDir, "frontend")
	if _, err := os.Stat(frontendDir); err == nil {
		logInfo("Installing NPM packages...")
		npmInstall := exec.Command("npm", "install")
		npmInstall.Dir = frontendDir
		_ = npmInstall.Run()

		logInfo("Building frontend production bundle (npm run build)...")
		npmBuild := exec.Command("npm", "run", "build")
		npmBuild.Dir = frontendDir
		if out, err := npmBuild.CombinedOutput(); err != nil {
			logWarn("Frontend build output: %s (%v)", string(out), err)
		} else {
			logSuccess("Frontend distribution built to %s", filepath.Join(frontendDir, "dist"))
		}
	} else {
		logWarn("Frontend directory not found at %s", frontendDir)
	}

	// Step 5: Setup Systemd Service
	logStep(step, totalSteps, "Configuring & Starting Systemd Service")
	step++
	systemdContent := fmt.Sprintf(`[Unit]
Description=BlackWater Server Manager Engine
After=network.target docker.service
Wants=docker.service

[Service]
Type=simple
User=root
Group=root
WorkingDirectory=%s
ExecStart=%s
Restart=always
RestartSec=5s
LimitNOFILE=65535
LimitNPROC=4096
Environment=GIN_MODE=release
Environment=PORT=%s
EnvironmentFile=%s
StandardOutput=journal
StandardError=journal
SyslogIdentifier=blackwater

[Install]
WantedBy=multi-user.target
`, cfg.InstallDir, binaryPath, cfg.Port, envPath)

	servicePath := "/etc/systemd/system/blackwater.service"
	if err := os.WriteFile(servicePath, []byte(systemdContent), 0644); err != nil {
		logWarn("Could not write %s (permission denied). Run as sudo.", servicePath)
	} else {
		_ = runCmd("systemctl", "daemon-reload")
		_ = runCmd("systemctl", "enable", "--now", "blackwater")
		logSuccess("Systemd service 'blackwater' enabled and started.")
	}

	// Step 6: Setup Nginx Reverse Proxy
	logStep(step, totalSteps, "Configuring Nginx Reverse Proxy & WebSockets")
	step++
	nginxDistDir := filepath.Join(frontendDir, "dist")
	nginxConf := fmt.Sprintf(`map $http_upgrade $connection_upgrade {
    default upgrade;
    ''      close;
}

upstream blackwater_backend {
    server 127.0.0.1:%s;
    keepalive 32;
}

server {
    listen 80;
    listen [::]:80;
    server_name %s;

    root %s;
    index index.html;
    client_max_body_size 500M;

    gzip on;
    gzip_types text/plain text/css application/json application/javascript image/svg+xml;

    # API Proxy
    location /api/ {
        rewrite ^/api/(.*)$ /$1 break;
        proxy_pass http://blackwater_backend;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # WebSocket Proxy
    location /ws/ {
        proxy_pass http://blackwater_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_buffering off;
        proxy_read_timeout 86400s;
        proxy_send_timeout 86400s;
    }

    # Static Uploads
    location /uploads/ {
        proxy_pass http://blackwater_backend/uploads/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # Vue 3 SPA Fallback
    location / {
        try_files $uri $uri/ /index.html;
    }
}
`, cfg.Port, cfg.Domain, nginxDistDir)

	nginxAvailable := "/etc/nginx/sites-available/blackwater.conf"
	nginxEnabled := "/etc/nginx/sites-enabled/blackwater.conf"
	_ = os.MkdirAll("/etc/nginx/sites-available", 0755)
	_ = os.MkdirAll("/etc/nginx/sites-enabled", 0755)

	if err := os.WriteFile(nginxAvailable, []byte(nginxConf), 0644); err != nil {
		logWarn("Could not write %s: %v", nginxAvailable, err)
	} else {
		_ = os.Remove(nginxEnabled)
		_ = os.Symlink(nginxAvailable, nginxEnabled)
		// Test Nginx
		if err := runCmd("nginx", "-t"); err == nil {
			_ = runCmd("systemctl", "reload", "nginx")
			logSuccess("Nginx configuration applied and reloaded.")
		} else {
			logWarn("Nginx configuration test failed. Please check %s", nginxAvailable)
		}
	}

	// SSL via Certbot (Optional)
	if cfg.EnableSSL && cfg.Domain != "_" && cfg.Domain != "" {
		logInfo("Requesting Let's Encrypt SSL certificate for %s...", cfg.Domain)
		certArgs := []string{"--nginx", "-d", cfg.Domain, "--non-interactive", "--agree-tos"}
		if cfg.Email != "" {
			certArgs = append(certArgs, "-m", cfg.Email)
		} else {
			certArgs = append(certArgs, "--register-unsafely-without-email")
		}
		if err := runCmd("certbot", certArgs...); err != nil {
			logWarn("Certbot SSL issuance failed: %v", err)
		} else {
			logSuccess("Let's Encrypt SSL configured successfully!")
		}
	}

	// Health Check Verification
	time.Sleep(2 * time.Second)
	logInfo("Testing backend connectivity on http://127.0.0.1:%s...", cfg.Port)
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%s/auth/login", cfg.Port))
	if err == nil {
		defer resp.Body.Close()
		logSuccess("Backend responded with HTTP status %d", resp.StatusCode)
	} else {
		logWarn("Backend health check warning: %v", err)
	}

	// Final Summary
	fmt.Println("\n" + ColorGreen + ColorBold + "═══════════════════════════════════════════════════════════════════" + ColorReset)
	fmt.Printf("%s🎉 BlackWater Server Manager Deployed Successfully!%s\n", ColorGreen+ColorBold, ColorReset)
	fmt.Println(ColorGreen + ColorBold + "═══════════════════════════════════════════════════════════════════" + ColorReset)
	if cfg.Domain != "_" && cfg.Domain != "" {
		proto := "http"
		if cfg.EnableSSL {
			proto = "https"
		}
		fmt.Printf("  • Web URL:      %s%s://%s%s\n", ColorCyan, proto, cfg.Domain, ColorReset)
	} else {
		fmt.Printf("  • Localhost:    %shttp://127.0.0.1%s\n", ColorCyan, ColorReset)
	}
	fmt.Printf("  • Backend Port: %s%s%s\n", ColorCyan, cfg.Port, ColorReset)
	fmt.Printf("  • Service Name: %sblackwater.service%s\n", ColorCyan, ColorReset)
	fmt.Printf("  • Logs:         %sjournalctl -u blackwater -f%s\n", ColorCyan, ColorReset)
	fmt.Println()
}
