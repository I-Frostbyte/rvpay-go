#!/usr/bin/env bash
# Prepares an Ubuntu OCI Always Free A1 VM for this lightweight Docker Compose deployment.
set -euo pipefail

if [[ "${EUID}" -ne 0 ]]; then
    echo "Run this script as root (for example: sudo ./deploy-oci-bootstrap.sh)." >&2
    exit 1
fi

apt-get update
apt-get install -y ca-certificates curl git gnupg jq unzip ufw

install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
chmod a+r /etc/apt/keyrings/docker.gpg
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo \"$VERSION_CODENAME\") stable" | \
  tee /etc/apt/sources.list.d/docker.list > /dev/null

apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

ufw allow OpenSSH
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

systemctl enable --now docker
echo "OCI host bootstrap complete. Configure OCI Security List ingress for TCP 80 and 443, then deploy the stack."
