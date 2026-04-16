# Blackwater

> [!WARNING]
> **FRONTEND STATUS:** The included Vue.js frontend is provided for **testing and demonstration purposes only**. It serves as a reference interface to interact with the Blackwater API and is not intended for production environments.

A comprehensive server management and monitoring tool built with Go and Gin. Named after the iconic city in Red Dead Redemption (RDR1 & RDR2), Blackwater aims to provide a robust and organized way to manage your server infrastructure.

This application allows you to monitor hardware performance (CPU, GPU, RAM, Disk), manage system processes, and handle user authentication with role-based access control.

## 🚀 Features

- **Hardware Monitoring:** Real-time information about CPU, GPU, RAM, Disk, and Network usage.
- **Historical Hardware Reports:** Analyze detailed historical performance data for CPU, Memory, and Disk usage over time with average usage insights and interactive trends.
- **Firewall Management:** Multi-distro support for Debian/Ubuntu (UFW), Arch Linux (UFW), and Red Hat-based distributions (Firewalld).
- **Audit Logging:** Automatically record system actions (firewall changes, etc.) with user attribution for security and accountability.
- **Docker Auto-Heal & Resource Monitoring:** Automatically monitor container resource usage (CPU/Memory) and health status (e.g., detecting stopped containers), triggering automated actions (Restart, Start, Stop, etc.) when thresholds are exceeded or abnormalities are detected.
- **Docker Notifications:** Receive instant alerts via Discord, Telegram, or Custom Webhooks when containers start, stop, or restart, or when resource thresholds are exceeded.
- **Docker Management & Auto-Discovery:** Automatically discover and persist running containers on the host, monitor their metrics, and stream live logs.
- **Resource Limits & Automated Actions:** Define CPU and Memory consumption thresholds for containers with automated response actions (Stop, Restart, etc.).
- **Process Management:** View detailed information about running system processes, start new ones, and terminate existing ones.
- **Site Health Monitoring:** Monitor external sites' availability and performance, logging status history (UP, Redirection, Not Found, Server Error).
- **Process Ownership Tracking:** Automatically record which user started each process for accountability and logging.
- **System Audit Logging:** Track and persist administrative actions, such as Firewall state changes, for security and compliance.
- **Interactive System Terminal (New):** Execute host machine shell commands securely via a deeply integrated, high-performance web terminal interface natively connected to the backend.
- **Real-Time Monitoring (WebSockets):** Efficiently stream process updates, container metrics, and **live container logs** to multiple clients.
- **Background Synchronization:** A background manager periodically (every 10s) synchronizes the state of all containers on the host with the database.
- **User Authentication:** Secure JWT-based login and registration.
- **Role-Based Access Control (RBAC):** Granular control over system features using roles and permissions.
- **Database Seeding:** Quick setup with initial roles, permissions, and default admin user.
- **Static File Serving:** Built-in support for handling file uploads.

## 🌐 Microservice Integration

Blackwater is designed to be language-agnostic. Since it provides a standard RESTful API secured by JWT, it can seamlessly integrate as a specialized microservice into any modern architecture.

### Real-Time Capabilities

The service includes multiple WebSocket Hubs that fetch system data once and broadcast it to all connected clients. This minimizes CPU overhead while providing live updates for dashboards and monitoring tools.

### 🐳 Docker Auto-Discovery & Persistence

Blackwater implements a background **Docker Manager** that automatically detects all running containers on the host system upon startup and continues to monitor them.

- **Automatic Registration:** Any new container running on the server is automatically discovered and persisted in the database.
- **State Synchronization:** The system periodically (every 10 seconds) syncs container metadata (Status, Image, Ports, Command) between the Docker Engine and the local database.
- **Policy Management:** Users can define resource limits (CPU/Memory) and automated actions (Restart/Stop) for each container, which the manager can enforce.

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

   ### 🗄️ Database Setup
   Blackwater supports both **MySQL** and **SQLite**. This flexibility allows the system to run efficiently on everything from high-performance clusters to **low-end servers and machines** with limited resources.

   You can choose your preferred driver in the `.env` file:

   - **For MySQL:**
     ```env
     DB_DRIVER=mysql
     DB_HOST=127.0.0.1
     DB_PORT=3306
     DB_NAME=go_server
     DB_USER=root
     DB_PASSWORD=your_password
     ```

   - **For SQLite (Recommended for Low-End Systems):**
     ```env
     DB_DRIVER=sqlite
     DB_NAME=blackwater
     ```
     *Note: SQLite is zero-config, requires no separate database server, and is ideal for environments with restricted CPU or memory.*

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
- `POST /hardware-report/by-time-range` - Detailed historical usage reports for a specific range
- `POST /hardware-report/average-usage-by-time-range` - Average usage metrics for a specific range
- `GET /hardware-report/latest` - Get the most recent report entry

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
- `POST /docker/container` - Register/Create a new container management record
- `PUT /docker/container/:id` - Update an existing container management record (e.g., update limits/policies)
- `DELETE /docker/container/:id` - Remove a container management record

