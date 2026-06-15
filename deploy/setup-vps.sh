#!/bin/bash
# Первоначальная настройка VPS 188.244.115.135
# Запустить один раз: bash setup-vps.sh

set -e

echo "=== Установка Docker ==="
apt-get update -y
apt-get install -y ca-certificates curl gnupg

install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
chmod a+r /etc/apt/keyrings/docker.gpg

echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
  https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" \
  | tee /etc/apt/sources.list.d/docker.list > /dev/null

apt-get update -y
apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

systemctl enable docker
systemctl start docker

echo "=== Создание директории проекта ==="
mkdir -p /opt/lims-blockchain

echo "=== Настройка .env на VPS ==="
cat > /opt/lims-blockchain/.env << 'EOF'
DOCKERHUB_USERNAME=ЗАМЕНИТЕ_НА_ВАШ_ЛОГИН
CONTRACT_ADDRESS=
EOF

echo ""
echo "=== ГОТОВО ==="
echo "Отредактируйте /opt/lims-blockchain/.env"
echo "Затем запустите:"
echo "  cd /opt/lims-blockchain"
echo "  docker compose up -d"
