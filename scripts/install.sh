#!/usr/bin/env bash
set -euo pipefail

REPO="${REPO:-willbon-dev/UniSub}"
RELEASE_VERSION="latest"
INSTALL_ROOT="/opt/unisub"
CONFIG_ROOT="/etc/unisub"
STATE_ROOT="/var/lib/unisub"
SERVICE_NAME="unisub"
BIN_NAME="unisub"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
TMP_DIR=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      RELEASE_VERSION="${2:?missing value for --version}"
      shift 2
      ;;
    *)
      echo "unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

log() {
  echo "[unisub-install] $*"
}

cleanup() {
  if [[ -n "${TMP_DIR}" && -d "${TMP_DIR}" ]]; then
    rm -rf "${TMP_DIR}"
  fi
}

require_root() {
  if [[ "${EUID}" -ne 0 ]]; then
    echo "please run as root" >&2
    exit 1
  fi
}

check_os() {
  if [[ ! -f /etc/os-release ]]; then
    echo "unsupported system" >&2
    exit 1
  fi
  # shellcheck disable=SC1091
  . /etc/os-release
  if [[ "${ID:-}" != "ubuntu" || "${VERSION_ID:-}" != "24.04" ]]; then
    echo "this installer supports Ubuntu 24.04 only" >&2
    exit 1
  fi
}

detect_arch() {
  case "$(uname -m)" in
    x86_64) GOARCH="amd64" ;;
    aarch64|arm64) GOARCH="arm64" ;;
    *)
      echo "unsupported architecture: $(uname -m)" >&2
      exit 1
      ;;
  esac
}

fetch() {
  local url="$1"
  log "fetching metadata from ${url}"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url"
  elif command -v wget >/dev/null 2>&1; then
    wget -qO- "$url"
  else
    echo "curl or wget is required" >&2
    exit 1
  fi
}

download_to() {
  local url="$1"
  local output="$2"
  log "downloading ${url} -> ${output}"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$output"
  elif command -v wget >/dev/null 2>&1; then
    wget -qO "$output" "$url"
  else
    echo "curl or wget is required" >&2
    exit 1
  fi
}

release_download_url() {
  local asset_name="$1"
  if [[ "${RELEASE_VERSION}" == "latest" ]]; then
    echo "https://github.com/${REPO}/releases/latest/download/${asset_name}"
  else
    echo "https://github.com/${REPO}/releases/download/${RELEASE_VERSION}/${asset_name}"
  fi
}

write_service_file() {
  cat >"${SERVICE_FILE}" <<EOF
[Unit]
Description=UniSub unified subscription service
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=${SERVICE_NAME}
Group=${SERVICE_NAME}
ExecStart=${INSTALL_ROOT}/bin/${BIN_NAME} -config ${CONFIG_ROOT}/config.yaml
Restart=on-failure
RestartSec=5
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=${STATE_ROOT}
WorkingDirectory=${STATE_ROOT}

[Install]
WantedBy=multi-user.target
EOF
}

main() {
  trap cleanup EXIT
  require_root
  check_os
  detect_arch

  local asset_name="unisub_linux_${GOARCH}.tar.gz"
  log "resolving release asset ${asset_name}"
  local download_url
  download_url="$(release_download_url "${asset_name}")"
  log "resolved release download URL: ${download_url}"

  TMP_DIR="$(mktemp -d)"

  log "downloading ${download_url}"
  download_to "$download_url" "${TMP_DIR}/${asset_name}"
  tar -xzf "${TMP_DIR}/${asset_name}" -C "${TMP_DIR}"

  if ! id -u "${SERVICE_NAME}" >/dev/null 2>&1; then
    useradd --system --home "${STATE_ROOT}" --shell /usr/sbin/nologin "${SERVICE_NAME}"
  fi

  mkdir -p "${INSTALL_ROOT}/bin" "${CONFIG_ROOT}" "${STATE_ROOT}"
  install -m 0755 "${TMP_DIR}/${BIN_NAME}" "${INSTALL_ROOT}/bin/${BIN_NAME}"

  if [[ ! -f "${CONFIG_ROOT}/config.yaml" ]]; then
    log "installing example config to ${CONFIG_ROOT}/config.yaml"
    download_to "https://raw.githubusercontent.com/${REPO}/main/docs/config.example.yaml" "${CONFIG_ROOT}/config.yaml"
    chmod 0640 "${CONFIG_ROOT}/config.yaml"
  else
    log "keeping existing config at ${CONFIG_ROOT}/config.yaml"
  fi

  write_service_file
  systemctl daemon-reload
  systemctl enable --now "${SERVICE_NAME}"
  systemctl --no-pager --full status "${SERVICE_NAME}" || true
  log "configuration file: ${CONFIG_ROOT}/config.yaml"
  log "after editing config, restart the service with: systemctl restart ${SERVICE_NAME}"
}

main
