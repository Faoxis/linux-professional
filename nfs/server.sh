#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <CLIENT_IP/CIDR>"
  echo "Example: $0 10.0.0.9/32"
  exit 1
fi

CLIENT_CIDR="$1"
EXPORT_DIR="/srv/share"
UPLOAD_DIR="${EXPORT_DIR}/upload"
EXPORTS_FILE="/etc/exports"

echo "[1/6] Installing NFS server packages..."
sudo apt update
sudo apt install -y nfs-kernel-server

echo "[2/6] Checking listening ports (2049, 111) ..."
sudo ss -tnplu | egrep ':(2049|111)\b' || true

echo "[3/6] Creating export directories..."
sudo mkdir -p "$UPLOAD_DIR"

echo "[4/6] Setting permissions..."
sudo chown -R nobody:nogroup "$EXPORT_DIR"
sudo chmod 0777 "$UPLOAD_DIR"

echo "[5/6] Writing ${EXPORTS_FILE} ..."
# Overwrite /etc/exports with exactly one export rule.
cat <<EOF | sudo tee "$EXPORTS_FILE" > /dev/null
${EXPORT_DIR} ${CLIENT_CIDR}(rw,sync,root_squash)
EOF

echo "[6/6] Applying exports and showing result..."
sudo exportfs -r
sudo exportfs -s

echo "Done."
