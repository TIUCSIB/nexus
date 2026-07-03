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
    case \ in
        --panel)  PANEL_URL="\"; shift 2 ;;
        --token)  TOKEN="\"; shift 2 ;;
        --name)   NODE_NAME="\"; shift 2 ;;
        --port)   STATS_PORT="\"; shift 2 ;;
        --dir)    INSTALL_DIR="\"; shift 2 ;;
        *)        echo "Unknown: \"; exit 1 ;;
    esac
done

if [[ -z "\" || -z "\" ]]; then
    echo "Usage: bash install.sh --panel https://panel.com --token TOKEN [--name NAME]"
    exit 1
fi

echo "=== Nexus Agent Installer ==="
echo "Panel: \"
echo "Name:  \"
echo ""

ARCH=
case \ in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    *)       echo "Unsupported: \"; exit 1 ;;
esac

OS=
BINARY_URL="\/downloads/nexus-agent-\-\"

mkdir -p "\"

echo "Downloading..."
if command -v wget &> /dev/null; then
    wget -q "\" -O "\/\"
elif command -v curl &> /dev/null; then
    curl -fsSL "\" -o "\/\"
else
    echo "Error: wget or curl required"; exit 1
fi

chmod +x "\/\"

cat > /etc/systemd/system/nexus-agent.service << EOF
[Unit]
Description=Nexus Agent
After=network.target
[Service]
Type=simple
ExecStart=\/\ --panel \ --token \ --name \
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
