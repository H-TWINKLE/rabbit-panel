#!/usr/bin/env bash

set -euo pipefail

MANIFEST_URL="${RABBIT_INSTALL_MANIFEST_URL:-https://reisen7.github.io/rabbit-panel/latest.json}"
INSTALL_DIR="${RABBIT_INSTALL_DIR:-/opt/rabbit-panel}"
BIN_PATH="${RABBIT_BIN_PATH:-/usr/local/bin/rabbit-panel}"
SERVICE_NAME="${RABBIT_SERVICE_NAME:-rabbit-panel}"
ENV_FILE="${RABBIT_ENV_FILE:-/etc/${SERVICE_NAME}.env}"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
TMP_DIR="$(mktemp -d)"

cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

require_root() {
  if [ "${EUID:-$(id -u)}" -ne 0 ]; then
    echo "[ERROR] Please run as root."
    exit 1
  fi
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo "linux-amd64" ;;
    aarch64|arm64) echo "linux-arm64" ;;
    armv7l|armv7) echo "linux-armv7" ;;
    *)
      echo "[ERROR] Unsupported architecture: $(uname -m)"
      exit 1
      ;;
  esac
}

require_cmd() {
  local cmd="$1"
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "[ERROR] Missing required command: $cmd"
    exit 1
  fi
}

fetch_manifest() {
  local manifest_path="$TMP_DIR/latest.json"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$MANIFEST_URL" -o "$manifest_path"
  elif command -v wget >/dev/null 2>&1; then
    wget -qO "$manifest_path" "$MANIFEST_URL"
  else
    echo "[ERROR] curl or wget is required."
    exit 1
  fi
  echo "$manifest_path"
}

read_json() {
  local file="$1"
  local key="$2"
  python3 - "$file" "$key" <<'PY'
import json, sys
file_path, key = sys.argv[1], sys.argv[2]
with open(file_path, 'r', encoding='utf-8') as f:
    data = json.load(f)
value = data
for part in key.split('.'):
    value = value[part]
print(value)
PY
}

download_file() {
  local url="$1"
  local output="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$output"
  else
    wget -qO "$output" "$url"
  fi
}

verify_sha256() {
  local file="$1"
  local expected="$2"
  if [ -z "$expected" ]; then
    return 0
  fi
  local actual=""
  if command -v sha256sum >/dev/null 2>&1; then
    actual="$(sha256sum "$file" | awk '{print $1}')"
  elif command -v shasum >/dev/null 2>&1; then
    actual="$(shasum -a 256 "$file" | awk '{print $1}')"
  elif command -v openssl >/dev/null 2>&1; then
    actual="$(openssl dgst -sha256 "$file" | awk '{print $NF}')"
  else
    echo "[ERROR] No SHA256 verification tool found."
    exit 1
  fi

  if [ "$actual" != "$expected" ]; then
    echo "[ERROR] SHA256 mismatch."
    echo "Expected: $expected"
    echo "Actual:   $actual"
    exit 1
  fi
}

ensure_env_file() {
  mkdir -p "$(dirname "$ENV_FILE")"
  if [ ! -f "$ENV_FILE" ]; then
    local jwt_secret
    local node_secret
    jwt_secret="$(tr -dc 'A-Za-z0-9' </dev/urandom | head -c 32)"
    node_secret="$(tr -dc 'A-Za-z0-9' </dev/urandom | head -c 32)"
    cat >"$ENV_FILE" <<EOF
MODE=master
HOST=0.0.0.0
PORT=3958
JWT_SECRET=${jwt_secret}
NODE_SECRET=${node_secret}
RABBIT_DEPLOY_MODE=binary
RABBIT_SERVICE_NAME=${SERVICE_NAME}
EOF
    chmod 600 "$ENV_FILE"
  fi
}

install_service() {
  cat >"$SERVICE_FILE" <<EOF
[Unit]
Description=Rabbit Panel
After=network.target docker.service
Wants=docker.service

[Service]
Type=simple
EnvironmentFile=${ENV_FILE}
WorkingDirectory=${INSTALL_DIR}
ExecStart=${BIN_PATH}
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF
}

main() {
  require_root
  require_cmd python3
  require_cmd systemctl

  local arch
  arch="$(detect_arch)"
  local manifest
  manifest="$(fetch_manifest)"

  local version
  local download_url
  local checksum
  version="$(read_json "$manifest" version)"
  download_url="$(read_json "$manifest" "assets.${arch}")"
  checksum="$(python3 - "$manifest" "$arch" <<'PY'
import json, sys
path, arch = sys.argv[1], sys.argv[2]
with open(path, 'r', encoding='utf-8') as f:
    data = json.load(f)
print(data.get('sha256', {}).get(arch, ''))
PY
)"

  echo "[INFO] Installing Rabbit Panel ${version} for ${arch}"

  mkdir -p "$INSTALL_DIR"
  local download_path="$TMP_DIR/rabbit-panel"
  download_file "$download_url" "$download_path"
  chmod +x "$download_path"
  verify_sha256 "$download_path" "$checksum"

  ensure_env_file

  if [ -f "$BIN_PATH" ]; then
    cp "$BIN_PATH" "${BIN_PATH}.bak"
  fi

  cp "$download_path" "$BIN_PATH"
  chmod +x "$BIN_PATH"

  install_service
  systemctl daemon-reload
  systemctl enable "$SERVICE_NAME"
  systemctl restart "$SERVICE_NAME"

  echo "[OK] Rabbit Panel installed."
  echo "[INFO] Service: ${SERVICE_NAME}"
  echo "[INFO] Binary:  ${BIN_PATH}"
  echo "[INFO] Config:  ${ENV_FILE}"
  echo "[INFO] Visit:   http://$(hostname -I | awk '{print $1}'):3958"
}

main "$@"
