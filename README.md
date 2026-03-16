# Blackwater

A comprehensive server management and monitoring tool built with Go and Gin. Named after the iconic city in Red Dead Redemption (RDR1 & RDR2), Blackwater aims to provide a robust and organized way to manage your server infrastructure.

This application allows you to monitor hardware performance (CPU, GPU, RAM, Disk), manage system processes, and handle user authentication with role-based access control.

## 🚀 Features

- **Hardware Monitoring:** Real-time information about CPU, GPU, RAM, Disk, and Network usage.
- **Process Management:** View detailed information about running system processes, start new ones, and terminate existing ones.
- **Real-Time Monitoring (WebSockets):** Efficiently stream process updates to multiple clients using a centralized Hub pattern.
- **User Authentication:** Secure JWT-based login and registration.
- **Role-Based Access Control (RBAC):** Granular control over system features using roles and permissions.
- **Database Seeding:** Quick setup with initial roles, permissions, and default admin user.
- **Static File Serving:** Built-in support for handling file uploads.

## 🌐 Microservice Integration

Blackwater is designed to be language-agnostic. Since it provides a standard RESTful API secured by JWT, it can seamlessly integrate as a specialized microservice into any modern architecture.

### Real-Time Capabilities
The service includes a WebSocket Hub that fetches system data once and broadcasts it to all connected clients. This minimizes CPU overhead while providing live updates for dashboards and monitoring tools.

## 🛠️ Tech Stack

- **Language:** Go (1.25+)
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

4. **Seed the Database:**
   ```bash
   go run seeder.go
   ```

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
- `GET /cpu` - CPU usage statistics
- `GET /gpu` - GPU information
- `GET /ram` - Memory usage
- `GET /disk` - Disk space information
- `GET /network` - Network usage statistics

### Process Management (Requires Auth)
- `GET /processes` - List all running system processes
- `GET /process/single/:pid` - Detailed info for a specific process
- `GET /process/log` - History of started processes
- `POST /process/start` - Start a new process
- `DELETE /process/kill/:pid` - Terminate a process

### Real-Time Monitoring (WebSockets)
- `WS /ws/processes` - Live process stream (Broadcasts every 5s)
- `WS /ws/cpu-temperature` - Live CPU temperature stream (Broadcasts every 1s)

## 🏗️ Scalable Hub Architecture

To support thousands of concurrent users (e.g., 2000+), Blackwater implements a **Centralized Hub Pattern** for WebSockets:

- **Single Source of Truth:** System metrics (processes, temperature) are fetched and processed once by a background worker.
- **Efficient Broadcasting:** Instead of each connection polling the kernel independently, a single JSON payload is generated and broadcasted to all active subscribers.
- **Low Latency:** This approach reduces I/O wait times and CPU usage from $O(N)$ to $O(1)$ for data gathering, ensuring consistent performance regardless of the number of connected clients.

## 🛡️ Permissions

| Permission | Description |
|---|---|
| `read_processes` | List all running system processes |
| `start_process` | Initiate new processes on the server |
| `kill_process` | Terminate running processes |
| `read_cpu` | Access CPU usage and information |
| `read_ram` | Access RAM usage and information |
| `read_disk` | Access Disk usage and information |
| `read_network` | Access Network usage and information |

---
Developed by [Ahmed Farghly](https://github.com/ahmedfargh)
