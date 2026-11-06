#!/usr/bin/env bash
set -euo pipefail

APP_DIR="/opt/hello-world"

echo "Removing hello-world application"

sudo systemctl stop hello-world || true
sudo systemctl disable hello-world || true
sudo rm -f /etc/systemd/system/hello-world.service
sudo rm -f /etc/default/hello-world
sudo rm -rf "${APP_DIR}"
sudo systemctl daemon-reload

echo "Remove completed"
