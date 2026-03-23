# Blackwater

A comprehensive server management and monitoring tool built with Go and Gin. Named after the iconic city in Red Dead Redemption (RDR1 & RDR2), Blackwater aims to provide a robust and organized way to manage your server infrastructure.

This application allows you to monitor hardware performance (CPU, GPU, RAM, Disk), manage system processes, and handle user authentication with role-based access control.

## 🚀 Features

- **Hardware Monitoring:** Real-time information about CPU, GPU, RAM, Disk, and Network usage.
- **Firewall Management:** Multi-distro support for Debian/Ubuntu (UFW), Arch Linux (UFW), and Red Hat-based distributions (Firewalld).
- **Audit Logging:** Automatically record system actions (firewall changes, etc.) with user attribution for security and accountability.
- **Docker Management:** View, inspect, and monitor Docker containers running on the host.
- **Process Management:** View detailed information about running system processes, start new ones, and terminate existing ones.
- **Process Ownership Tracking:** Automatically record which user started each process for accountability and logging.
- **System Audit Logging:** Track and persist administrative actions, such as Firewall state changes, for security and compliance.
- **Real-Time Monitoring (WebSockets):** Efficiently stream process updates, container metrics, and **live container logs** to multiple clients.
- **User Authentication:** Secure JWT-based login and registration.
- **Role-Based Access Control (RBAC):** Granular control over system features using roles and permissions.
- **Database Seeding:** Quick setup with initial roles, permissions, and default admin user.
- **Static File Serving:** Built-in support for handling file uploads.

## 🌐 Microservice Integration

Blackwater is designed to be language-agnostic. Since it provides a standard RESTful API secured by JWT, it can seamlessly integrate as a specialized microservice into any modern architecture.

### Real-Time Capabilities

The service includes multiple WebSocket Hubs that fetch system data once and broadcast it to all connected clients. This minimizes CPU overhead while providing live updates for dashboards and monitoring tools.

## 🛠️ Tech Stack