### Process Management (Requires Auth)

- `GET /info/processes` - List all running system processes
- `GET /info/process/single/:pid` - Detailed info for a specific process
- `GET /info/process/log` - History of started processes
- `POST /info/process/start` - Start a new process
- `DELETE /info/process/kill/:pid` - Terminate a process

### Site Health Monitoring (Requires Auth)

- `POST /site/create` - Add a new site for health monitoring
- `GET /site/list` - List all monitored sites with their status
- `GET /site/full-checkup` - Trigger an immediate health check for all sites
 - `GET /site/health-status/:site_id` - Get paginated health records for a site. Supports query parameters `page` (default `1`) and `limit` (default `10`). You can also filter by date range using `start_date` and `end_date` in `YYYY-MM-DD` format. Requires authentication and `site_read` permission.
 - `GET /site/status-report/:site_id` - Get a status report for a site (aggregated results). Optional query parameters: `start_date` and `end_date` (`YYYY-MM-DD`). Requires authentication and `site_read` permission.

### Real-Time Monitoring (WebSockets)

- `WS /ws/processes` - Live process stream (Broadcasts every 5s)
- `WS /ws/cpu-temperature` - Live CPU temperature stream (Broadcasts every 1s)
- `WS /ws/docker/:containerId` - Live container-specific metrics (CPU, Memory, Network, Block I/O)
- **`WS /ws/docker/:containerId/logs`** - Live real-time container log streaming (Follow mode)
- **`WS /ws/terminal`** - Interactive host system shell terminal allowing bidirectional secure command execution.

## 🔔 Notification System

Blackwater includes a robust notification system that keeps you informed about critical server events, specifically focusing on Docker container status and resource health.

### 🛡️ Docker Notifications

Notifications are triggered by several events:
- **Container Lifecycle:** Alerts when a container is started, stopped, restarted, or removed.
- **Auto-Heal Actions:** Notifications when the system automatically takes action (e.g., restarting a container) due to high CPU or memory consumption.
- **Abnormalities:** Alerts when resource thresholds (CPU/Memory) are exceeded, even if no automated action is configured.

### ⚙️ Supported Drivers

Blackwater supports **Telegram**, **Discord**, and **Custom Webhooks** for real-time notifications. Each user can configure their own preferred notification driver and credentials:

#### 🔹 Telegram Configuration
1. **Bot Token:** Obtain a bot token from [@BotFather](https://t.me/botfather).
2. **Chat ID:** Find your Telegram Chat ID (you can use [@userinfobot](https://t.me/userinfobot)).

#### 🔹 Discord Configuration
1. **Bot Token:** Create a bot on the [Discord Developer Portal](https://discord.com/developers/applications) and obtain its token.
2. **Channel ID:** Right-click the desired channel in Discord and select "Copy Channel ID" (requires Developer Mode enabled).

#### 🔹 Webhook Configuration
1. **Target URL:** Your external endpoint URL that will receive POST requests containing the event data.
2. **Webhook Secret:** An optional secret key.

### 🛠️ Updating Settings

Settings can be seamlessly updated directly from the **Profile Configuration panel in the Vue.js Frontend** or via the API by sending a `POST` request to `/users/users/notifications/settings`:

**Request Body:**
```json
{
    "notification_driver": "Webhook", // or "Telegram", "Discord"
    "telegram_bot_token": "your_telegram_bot_token",
    "telegram_chat_id": "your_telegram_chat_id",
    "discord_bot_token": "your_discord_bot_token",
    "discord_channel_id": "your_discord_channel_id",
    "webhook_url": "https://your.endpoint.com/webhook",
    "webhook_secret": "your_optional_secret"
}
```

These settings are stored per-user, ensuring that team members receive alerts on their own preferred platforms and accounts.

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
| `site_create`          | Create new sites for health monitoring     |
| `site_read`            | View monitored sites and health checkups   |

---

Developed by [Ahmed Farghly](https://github.com/ahmedfargh)
