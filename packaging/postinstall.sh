#!/bin/sh
# Generate a stable pairing token once: it survives restarts and upgrades,
# so the dashboard QR never changes. Never overwrite an existing file.
if [ ! -f /etc/mitt/pairing.env ]; then
  mkdir -p /etc/mitt
  token=$(head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n')
  printf 'MITT_TOKEN=%s\n' "$token" > /etc/mitt/pairing.env
  chmod 600 /etc/mitt/pairing.env
fi
systemctl daemon-reload >/dev/null 2>&1 || :
update-desktop-database >/dev/null 2>&1 || :
