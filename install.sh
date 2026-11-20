#!/bin/sh
# Installs ZExporter as a systemd service (Ubuntu, Debian and other systemd distros). Run it again to change the
# server or to upgrade: the saved address is reused unless --server is given.
#   curl -fsSL https://github.com/ZinOctopus/exporter/releases/latest/download/install.sh | sudo sh -s -- --server host:port
#   sudo sh install.sh --uninstall
set -eu

REPO=ZinOctopus/exporter
BIN=/usr/local/bin/zexporter
UNIT=/etc/systemd/system/zexporter.service
SERVER=""
UNINSTALL=0

while [ $# -gt 0 ]; do
  case "$1" in
    --server) [ $# -ge 2 ] || { echo "--server needs host:port" >&2; exit 1; }; SERVER=$2; shift 2 ;;
    --uninstall) UNINSTALL=1; shift ;;
    *) echo "usage: install.sh [--server host:port] [--uninstall]" >&2; exit 1 ;;
  esac
done

[ "$(id -u)" -eq 0 ] || { echo "run as root: sudo sh install.sh ..." >&2; exit 1; }
command -v systemctl >/dev/null || { echo "systemd not found; see the README to run the binary by hand" >&2; exit 1; }

if [ "$UNINSTALL" -eq 1 ]; then
  systemctl disable --now zexporter 2>/dev/null || true
  rm -f "$UNIT" "$BIN"
  systemctl daemon-reload
  echo "ZExporter removed (device id kept in /var/lib/zexporter)."
  exit 0
fi

case "$(uname -m)" in
  x86_64|amd64) ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  *) echo "unsupported CPU: $(uname -m)" >&2; exit 1 ;;
esac

# Reuse the saved address, else ask (from the terminal, since stdin may be a pipe).
if [ -z "$SERVER" ] && [ -f "$UNIT" ]; then
  SERVER=$(sed -n 's/^ExecStart=.* --server \([^ ]*\).*/\1/p' "$UNIT")
fi
if [ -z "$SERVER" ] && [ -r /dev/tty ]; then
  printf "BoxServer gRPC address (host:port): " >/dev/tty
  read -r SERVER </dev/tty
fi
# Only characters that are safe inside the unit file.
echo "$SERVER" | grep -Eq '^[A-Za-z0-9._-]+:[0-9]+$' || { echo "server must be host:port, e.g. 192.168.1.10:8081" >&2; exit 1; }

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
cd "$TMP"
curl -fsSLO "https://github.com/$REPO/releases/latest/download/zexporter-linux-$ARCH"
curl -fsSLO "https://github.com/$REPO/releases/latest/download/SHA256SUMS"
grep " zexporter-linux-$ARCH\$" SHA256SUMS | sha256sum -c -

systemctl stop zexporter 2>/dev/null || true
install -m 0755 "zexporter-linux-$ARCH" "$BIN"

cat > "$UNIT" <<UNITEOF
[Unit]
Description=ZExporter
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=$BIN --server $SERVER
Restart=always

[Install]
WantedBy=multi-user.target
UNITEOF

systemctl daemon-reload
systemctl enable zexporter >/dev/null 2>&1
systemctl restart zexporter
echo "ZExporter installed, reporting to $SERVER. Logs: journalctl -u zexporter -f"
