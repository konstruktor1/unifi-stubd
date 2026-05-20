#!/bin/sh
# Serial-friendly web stack diagnostic dump.
#
# This deliberately prints to journal+console through the service unit so a VM
# boot can be debugged from the serial console when networking is broken.
set +e

PATH=/usr/sbin:/usr/bin:/sbin:/bin

echo "===== unifi-stubd-vm-web-debug: begin ====="
date -Iseconds 2>/dev/null || date

echo "--- memory ---"
free -m 2>&1 || true

echo "--- filesystems ---"
df -h 2>&1 || true
df -i 2>&1 || true

echo "--- listeners ---"
ss -ltnp 2>&1 || true

echo "--- nginx config ---"
nginx -t 2>&1 || true

echo "--- nginx processes ---"
ps w 2>&1 | grep '[n]ginx' || true

echo "--- systemd status ---"
systemctl --no-pager --full status \
    nginx.service \
    unifi-core.service \
    unifi.service \
    unifi-directory.service \
    rabbitmq-server.service \
    postgresql@14-main.service \
    postgresql.service 2>&1 || true

for unit in nginx.service unifi-core.service unifi.service unifi-directory.service rabbitmq-server.service postgresql@14-main.service postgresql.service; do
    echo "--- journal $unit ---"
    journalctl --no-pager -n 220 -u "$unit" 2>&1 || true
done

echo "--- unifi-core logs ---"
find /data/unifi-core/logs /data/unifi/logs /usr/lib/unifi/logs \
    -maxdepth 1 -type f 2>/dev/null | sort || true
for file in \
    /data/unifi-core/logs/*.log \
    /data/unifi-core/logs/*.err \
    /data/unifi/logs/*.log \
    /data/unifi/logs/*.err \
    /usr/lib/unifi/logs/*.log \
    /usr/lib/unifi/logs/*.err \
    /tmp/udm-pro-se-webportal/*.log \
    /tmp/udm-pro-se-webportal/*.err; do
    [ -f "$file" ] || continue
    echo "--- tail $file ---"
    tail -140 "$file" 2>&1 || true
done

echo "===== unifi-stubd-vm-web-debug: end ====="
