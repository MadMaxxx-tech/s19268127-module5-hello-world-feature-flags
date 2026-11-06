#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

: "${VERSION:?VERSION is required}"
: "${NEXUS_URL:?NEXUS_URL is required}"
: "${NEXUS_USER:?NEXUS_USER is required}"
: "${NEXUS_PASSWORD:?NEXUS_PASSWORD is required}"
: "${NEXUS_REPO:?NEXUS_REPO is required}"
: "${FEATURE_IP:?FEATURE_IP is required}"
: "${FEATURE_COUNTRY:?FEATURE_COUNTRY is required}"
: "${FEATURE_FORM:?FEATURE_FORM is required}"

APP_DIR="/opt/hello-world"
ARCHIVE_NAME="hello-world_${VERSION}.tar.gz"
ARCHIVE_PATH="${APP_DIR}/${ARCHIVE_NAME}"
BINARY_NAME="hello-world_${VERSION}"
CURRENT_BINARY="${APP_DIR}/hello-world"
BACKUP_BINARY="${APP_DIR}/hello-world_bkp"

echo "Deploying hello-world version ${VERSION}"

sudo mkdir -p "${APP_DIR}"

# Формируем аргументы feature flags
HELLO_WORLD_ARGS=""

if [ "${FEATURE_IP}" = "true" ]; then
  HELLO_WORLD_ARGS="${HELLO_WORLD_ARGS} -ip"
fi

if [ "${FEATURE_COUNTRY}" = "true" ]; then
  HELLO_WORLD_ARGS="${HELLO_WORLD_ARGS} -country"
fi

if [ "${FEATURE_FORM}" = "true" ]; then
  HELLO_WORLD_ARGS="${HELLO_WORLD_ARGS} -form"
fi

# Сохраняем аргументы для systemd
printf 'HELLO_WORLD_ARGS="%s"\n' "${HELLO_WORLD_ARGS}" | sudo tee /etc/default/hello-world > /dev/null

# Устанавливаем systemd unit
sudo cp "${SCRIPT_DIR}/hello-world.service" /etc/systemd/system/hello-world.service

# Скачиваем архив из Nexus
curl -f --connect-timeout 10 --max-time 300 \
  -u "${NEXUS_USER}:${NEXUS_PASSWORD}" \
  -o "${ARCHIVE_PATH}" \
  "${NEXUS_URL}/repository/${NEXUS_REPO}/${VERSION}/${ARCHIVE_NAME}"

# Распаковываем архив
tar -xzf "${ARCHIVE_PATH}" -C "${APP_DIR}"

# Делаем backup предыдущей версии
if [ -f "${CURRENT_BINARY}" ]; then
  mv "${CURRENT_BINARY}" "${BACKUP_BINARY}"
fi

# Активируем новую версию под стабильным именем
mv "${APP_DIR}/${BINARY_NAME}" "${CURRENT_BINARY}"
chmod +x "${CURRENT_BINARY}"

# Удаляем временный архив
rm -f "${ARCHIVE_PATH}"

# Применяем unit и перезапускаем сервис
sudo systemctl daemon-reload
sudo systemctl enable hello-world
sudo systemctl restart hello-world

# Проверяем работоспособность
sleep 2
curl -fsS http://127.0.0.1:8080/

echo
echo "Deploy completed: hello-world ${VERSION}"
echo "Feature flags: IP=${FEATURE_IP}, Country=${FEATURE_COUNTRY}, Form=${FEATURE_FORM}"
