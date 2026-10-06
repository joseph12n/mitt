#!/bin/sh
systemctl daemon-reload >/dev/null 2>&1 || :
update-desktop-database >/dev/null 2>&1 || :
