package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ANSI Color Codes for beautiful terminal output
const (
	ColorReset  = "\033[0m"
	ColorBold   = "\033[1m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

func logSetup(format string, a ...interface{}) {
	fmt.Printf(ColorGreen+"[SETUP] "+ColorReset+format+"\n", a...)
}

func logWarn(format string, a ...interface{}) {
	fmt.Printf(ColorYellow+"[WARN]  "+ColorReset+format+"\n", a...)
}

func logError(format string, a ...interface{}) {
	fmt.Printf(ColorRed+"[ERROR] "+ColorReset+format+"\n", a...)
}

func logInfo(format string, a ...interface{}) {
	fmt.Printf(ColorCyan+"[INFO]  "+ColorReset+format+"\n", a...)
}

func printBanner() {
	banner := ColorCyan + ColorBold + `
╔═══════════════════════════════════════════════════════════════════╗
║                   BLACKWATER SERVER MANAGER                       ║
║            Unified Backend & Frontend Setup & Runner              ║
╚═══════════════════════════════════════════════════════════════════╝` + ColorReset
	fmt.Println(banner)
}

func generateRandomSecret(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "blackwater_jwt_secret_key_" + fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

func checkCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func ensureEnvFiles(dbDriver string, appPort string, frontPort string) error {
	rootDir, err := os.Getwd()
	if err != nil {
		return err
	}

	// 1. Root .env
	rootEnvPath := filepath.Join(rootDir, ".env")
	if _, err := os.Stat(rootEnvPath); os.IsNotExist(err) {
		logSetup("Root .env not found. Creating a new one...")
		examplePath := filepath.Join(rootDir, ".env.example")
		var content string
		if exampleBytes, err := os.ReadFile(examplePath); err == nil {
			content = string(exampleBytes)
		} else {
			content = `APP_PORT=":8080"
APP_URL="http://localhost:8080/"
DB_DRIVER=sqlite
DB_NAME=blackwater.db
JWT_SECRET=your_jwt_secret
AUDIT_PERIOD_TYPE=S
AUDIT_PERIOD_COUNTER=10
`
		}

		// Inject driver & secret if missing or placeholder
		jwtSecret := generateRandomSecret(32)
		if strings.Contains(content, "JWT_SECRET=your_jwt_secret") || strings.Contains(content, "JWT_SECRET=") {
			content = strings.Replace(content, "JWT_SECRET=your_jwt_secret", "JWT_SECRET="+jwtSecret, 1)
		}
		if !strings.Contains(content, "DB_DRIVER=") {
			content = "DB_DRIVER=" + dbDriver + "\n" + content
		} else {
			// Update DB_DRIVER value
			lines := strings.Split(content, "\n")
			for i, line := range lines {
				if strings.HasPrefix(line, "DB_DRIVER=") {
					lines[i] = "DB_DRIVER=" + dbDriver
				}
			}
			content = strings.Join(lines, "\n")
		}

		if err := os.WriteFile(rootEnvPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write root .env: %w", err)
		}
		logSetup("Created root .env (DB_DRIVER=%s, Generated secure JWT_SECRET)", dbDriver)
	} else {
		logInfo("Root .env already exists.")
	}

	// 2. Frontend .env
	frontEnvPath := filepath.Join(rootDir, "frontend", ".env")
	if _, err := os.Stat(frontEnvPath); os.IsNotExist(err) {
		logSetup("frontend/.env not found. Creating a new one...")
		frontExamplePath := filepath.Join(rootDir, "frontend", ".env.example")
		var frontContent string
		if exampleBytes, err := os.ReadFile(frontExamplePath); err == nil {
			frontContent = string(exampleBytes)
		} else {
			frontContent = fmt.Sprintf("VITE_BACKEND_URL=http://localhost:%s\nVITE_PORT=%s\n", strings.TrimPrefix(appPort, ":"), frontPort)
		}

		if err := os.WriteFile(frontEnvPath, []byte(frontContent), 0644); err != nil {
			return fmt.Errorf("failed to write frontend/.env: %w", err)
		}
		logSetup("Created frontend/.env (VITE_BACKEND_URL=http://localhost:%s, VITE_PORT=%s)", strings.TrimPrefix(appPort, ":"), frontPort)
	} else {
		logInfo("frontend/.env already exists.")
	}

	return nil
}

func setupGoDependencies() error {
	logSetup("Checking Go dependencies (go mod download)...")
	cmd := exec.Command("go", "mod", "download")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func seedDatabase() error {
	logSetup("Running Database Auto-Migration and Seeders...")
	cmd := exec.Command("go", "run", "seeder.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("database seeding failed: %w", err)
	}
	logSetup("Database successfully migrated & seeded!")
	return nil
}

func setupFrontendDependencies() error {
	frontDir := filepath.Join(".", "frontend")
	nodeModulesDir := filepath.Join(frontDir, "node_modules")

	if _, err := os.Stat(nodeModulesDir); err == nil {
		logInfo("frontend/node_modules already exists. Skipping npm install (use -force-npm to reinstall).")
		return nil
	}

	logSetup("Installing frontend npm dependencies (this may take a moment)...")
	cmd := exec.Command("npm", "install")
	cmd.Dir = frontDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npm install failed: %w", err)
	}
	logSetup("Frontend dependencies successfully installed!")
	return nil
}

func buildProjects() error {
	logSetup("Building Go Backend (binary: server-manager)...")
	backendCmd := exec.Command("go", "build", "-o", "server-manager", "main.go")
	backendCmd.Stdout = os.Stdout
	backendCmd.Stderr = os.Stderr
	if err := backendCmd.Run(); err != nil {
		return fmt.Errorf("backend build failed: %w", err)
	}
	logSetup("Go Backend binary created: ./server-manager")

	logSetup("Building Frontend bundle (Vite build)...")
	frontendCmd := exec.Command("npm", "run", "build")
	frontendCmd.Dir = filepath.Join(".", "frontend")
	frontendCmd.Stdout = os.Stdout
	frontendCmd.Stderr = os.Stderr
	if err := frontendCmd.Run(); err != nil {
		return fmt.Errorf("frontend build failed: %w", err)
	}
	logSetup("Frontend production build created in ./frontend/dist/")
	return nil
}

// pipeStream reads lines from an io.ReadCloser and outputs them with a colored prefix
func pipeStream(reader io.ReadCloser, prefix string, color string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fmt.Printf("%s%s%s %s\n", color, prefix, ColorReset, scanner.Text())
	}
}

func runConcurrently(appPort, frontPort string) error {
	printBanner()
	logInfo("Starting Backend and Frontend concurrently...")
	logInfo("Press Ctrl+C to terminate both servers.")

	fmt.Println(ColorGreen + ColorBold + `
===================================================================
  🌐 Frontend UI:     http://localhost:` + frontPort + `
  🔌 Backend API:    http://localhost:` + strings.TrimPrefix(appPort, ":") + `
  👤 Default User:   admin / password (or admin@example.com)
===================================================================
` + ColorReset)

	// Context / Cancellation via signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Command 1: Backend
	var backendCmd *exec.Cmd
	if _, err := os.Stat("./server-manager"); err == nil {
		backendCmd = exec.Command("./server-manager")
	} else {
		backendCmd = exec.Command("go", "run", "main.go")
	}

	backendStdout, err := backendCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("backend stdout pipe error: %w", err)
	}
	backendStderr, err := backendCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("backend stderr pipe error: %w", err)
	}

	// Command 2: Frontend
	frontendCmd := exec.Command("npm", "run", "dev")
	frontendCmd.Dir = filepath.Join(".", "frontend")

	frontendStdout, err := frontendCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("frontend stdout pipe error: %w", err)
	}
	frontendStderr, err := frontendCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("frontend stderr pipe error: %w", err)
	}

	// Start servers
	if err := backendCmd.Start(); err != nil {
		return fmt.Errorf("failed to start Go backend: %w", err)
	}
	if err := frontendCmd.Start(); err != nil {
		_ = backendCmd.Process.Kill()
		return fmt.Errorf("failed to start Frontend dev server: %w", err)
	}

	go pipeStream(backendStdout, "[BACKEND] ", ColorBlue)
	go pipeStream(backendStderr, "[BACKEND] ", ColorBlue)
	go pipeStream(frontendStdout, "[FRONTEND]", ColorPurple)
	go pipeStream(frontendStderr, "[FRONTEND]", ColorPurple)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_ = backendCmd.Wait()
		logWarn("Backend server exited.")
	}()

	go func() {
		defer wg.Done()
		_ = frontendCmd.Wait()
		logWarn("Frontend server exited.")
	}()

	// Wait for termination signal
	sig := <-sigChan
	fmt.Printf("\n%s[SHUTDOWN] Received signal [%v]. Stopping services cleanly...%s\n", ColorYellow, sig, ColorReset)

	// Terminate processes
	if backendCmd.Process != nil {
		_ = backendCmd.Process.Signal(syscall.SIGTERM)
	}
	if frontendCmd.Process != nil {
		_ = frontendCmd.Process.Signal(syscall.SIGTERM)
	}

	// Give a grace period of 2 seconds, then kill if still running
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logSetup("All services stopped successfully. Goodbye!")
	case <-time.After(3 * time.Second):
		logWarn("Forcing processes to terminate...")
		if backendCmd.Process != nil {
			_ = backendCmd.Process.Kill()
		}
		if frontendCmd.Process != nil {
			_ = frontendCmd.Process.Kill()
		}
		logSetup("Cleanup complete.")
	}

	return nil
}

