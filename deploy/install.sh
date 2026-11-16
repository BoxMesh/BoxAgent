#!/bin/sh
# Installs /tmp/boxagent as a systemd service that starts on boot. Run as root by `make deploy`.
set -eu

install -m 0755 /tmp/boxagent /usr/local/bin/boxagent
# shortcut: no log rotation (~15 MB/day), drop when the agent ships metrics to BoxServer instead of the log.
LOG=/var/log/boxagent.log

cat > /etc/systemd/system/boxagent.service <<EOF
[Unit]
Description=BoxAgent
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=/usr/local/bin/boxagent
Restart=always
StandardOutput=append:$LOG
StandardError=append:$LOG

[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
systemctl enable boxagent
systemctl restart boxagent