- **Language:** Go (1.24+)
- **Web Framework:** [Gin](https://github.com/gin-gonic/gin)
- **WebSocket:** [Gorilla WebSocket](https://github.com/gorilla/websocket)
- **ORM:** [GORM](https://gorm.io/)
- **Database:** MySQL
- **Monitoring:** [gopsutil](https://github.com/shirou/gopsutil)
- **Authentication:** JWT (JSON Web Tokens)

## 📋 Prerequisites

Before you begin, ensure you have the following installed:

- [Go](https://golang.org/doc/install) (version 1.25 or higher)
- [MySQL](https://dev.mysql.com/downloads/installer/)
- Git

## ⚙️ Installation

1. **Clone the repository:**

   ```bash
   git clone https://github.com/ahmedfargh/server-manager.git
   cd golang-server-controller
   ```

2. **Install dependencies:**

   ```bash
   go mod tidy
   ```

3. **Configure Environment Variables:**
   Copy the `.env.example` file to `.env` and update the values.

   ```bash
   cp .env.example .env
   ```

## 🐳 Docker Deployment (Zero-Config Testing)

For developers on **Mac, Windows, or Linux**, Docker provides an isolated environment to test Blackwater's API and Database orchestration immediately without manual configuration.

1. **Configure Environment:**
   Ensure your `.env` file has the correct `DB_PASSWORD` and `JWT_SECRET`.
   *Note: For Docker, we recommend `DB_PORT=3307` to avoid conflicts with local MySQL instances.*

2. **Start the Engine:**
   ```bash
   docker compose up --build -d
   ```

3. **Access the API:**
   The server will be available at `http://localhost:8080`.

> [!TIP]
> **Cross-Platform Compatibility:** While Docker allows you to test the API logic and project structure on Mac and Windows, please note that low-level hardware and firewall features require a native Linux host for full functionality.

## 🚀 Usage

1. **Start the server:**

   ```bash
   go run main.go
   ```

2. **API Documentation:**
   The project includes Postman collections for testing:
   - `Authentication.postman_collection.json`
   - `HardWare Info.postman_collection.json`
   - `Proccess Collection.postman_collection.json`
   - `User Mangement.postman_collection.json`
   - `RealTime.postman_collection.json` (WebSockets)

   ## Command-Line Interface (CLI)

   A companion CLI is available at `cmd/cli/bwcli` to provide quick, local interactions with the service.

   - **Build the CLI:**

      ```bash
      go build -o bwcli ./cmd/cli
      ```

   - **Run with Go:**

      ```bash
      go run ./cmd/cli
      ```

   - **List available commands:**

      ```bash
      ./bwcli --help
      ```

   Use `./bwcli --help` to discover available subcommands (for example: `auth`, `docker`, `system`).

## 🔗 Key Endpoints

### Authentication

- `POST /login` - User login
- `POST /register` - User registration

### User Management (Requires Auth)

- `GET /users/` - List users
- `POST /users/acount/update` - Update profile
- `GET /users/profile/me` - Get own profile
- `GET /users/crud/users/:id` - Get user by ID
- `GET /users/roles` - List all roles

### Hardware Info (Requires Auth)

- `GET /info/cpu` - CPU usage statistics
- `GET /info/gpu` - GPU information
- `GET /info/ram` - Memory usage
- `GET /info/disk` - Disk space information
- `GET /network` - Network usage statistics
- `GET /network/connections` - List of active network connections with process info

### Firewall Management (Requires Auth)

- `GET /firewall/status` - Get firewall status (UFW on Debian/Arch, Firewalld on Red Hat)
- `GET /firewall/enable` - Enable firewall
- `GET /firewall/disable` - Disable firewall
- `GET /firewall/rules` - List numbered/detailed firewall rules
- `GET /firewall/list` - List active firewall rules

### Audit Logs (Requires Auth)

- `GET /audit/list` - List system audit logs (supports `page`, `limit`, and `type` filters)

### Docker Management (Requires Auth)

- `GET /docker/containers` - List all containers running on the host
- `GET /docker/container/:id` - Get detailed information for a specific container
- `GET /docker/container/:id/status` - Get real-time container metrics (CPU %, Memory Usage/Limit, Network I/O, Block I/O, Pids)
- `POST /docker/container/:id/:action` - Perform an action on a container (`start`, `stop`, `restart`)

### Process Management (Requires Auth)

- `GET /info/processes` - List all running system processes
- `GET /info/process/single/:pid` - Detailed info for a specific process
- `GET /info/process/log` - History of started processes
- `POST /info/process/start` - Start a new process
- `DELETE /info/process/kill/:pid` - Terminate a process

### Real-Time Monitoring (WebSockets)

- `WS /ws/processes` - Live process stream (Broadcasts every 5s)
- `WS /ws/cpu-temperature` - Live CPU temperature stream (Broadcasts every 1s)
- `WS /ws/docker/:containerId` - Live container-specific metrics (CPU, Memory, Network, Block I/O)
- **`WS /ws/docker/:containerId/logs`** - Live real-time container log streaming (Follow mode)

## 🏗️ Dynamic Hub Architecture

To support large-scale monitoring without overwhelming the host, Blackwater implements a sophisticated WebSocket management system:

- **Centralized System Hubs:** System-wide metrics (processes, temperature) use a single source of truth fetched once and broadcasted to all subscribers.
- **On-Demand Container Hubs:** Monitoring for specific Docker containers and their logs is handled by dynamically created hubs.
  - **Resource Efficient:** A hub is only created when the first user starts monitoring a specific container.
  - **Automatic Lifecycle:** Hubs and their associated monitoring workers (status pollers and log streamers) are automatically destroyed when the last client disconnects, saving CPU and memory resources.
- **True Streaming:** Container logs use a direct stream from the Docker daemon (`Follow: true`), ensuring zero-latency updates for log dashboards.
- **$O(1)$ Efficiency:** For any number of clients monitoring the same resource, the system only performs one data-gathering operation, ensuring linear scalability.

## 🛡️ Permissions

| Permission             | Description                                |
| ---------------------- | ------------------------------------------ |
| `read_processes`       | List all running system processes          |
| `start_process`        | Initiate new processes on the server       |
| `kill_process`         | Terminate running processes                |
| `read_cpu`             | Access CPU usage and information           |
| `read_ram`             | Access RAM usage and information           |
| `read_disk`            | Access Disk usage and information          |
| `read_network`         | Access Network usage and information       |
| `view_firewall_status` | View the current status of the firewall    |
| `enable_firewall`      | Enable the system firewall (UFW/Firewalld) |
| `disable_firewall`     | Disable the system firewall                |
| `view_firewall_rules`  | List active and numbered firewall rules    |
| `read_containers`      | List Docker containers                     |
| `manage_containers`    | Start, stop, or restart Docker containers  |
| `view_audit_logs`      | Access and filter system audit logs        |

---

Developed by [Ahmed Farghly](https://github.com/ahmedfargh)
