#!/usr/bin/env bash
# ==============================================================================
# Blackwater Server Manager - Production Update & Zero-Downtime Rebuild Script
# ==============================================================================

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

echo -e "${CYAN}${BOLD}"
echo "╔═══════════════════════════════════════════════════════════════════╗"
echo "║             BLACKWATER SERVER MANAGER - UPDATE ENGINE             ║"
echo "║          Pulling Latest Commits, Building & Reloading             ║"
echo "╚═══════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}[ERROR] Please run this update script as root or with sudo.${NC}"
    exit 1
fi

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_DIR"

echo -e "${CYAN}[1/5] Pulling latest changes from Git repository...${NC}"
git pull origin main || git pull

echo -e "${CYAN}[2/5] Updating Go backend dependencies & rebuilding binary...${NC}"
go mod download
go build -o server-manager main.go

echo -e "${CYAN}[3/5] Running Database Migrations & Seeders...${NC}"
go run seeder.go || true

echo -e "${CYAN}[4/5] Building Vue 3 Frontend production SPA...${NC}"
if [ -d "frontend" ]; then
    cd frontend
    if command -v npm >/dev/null 2>&1; then
        npm install --silent
        npm run build
    else
        echo -e "${YELLOW}[WARN] npm not found; skipping frontend build.${NC}"
    fi
    cd "$PROJECT_DIR"
fi

echo -e "${CYAN}[5/5] Restarting Blackwater Systemd daemon...${NC}"
if systemctl is-active --quiet blackwater; then
    systemctl restart blackwater
    echo -e "${GREEN}${BOLD}✔ Blackwater service restarted successfully!${NC}"
else
    echo -e "${YELLOW}[INFO] blackwater.service not currently active. Start it with: sudo systemctl start blackwater${NC}"
fi

echo -e "\n${GREEN}${BOLD}===================================================================${NC}"
echo -e "${GREEN}${BOLD}           ✔ BLACKWATER UPDATE COMPLETED SUCCESSFULLY!             ${NC}"
echo -e "${GREEN}${BOLD}===================================================================${NC}\n"
