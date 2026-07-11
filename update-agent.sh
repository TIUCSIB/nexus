#!/bin/bash
# Nexus Agent 自动更新脚本
# Usage: bash <(curl -fsSL https://raw.githubusercontent.com/TIUCSIB/Nexus/main/update-agent.sh)

set -e

red='\033[0;31m'
green='\033[0;32m'
yellow='\033[0;33m'
cyan='\033[0;36m'
plain='\033[0m'

AGENT_DIR="/opt/nexus-agent"
AGENT_BIN="${AGENT_DIR}/nexus-agent"
GITHUB_REPO="TIUCSIB/Nexus"

echo -e "${cyan}========================================${plain}"
echo -e "${cyan}    Nexus Agent 自动更新                ${plain}"
echo -e "${cyan}========================================${plain}"
echo ""

# 检测架构
arch=$(uname -m)
if [[ $arch == "x86_64" || $arch == "x64" || $arch == "amd64" ]]; then
    arch="amd64"
elif [[ $arch == "aarch64" || $arch == "arm64" ]]; then
    arch="arm64"
else
    echo -e "${red}不支持的架构: ${arch}${plain}"
    exit 1
fi

echo -e "  架构: ${green}${arch}${plain}"

# 获取最新版本
echo -e "${yellow}[1/8]${plain} 获取最新版本..."
latest_version=$(curl -Ls "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [[ -z "${latest_version}" ]]; then
    echo -e "${red}无法获取最新版本，使用 latest 链接${plain}"
    download_url="https://github.com/${GITHUB_REPO}/releases/latest/download/nexus-agent-linux-${arch}"
else
    echo -e "  最新版本: ${green}${latest_version}${plain}"
    download_url="https://github.com/${GITHUB_REPO}/releases/download/${latest_version}/nexus-agent-linux-${arch}"
fi

# 下载新版本
echo -e "${yellow}[2/8]${plain} 下载新版本..."
wget --no-check-certificate -q --show-progress -O "${AGENT_BIN}.new" "${download_url}" 2>&1 || {
    echo -e "${red}下载失败${plain}"
    exit 1
}
chmod +x "${AGENT_BIN}.new"

# 停止 Agent 服务
echo -e "${yellow}[3/8]${plain} 停止 Agent 服务..."
if command -v rc-service &>/dev/null; then
    rc-service nexus-agent stop 2>/dev/null || true
elif command -v systemctl &>/dev/null; then
    systemctl stop nexus-agent 2>/dev/null || true
else
    pkill -f nexus-agent || true
fi
sleep 2

# 备份旧版本
echo -e "${yellow}[4/8]${plain} 备份旧版本..."
if [ -f "${AGENT_BIN}" ]; then
    backup_file="${AGENT_BIN}.old.$(date +%Y%m%d_%H%M%S)"
    cp "${AGENT_BIN}" "${backup_file}"
    echo -e "  已备份到 ${backup_file}"
fi

# 替换新版本
echo -e "${yellow}[5/8]${plain} 替换新版本..."
mv "${AGENT_BIN}.new" "${AGENT_BIN}"
chmod +x "${AGENT_BIN}"

# 清理旧配置（让 Agent 重新生成）
echo -e "${yellow}[6/8]${plain} 清理旧配置..."
cd "${AGENT_DIR}"
if [ -f "singbox-1.json" ]; then
    mv singbox-1.json "singbox-1.json.old.$(date +%Y%m%d_%H%M%S)"
    echo -e "  已备份旧配置"
fi

# 启动 Agent
echo -e "${yellow}[7/8]${plain} 启动 Agent..."

# 修复 OpenRC/systemd 工作目录与参数，避免相对路径写配置失败
if [[ -f /etc/init.d/nexus-agent ]]; then
    cat > /etc/init.d/nexus-agent << EOF
#!/sbin/openrc-run
name="nexus-agent"
command="${AGENT_DIR}/nexus-agent"
command_args="-c ${AGENT_DIR}/agent.yaml"
command_background="yes"
directory="${AGENT_DIR}"
pidfile="/run/nexus-agent.pid"
depend() { need net; }
EOF
    chmod +x /etc/init.d/nexus-agent
    echo -e "  已修复 OpenRC 服务定义"
fi
if [[ -f /etc/systemd/system/nexus-agent.service ]]; then
    cat > /etc/systemd/system/nexus-agent.service << EOF
[Unit]
Description=Nexus Agent
After=network.target

[Service]
Type=simple
WorkingDirectory=${AGENT_DIR}
ExecStart=${AGENT_DIR}/nexus-agent -c ${AGENT_DIR}/agent.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
    systemctl daemon-reload 2>/dev/null || true
    echo -e "  已修复 systemd 服务定义"
fi

# 确保 agent.yaml 中 working_dir 为绝对路径
if [[ -f "${AGENT_DIR}/agent.yaml" ]]; then
    sed -i "s|working_dir: \.|working_dir: ${AGENT_DIR}|g" "${AGENT_DIR}/agent.yaml"
    sed -i "s|working_dir: \"\.\"|working_dir: \"${AGENT_DIR}\"|g" "${AGENT_DIR}/agent.yaml"
fi

if command -v rc-service &>/dev/null; then
    rc-service nexus-agent start
elif command -v systemctl &>/dev/null; then
    systemctl start nexus-agent
else
    nohup ${AGENT_BIN} -c ${AGENT_DIR}/agent.yaml > nexus-agent.log 2>&1 &
    echo "  Agent PID: $!"
fi

# 等待启动
echo -e "${yellow}[8/8]${plain} 等待服务启动..."
sleep 5

# 检查状态
echo ""
echo -e "${green}========================================${plain}"
echo -e "${green}    检查服务状态                        ${plain}"
echo -e "${green}========================================${plain}"
echo ""

# 检查 sing-box 进程
if ps aux | grep -v grep | grep sing-box > /dev/null; then
    echo -e "  ${green}✓ sing-box 正在运行${plain}"
else
    echo -e "  ${red}✗ sing-box 未运行${plain}"
    echo -e "  ${yellow}提示: 可能需要等待几秒，或检查配置${plain}"
fi

# 检查端口监听
if netstat -tlnp 2>/dev/null | grep -E ':(20442|9090)' > /dev/null; then
    echo -e "  ${green}✓ 端口正常监听${plain}"
    netstat -tlnp 2>/dev/null | grep -E ':(20442|9090)' | head -3
else
    echo -e "  ${yellow}⚠ 端口监听检查失败${plain}"
fi

echo ""
echo -e "${green}更新完成！${plain}"
echo ""
echo -e "查看日志："
echo -e "  tail -f ${AGENT_DIR}/*.log"
echo ""
echo -e "查看配置："
echo -e "  cat ${AGENT_DIR}/singbox-1.json"
echo ""
echo -e "如果有问题，回滚到旧版本："
echo -e "  rc-service nexus-agent stop"
echo -e "  cp ${AGENT_BIN}.old.* ${AGENT_BIN}"
echo -e "  rc-service nexus-agent start"
echo ""
