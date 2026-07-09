#!/bin/bash
# Nexus Agent Installer
# Usage: bash install.sh --panel https://panel.com --token YOUR_TOKEN --name my-node
set -e

PANEL_URL=""
TOKEN=""
NODE_NAME="node-1"
STATS_PORT=9090
INSTALL_DIR="/opt/nexus-agent"
BINARY_NAME="nexus-agent"

while [[ $# -gt 0 ]]; do
    case $1 in
        --panel)  PANEL_URL="$2"; shift 2 ;;
        --token)  TOKEN="$2"; shift 2 ;;
        --name)   NODE_NAME="$2"; shift 2 ;;
        --port)   STATS_PORT="$2"; shift 2 ;;
        --dir)    INSTALL_DIR="$2"; shift 2 ;;
        *)        echo "Unknown: $1"; exit 1 ;;
    esac
done

if [[ -z "$PANEL_URL" || -z "$TOKEN" ]]; then
    echo "Usage: bash install.sh --panel https://panel.com --token TOKEN [--name NAME]"
    exit 1
fi

echo "=== Nexus Agent Installer ==="
echo "Panel: $PANEL_URL"
echo "Name:  $NODE_NAME"
echo ""

ARCH=
case $(uname -m) in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    *)       echo "Unsupported: $(uname -m)"; exit 1 ;;
esac

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
BINARY_URL="$PANEL_URL/downloads/nexus-agent-$OS-$ARCH"

mkdir -p "$INSTALL_DIR"

echo "Downloading $BINARY_URL ..."
if command -v wget &> /dev/null; then
    wget -q "$BINARY_URL" -O "$INSTALL_DIR/$BINARY_NAME"
elif command -v curl &> /dev/null; then
    curl -fsSL "$BINARY_URL" -o "$INSTALL_DIR/$BINARY_NAME"
else
    echo "Error: wget or curl required"; exit 1
fi

chmod +x "$INSTALL_DIR/$BINARY_NAME"

cat > /etc/systemd/system/nexus-agent.service << EOF
[Unit]
Description=Nexus Agent
After=network.target
[Service]
Type=simple
ExecStart=$INSTALL_DIR/$BINARY_NAME --panel $PANEL_URL --token $TOKEN --name $NODE_NAME
Restart=always
RestartSec=5
[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable nexus-agent
systemctl restart nexus-agent

echo ""
echo "=== Done ==="
echo "Status: systemctl status nexus-agent"
echo "Logs:   journalctl -u nexus-agent -f"