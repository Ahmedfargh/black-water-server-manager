# Blackwater

A comprehensive server management and monitoring tool built with Go and Gin. Named after the iconic city in Red Dead Redemption (RDR1 & RDR2), Blackwater aims to provide a robust and organized way to manage your server infrastructure.

This application allows you to monitor hardware performance (CPU, GPU, RAM, Disk), manage system processes, and handle user authentication with role-based access control.

## 🚀 Features

- **Hardware Monitoring:** Real-time information about CPU, GPU, RAM, and Disk usage.
- **Process Management:** View detailed information about running system processes, start new ones, and terminate existing ones.
- **User Authentication:** Secure JWT-based login and registration.
- **Role-Based Access Control (RBAC):** Granular control over system features using roles and permissions.
- **Database Seeding:** Quick setup with initial roles, permissions, and default admin user.
- **Static File Serving:** Built-in support for handling file uploads.

## 🌐 Microservice Integration

Blackwater is designed to be language-agnostic. Since it provides a standard RESTful API secured by JWT, it can seamlessly integrate as a specialized microservice into any modern architecture. This allows you to offload low-level system monitoring and process management to a dedicated Go-powered service while maintaining your primary application logic in your preferred framework.

### How to Integrate
Your main application (regardless of language) can interact with Blackwater by:
1. **Authentication:** Authenticate your service or users via the `/login` endpoint to receive a JWT token.
2. **RESTful Communication:** Use standard HTTP clients to consume hardware and process data.
3. **Cross-Framework Examples:**
   - **PHP (Laravel):** Use `Guzzle` or the `Http` facade to monitor server health from your dashboard.
   - **Python (Django/FastAPI):** Use `requests` or `httpx` to trigger and manage background system processes.
   - **Node.js (Express/NestJS):** Use `axios` or `fetch` for real-time hardware monitoring.
   - **C# (.NET):** Use `HttpClient` to integrate deep server control into enterprise Windows or Linux applications.
   - **Java (Spring Boot):** Use `RestTemplate` or `WebClient` for system-level resource orchestration.

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

## 🛡️ RBAC & Permissions

The system uses a flexible Role-Based Access Control (RBAC) system. Each user is assigned a Role, and each Role has a set of Permissions. Permissions can also be assigned directly to a User.

### Available Permissions

| Permission | Description |
|---|---|
| `create_user` | Create new users |
| `read_user` | View user details and list users |
| `update_user` | Update existing users |
| `delete_user` | Remove users from the system |
| `manage_roles` | Create, read, and update roles |
| `manage_permissions` | Manage system-wide permissions |
| `read_processes` | List all running system processes |
| `read_process` | View detailed info for a single process |
| `start_process` | Initiate new processes on the server |
| `read_process_log` | Access the log of started processes |
| `kill_process` | Terminate running processes |
| `read_cpu` | Access CPU usage and information |
| `read_gpu` | Access GPU usage and information |
| `read_ram` | Access RAM usage and information |
| `read_disk` | Access Disk usage and information |

## 🔗 Key Endpoints

### Authentication
- `POST /login` - User login
- `POST /register` - User registration

### User Management (Requires Auth)
- `GET /users/` - List users (`read_user`)
- `POST /users/acount/update` - Update own profile
- `GET /users/profile/me` - Get own profile details
- `GET /users/crud/users/:id` - Get user by ID (`read_user`)
- `POST /users/crud/users/` - Create user (`create_user`)
- `PUT /users/crud/users/:id` - Update user (`update_user`)
- `DELETE /users/crud/users/:id` - Delete user (`delete_user`)
- `GET /users/crud/users/list` - List all users (`read_user`)

### Role Management (Requires Auth)
- `GET /users/roles` - List all roles (`manage_roles`)
- `POST /users/roles` - Create new role (`manage_roles`)
- `GET /users/role/:id` - Get role by ID (`manage_roles`)
- `POST /users/roles/update/:id` - Update role (`manage_roles`)

### Hardware Info (Requires Auth)
- `GET /info/cpu` - CPU usage statistics (`read_cpu`)
- `GET /info/gpu` - GPU information (`read_gpu`)
- `GET /info/ram` - Memory usage (`read_ram`)
- `GET /info/disk` - Disk space information (`read_disk`)

### Process Management (Requires Auth)
- `GET /info/processes` - List all running system processes (`read_processes`)
- `GET /info/process/single/:pid` - Detailed info for a specific process (`read_process`)
- `GET /info/process/log` - Paginated history of started processes (`read_process_log`)
- `POST /info/process/start` - Start a new process (`start_process`)
- `DELETE /info/process/kill/:pid` - Terminate a process (`kill_process`)

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
