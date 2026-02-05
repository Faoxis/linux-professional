#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <NFS_SERVER_IP>"
  echo "Example: $0 10.0.0.8"
  exit 1
fi

SERVER_IP="$1"
MOUNT_POINT="/mnt"
REMOTE_EXPORT="/srv/share"
FSTAB_LINE="${SERVER_IP}:${REMOTE_EXPORT} ${MOUNT_POINT} nfs vers=3,noauto,x-systemd.automount,_netdev 0 0"

echo "[1/6] Installing NFS client packages..."
sudo apt update
sudo apt install -y nfs-common

echo "[2/6] Ensuring mount point exists: ${MOUNT_POINT}"
sudo mkdir -p "${MOUNT_POINT}"

echo "[3/6] Adding /etc/fstab entry (idempotent)..."
# Remove previous entries for the same mountpoint/export (best-effort), then add ours once.
sudo sed -i "\|^[[:space:]]*${SERVER_IP}:${REMOTE_EXPORT}[[:space:]]\+${MOUNT_POINT}[[:space:]]\+nfs\b|d" /etc/fstab
sudo sed -i "\|[[:space:]]${MOUNT_POINT}[[:space:]]\+nfs\b.*x-systemd\.automount|d" /etc/fstab
echo "${FSTAB_LINE}" | sudo tee -a /etc/fstab > /dev/null

echo "[4/6] Reloading systemd units..."
sudo systemctl daemon-reload

echo "[5/6] Restarting remote-fs.target..."
sudo systemctl restart remote-fs.target

echo "[6/6] Triggering automount (first access to ${MOUNT_POINT})..."
sudo ls -la "${MOUNT_POINT}" > /dev/null || true

echo "Mount status (expect autofs + nfs lines if success):"
mount | grep " ${MOUNT_POINT} " || true

echo "Done."
