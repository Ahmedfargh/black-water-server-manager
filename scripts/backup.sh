#!/usr/bin/env bash
# ==============================================================================
# Blackwater Server Manager - Automated Backup Script
# ==============================================================================

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKUP_DIR="${PROJECT_DIR}/backups"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
ARCHIVE_NAME="blackwater_backup_${TIMESTAMP}.tar.gz"

echo -e "${CYAN}${BOLD}[BACKUP] Initiating Blackwater snapshot at ${TIMESTAMP}...${NC}"

mkdir -p "${BACKUP_DIR}"
TEMP_BACKUP="${BACKUP_DIR}/temp_${TIMESTAMP}"
mkdir -p "${TEMP_BACKUP}"

cd "${PROJECT_DIR}"

# 1. Backup .env configuration
if [ -f .env ]; then
    cp .env "${TEMP_BACKUP}/.env"
    echo -e "${GREEN}  ✔ Copied .env configuration${NC}"
fi

# 2. Backup SQLite Database (if present)
if [ -f blackwater.db ]; then
    cp blackwater.db "${TEMP_BACKUP}/blackwater.db"
    echo -e "${GREEN}  ✔ Copied SQLite database (blackwater.db)${NC}"
fi

# 3. Compress into .tar.gz
tar -czf "${BACKUP_DIR}/${ARCHIVE_NAME}" -C "${TEMP_BACKUP}" .
rm -rf "${TEMP_BACKUP}"

echo -e "${GREEN}${BOLD}✔ Backup created successfully: ${BACKUP_DIR}/${ARCHIVE_NAME}${NC}"
