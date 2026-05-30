#!/bin/bash
set -e

echo "=== CVSE xTower 首次部署安装脚本 ==="

# --- 参数 ---
INSTALL_PATH="${1:-/opt/cvse}"
DOMAIN="${2:-cvse.xtower.site}"

echo "目标路径: $INSTALL_PATH"
echo "域名: $DOMAIN"

# --- 创建目录 ---
mkdir -p "$INSTALL_PATH/build"

# --- 安装 Nginx（如果未安装） ---
if ! command -v nginx &>/dev/null; then
    echo ">>> 安装 Nginx..."
    apt-get update -qq
    apt-get install -y -qq nginx
fi

# --- 配置 Nginx ---
echo ">>> 配置 Nginx..."
cp "$INSTALL_PATH/deploy/nginx.conf" /etc/nginx/sites-available/"$DOMAIN"
# 替换 server_name
sed -i "s/server_name cvse.xtower.site;/server_name $DOMAIN;/" /etc/nginx/sites-available/"$DOMAIN"
# 替换 root
sed -i "s|root /opt/cvse/build;|root $INSTALL_PATH/build;|" /etc/nginx/sites-available/"$DOMAIN"

# 启用站点
if [ ! -L /etc/nginx/sites-enabled/"$DOMAIN" ]; then
    ln -s /etc/nginx/sites-available/"$DOMAIN" /etc/nginx/sites-enabled/
fi
rm -f /etc/nginx/sites-enabled/default

# 测试配置
nginx -t
systemctl restart nginx

# --- 注册 systemd 服务 ---
echo ">>> 注册 cvse-api 服务..."
cp "$INSTALL_PATH/deploy/cvse-api.service" /etc/systemd/system/cvse-api.service
sed -i "s|WorkingDirectory=/opt/cvse|WorkingDirectory=$INSTALL_PATH|" /etc/systemd/system/cvse-api.service
sed -i "s|ExecStart=/opt/cvse/cvse-api|ExecStart=$INSTALL_PATH/cvse-api|" /etc/systemd/system/cvse-api.service

systemctl daemon-reload
systemctl enable cvse-api
systemctl restart cvse-api

# --- 状态检查 ---
echo ""
echo "=== 安装完成 ==="
echo "Nginx:   $(systemctl is-active nginx)"
echo "API:     $(systemctl is-active cvse-api)"
echo "前端:    $INSTALL_PATH/build"
echo "API 地址: http://127.0.0.1:8080/api/health"
echo ""

# 检查 API 是否响应
sleep 1
if curl -sf http://127.0.0.1:8080/api/health > /dev/null 2>&1; then
    echo "✅ API 运行正常"
else
    echo "⚠️  API 未响应，请检查: systemctl status cvse-api"
fi
