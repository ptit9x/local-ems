#!/usr/bin/env bash
# ============================================================
#  Local EMS — Linux Uninstaller
#  Run: sudo bash uninstall.sh
# ============================================================
set -e

SERVICE_NAME="ems"
INSTALL_DIR="/opt/ems"
SERVICE_USER="ems"

echo ""
echo "╔══════════════════════════════════════════════╗"
echo "║       Local EMS — Uninstaller                ║"
echo "╚══════════════════════════════════════════════╝"
echo ""

if [ "$(id -u)" -ne 0 ]; then
    echo "❌ Please run as root: sudo bash uninstall.sh"
    exit 1
fi

# --- Stop and disable service ---
if systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
    echo "⏹  Stopping EMS service..."
    systemctl stop "$SERVICE_NAME"
fi

if systemctl is-enabled --quiet "$SERVICE_NAME" 2>/dev/null; then
    echo "🔧 Disabling EMS service..."
    systemctl disable "$SERVICE_NAME"
fi

# --- Remove service file ---
if [ -f "/etc/systemd/system/${SERVICE_NAME}.service" ]; then
    echo "🗑  Removing systemd service..."
    rm -f "/etc/systemd/system/${SERVICE_NAME}.service"
    systemctl daemon-reload
fi

# --- Ask about data ---
if [ -d "$INSTALL_DIR/data" ]; then
    echo ""
    read -p "🗄  Delete database and historical data? [y/N] " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "🗑  Removing $INSTALL_DIR (including data)..."
        rm -rf "$INSTALL_DIR"
    else
        echo "🗑  Removing binaries and config (keeping data)..."
        rm -f "$INSTALL_DIR/ems-core"
        rm -f "$INSTALL_DIR/config.yaml"
        rm -f "$INSTALL_DIR/.env"
        echo "   Data preserved at: $INSTALL_DIR/data/"
    fi
else
    rm -rf "$INSTALL_DIR"
fi

# --- Remove user ---
if id "$SERVICE_USER" &>/dev/null; then
    echo "👤 Removing service user: $SERVICE_USER"
    userdel "$SERVICE_USER" 2>/dev/null || true
fi

echo ""
echo "✅ EMS uninstalled."
echo ""
