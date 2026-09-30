#!/usr/bin/env bash
# ============================================================
#  Local EMS — Linux Installer
#  Run: sudo bash install.sh
# ============================================================
set -e

APP_NAME="ems-core"
INSTALL_DIR="/opt/ems"
SERVICE_USER="ems"
SERVICE_NAME="ems"

echo ""
echo "╔══════════════════════════════════════════════╗"
echo "║       Local EMS — Linux Installer            ║"
echo "╚══════════════════════════════════════════════╝"
echo ""

# --- Check root ---
if [ "$(id -u)" -ne 0 ]; then
    echo "❌ Please run as root: sudo bash install.sh"
    exit 1
fi

# --- Check files exist ---
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BINARY="$SCRIPT_DIR/$APP_NAME"

if [ ! -f "$BINARY" ]; then
    echo "❌ Binary not found: $BINARY"
    echo "   Make sure install.sh is in the same folder as $APP_NAME"
    exit 1
fi

if [ ! -f "$SCRIPT_DIR/config.yaml" ]; then
    echo "❌ config.yaml not found in $SCRIPT_DIR"
    exit 1
fi

if [ ! -f "$SCRIPT_DIR/.env" ]; then
    echo "❌ .env not found in $SCRIPT_DIR"
    exit 1
fi

# --- Stop existing service if running ---
if systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
    echo "⏹  Stopping existing EMS service..."
    systemctl stop "$SERVICE_NAME"
fi

# --- Create service user ---
if ! id "$SERVICE_USER" &>/dev/null; then
    echo "👤 Creating service user: $SERVICE_USER"
    useradd -r -s /sbin/nologin -d "$INSTALL_DIR" "$SERVICE_USER"
fi

# --- Install files ---
echo "📁 Installing to $INSTALL_DIR..."
mkdir -p "$INSTALL_DIR/data"
cp "$BINARY" "$INSTALL_DIR/$APP_NAME"
cp "$SCRIPT_DIR/config.yaml" "$INSTALL_DIR/config.yaml"
cp "$SCRIPT_DIR/.env" "$INSTALL_DIR/.env"
chmod +x "$INSTALL_DIR/$APP_NAME"
chown -R "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR"

# --- Create systemd service ---
echo "⚙️  Creating systemd service..."
cat > /etc/systemd/system/${SERVICE_NAME}.service <<EOF
[Unit]
Description=Local Energy Management System
After=network.target
StartLimitIntervalSec=60
StartLimitBurst=5

[Service]
Type=simple
User=${SERVICE_USER}
Group=${SERVICE_USER}
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/${APP_NAME} --config ${INSTALL_DIR}/config.yaml
Restart=always
RestartSec=5
LimitNOFILE=65536

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=${SERVICE_NAME}

[Install]
WantedBy=multi-user.target
EOF

# --- Enable and start ---
echo "🚀 Starting EMS service..."
systemctl daemon-reload
systemctl enable "$SERVICE_NAME"
systemctl start "$SERVICE_NAME"

# --- Show status ---
sleep 2
echo ""
echo "════════════════════════════════════════════════"

if systemctl is-active --quiet "$SERVICE_NAME"; then
    # Read UI port from .env
    UI_PORT=$(grep -oP 'EMS_UI_PORT=\K\d+' "$INSTALL_DIR/.env" 2>/dev/null || echo "8080")
    echo "  ✅ EMS installed and running!"
    echo ""
    echo "  Dashboard:  http://$(hostname -I | awk '{print $1}'):${UI_PORT}"
    echo "  Install dir: $INSTALL_DIR"
    echo "  Config:      $INSTALL_DIR/.env"
    echo ""
    echo "  Commands:"
    echo "    sudo systemctl status ems    — check status"
    echo "    sudo systemctl restart ems   — restart"
    echo "    sudo systemctl stop ems      — stop"
    echo "    sudo journalctl -u ems -f    — view logs"
else
    echo "  ⚠️  Service created but may not be running."
    echo "  Check: sudo systemctl status ems"
    echo "  Logs:  sudo journalctl -u ems -n 50"
fi
echo "════════════════════════════════════════════════"
echo ""
