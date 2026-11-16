#!/bin/sh
# Installs /tmp/boxagent as a systemd service that starts on boot and reports to the BoxServer given as $1
# (host:port of its gRPC port). Run as root by `make deploy`.
set -eu

SERVER=${1:?usage: install.sh <boxserver host:port>}
install -m 0755 /tmp/boxagent /usr/local/bin/boxagent
# Holds only start-up lines and errors: samples go to BoxServer.
LOG=/var/log/boxagent.log

cat > /etc/systemd/system/boxagent.service <<EOF
[Unit]
Description=BoxAgent
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=/usr/local/bin/boxagent -server $SERVER
Restart=always
StandardOutput=append:$LOG
StandardError=append:$LOG

[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
systemctl enable boxagent
systemctl restart boxagent
