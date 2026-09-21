#!/usr/bin/env bash
# ==============================================================================
# Blackwater Server Manager - Automated VPS Deployment Script
# ==============================================================================

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

echo -e "${CYAN}${BOLD}"
echo "╔═══════════════════════════════════════════════════════════════════╗"
echo "║          BLACKWATER SERVER MANAGER - VPS AUTO DEPLOYER            ║"
echo "╚═══════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# 1. Check Root Privileges
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}[ERROR] Please run this script as root or with sudo.${NC}"
    exit 1
fi

PROJECT_DIR=$(pwd)
echo -e "${CYAN}[INFO] Working directory: ${PROJECT_DIR}${NC}"

# 2. Compile and Launch the Go Deployer
echo -e "${CYAN}[INFO] Compiling Blackwater Auto-Deployer engine...${NC}"
if command -v go >/dev/null 2>&1; then
    go run cmd/deployer/main.go "$@"
else
    echo -e "${YELLOW}[WARN] Go compiler not found. Installing Go directly...${NC}"
    if [ -f /etc/debian_version ]; then
        apt-get update -y && apt-get install -y golang-go
    elif [ -f /etc/redhat-release ]; then
        dnf install -y golang || yum install -y golang
    fi
    go run cmd/deployer/main.go "$@"
fi
DB_CONNECTION=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=centeral_clinic_system
DB_USERNAME=clinic_user
DB_PASSWORD=ClinicSaas@2026!