func main() {
	mode := flag.String("mode", "all", "Execution mode: 'all' (setup + run dev), 'dev' (run only), 'install' (env & deps only), 'seed' (run db seeders), 'build' (compile binary & frontend)")
	dbDriver := flag.String("db", "sqlite", "Database driver: 'sqlite' or 'mysql'")
	appPort := flag.String("port", "8080", "Backend HTTP port")
	frontPort := flag.String("front-port", "5173", "Frontend HTTP port")
	skipNpm := flag.Bool("skip-npm", false, "Skip npm install step")
	skipSeed := flag.Bool("skip-seed", false, "Skip database seeding step")
	forceNpm := flag.Bool("force-npm", false, "Force npm install even if node_modules exists")
	flag.Parse()

	printBanner()
	logInfo("Operating System: %s (%s)", runtime.GOOS, runtime.GOARCH)

	// 1. Toolchain verification
	goVer, err := checkCommand("go", "version")
	if err != nil {
		logError("Go is not installed or not in PATH!")
		os.Exit(1)
	}
	logInfo("Go Version: %s", goVer)

	nodeVer, err := checkCommand("node", "-v")
	if err != nil {
		logWarn("Node.js is not installed or not in PATH! Frontend actions might fail.")
	} else {
		npmVer, _ := checkCommand("npm", "-v")
		logInfo("Node.js: %s (npm %s)", nodeVer, npmVer)
	}

	// Mode handling
	switch *mode {
	case "install":
		if err := ensureEnvFiles(*dbDriver, *appPort, *frontPort); err != nil {
			logError("Environment error: %v", err)
			os.Exit(1)
		}
		if err := setupGoDependencies(); err != nil {
			logError("Go dependencies error: %v", err)
			os.Exit(1)
		}
		if !*skipNpm {
			if *forceNpm {
				_ = os.RemoveAll(filepath.Join(".", "frontend", "node_modules"))
			}
			if err := setupFrontendDependencies(); err != nil {
				logError("Frontend dependencies error: %v", err)
				os.Exit(1)
			}
		}
		logSetup("Setup complete! You can now start the services with: go run setup.go -mode=dev")

	case "seed":
		if err := ensureEnvFiles(*dbDriver, *appPort, *frontPort); err != nil {
			logError("Environment error: %v", err)
			os.Exit(1)
		}
		if err := seedDatabase(); err != nil {
			logError("Seeding error: %v", err)
			os.Exit(1)
		}

	case "build":
		if err := ensureEnvFiles(*dbDriver, *appPort, *frontPort); err != nil {
			logError("Environment error: %v", err)
			os.Exit(1)
		}
		if err := buildProjects(); err != nil {
			logError("Build error: %v", err)
			os.Exit(1)
		}
		logSetup("Build completed successfully!")

	case "dev":
		if err := runConcurrently(*appPort, *frontPort); err != nil {
			logError("Server error: %v", err)
			os.Exit(1)
		}

	case "all":
		// 1. Env files
		if err := ensureEnvFiles(*dbDriver, *appPort, *frontPort); err != nil {
			logError("Environment error: %v", err)
			os.Exit(1)
		}
		// 2. Go deps
		if err := setupGoDependencies(); err != nil {
			logError("Go dependencies error: %v", err)
			os.Exit(1)
		}
		// 3. Frontend deps
		if !*skipNpm {
			if *forceNpm {
				_ = os.RemoveAll(filepath.Join(".", "frontend", "node_modules"))
			}
			if err := setupFrontendDependencies(); err != nil {
				logError("Frontend dependencies error: %v", err)
				os.Exit(1)
			}
		}
		// 4. Seeding if DB doesn't exist or not skipped
		if !*skipSeed {
			if _, err := os.Stat("blackwater.db"); os.IsNotExist(err) || *dbDriver != "sqlite" {
				if err := seedDatabase(); err != nil {
					logWarn("Initial database seeding warning: %v", err)
				}
			} else {
				logInfo("SQLite database 'blackwater.db' found. (Run with -mode=seed to re-seed)")
			}
		}
		// 5. Run dev servers concurrently
		if err := runConcurrently(*appPort, *frontPort); err != nil {
			logError("Server error: %v", err)
			os.Exit(1)
		}

	default:
		logError("Unknown mode '%s'. Available modes: all, dev, install, seed, build", *mode)
		flag.Usage()
		os.Exit(1)
	}
}
