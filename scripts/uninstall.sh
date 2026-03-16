#!/usr/bin/env bash
set -euo pipefail

INSTALL_ROOT="/opt/unisub"
CONFIG_ROOT="/etc/unisub"
STATE_ROOT="/var/lib/unisub"
SERVICE_NAME="unisub"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
PURGE_CONFIG=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --purge-config)
      PURGE_CONFIG=1
      shift
      ;;
    *)
      echo "unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

log() {
  echo "[unisub-uninstall] $*"
}

require_root() {
  if [[ "${EUID}" -ne 0 ]]; then
    echo "please run as root" >&2
    exit 1
  fi
}

main() {
  require_root

  if systemctl list-unit-files | grep -q "^${SERVICE_NAME}.service"; then
    systemctl stop "${SERVICE_NAME}" || true
    systemctl disable "${SERVICE_NAME}" || true
  fi

  rm -f "${SERVICE_FILE}"
  systemctl daemon-reload
  systemctl reset-failed || true

  rm -rf "${INSTALL_ROOT}"

  if [[ "${PURGE_CONFIG}" -eq 1 ]]; then
    log "removing config and state"
    rm -rf "${CONFIG_ROOT}" "${STATE_ROOT}"
    if id -u "${SERVICE_NAME}" >/dev/null 2>&1; then
      userdel "${SERVICE_NAME}" || true
    fi
  else
    log "keeping config under ${CONFIG_ROOT} and state under ${STATE_ROOT}"
    log "kept configuration file path: ${CONFIG_ROOT}/config.yaml"
  fi
}

main
