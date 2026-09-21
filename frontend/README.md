# Blackwater Frontend (Vue 3 + Vite)

High-contrast, industrial outlaw-tech dashboard for **Blackwater Server Manager**.

## 🛠️ Tech Stack
- **Framework:** Vue 3 (Composition API `<script setup>`)
- **Build Tool:** Vite
- **Routing:** Vue Router (HTML5 History Mode)
- **State Management:** Pinia
- **Styling:** Custom Scoped CSS & Outlaw Design System (Crimson, Charcoal, Amber, Brass)
- **Networking:** Axios (`/api`) & Native WebSockets (`/ws`)
- **Localization:** Vue I18n (Multi-language & RTL support)

---

## ⚡ Development Setup

### 1. Environment Configuration
Copy the example environment file:
```bash
cp .env.example .env
```
Default parameters in `.env`:
```env
VITE_BACKEND_URL=http://localhost:8080
VITE_PORT=5173
```

### 2. Install Dependencies
```bash
npm install
```

### 3. Start Development Server
```bash
npm run dev
```
Dashboard will be available at `http://localhost:5173`. Requests to `/api` and `/ws` will be proxied automatically to `http://localhost:8080` via Vite dev proxy.

---

## 🏗️ Production Build

To build the static distribution for Nginx or production hosting:
```bash
npm run build
```
The compiled SPA bundle will be placed in the `dist/` directory.

---

## 🌐 Production Nginx Routing

When serving `dist/` via Nginx, ensure SPA history mode fallback is enabled:
```nginx
location / {
    root /var/www/black-water-server-manager/frontend/dist;
    try_files $uri $uri/ /index.html;
}
```
