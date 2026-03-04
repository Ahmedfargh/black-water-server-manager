# Blackwater

A comprehensive server management and monitoring tool built with Go and Gin. Named after the iconic city in Red Dead Redemption (RDR1 & RDR2), Blackwater aims to provide a robust and organized way to manage your server infrastructure.

This application allows you to monitor hardware performance (CPU, GPU, RAM, Disk), manage system processes, and handle user authentication with role-based access control.

## 🚀 Features

- **Hardware Monitoring:** Real-time information about CPU, GPU, RAM, and Disk usage.
- **Process Management:** View detailed information about running system processes.
- **User Authentication:** Secure JWT-based login and registration.
- **Role-Based Access Control (RBAC):** Manage users, roles, and permissions effectively.
- **Database Seeding:** Quick setup with initial roles and permissions.
- **Static File Serving:** Built-in support for handling file uploads.

## 🛠️ Tech Stack

- **Language:** Go (1.25+)
- **Web Framework:** [Gin](https://github.com/gin-gonic/gin)
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
   Copy the `.env.example` file to `.env` and update the values with your database credentials and JWT secret.
   ```bash
   cp .env.example .env
   ```
   Edit `.env`:
   ```env
   APP_PORT=":8080"
   DB_HOST=127.0.0.1
   DB_PORT=3306
   DB_NAME=go_server
   DB_USER=root
   DB_PASSWORD=your_password
   JWT_SECRET=your_jwt_secret
   ```

4. **Setup Database:**
   Create a MySQL database named `go_server` (or whatever you specified in `.env`).

5. **Seed the Database:**
   Run the seeder to initialize roles, permissions, and a default user.
   ```bash
   go run seeder.go
   ```

## 🚀 Usage

1. **Start the server:**
   ```bash
   go run main.go
   ```
   The server will start at `http://localhost:8080` (or your configured port).

2. **API Documentation:**
   The project includes Postman collections for testing the endpoints:
   - `Authentication.postman_collection.json`
   - `HardWare Info.postman_collection.json`
   - `Proccess Collection.postman_collection.json`
   - `User Mangement.postman_collection.json`

### Key Endpoints

#### Authentication
- `POST /login` - User login
- `POST /register` - User registration

#### Hardware Info (Requires Auth)
- `GET /info/cpu` - CPU usage statistics
- `GET /info/gpu` - GPU information
- `GET /info/ram` - Memory usage
- `GET /info/disk` - Disk space information

#### Process Management (Requires Auth)
- `GET /info/processes` - List all running system processes
- `GET /info/process/single/:pid` - Detailed info for a specific process
- `GET /info/process/log` - Paginated history of started processes (Query params: `page`, `pageSize`)
- `POST /info/process/start` - Start a new process (JSON body: `command`, `args`)

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. **Fork the project.**
2. **Create your feature branch:** `git checkout -b feature/AmazingFeature`
3. **Commit your changes:** `git commit -m 'Add some AmazingFeature'`
4. **Push to the branch:** `git push origin feature/AmazingFeature`
5. **Open a Pull Request.**

### Coding Standards
- Follow standard Go idioms and formatting (`go fmt`).
- Ensure all new features are documented.
- Add tests for new functionality where applicable.

## 📝 License

Distributed under the MIT License. See `LICENSE` for more information (if applicable).

---
Developed by [Ahmed Farghly](https://github.com/ahmedfargh)
