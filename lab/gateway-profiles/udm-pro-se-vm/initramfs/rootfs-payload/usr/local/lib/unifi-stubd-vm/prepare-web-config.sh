#!/bin/sh
# Prepare the stock UniFi OS setup nginx surface for the VM.
#
# The goal is not to replace UniFi Core. This only supplies the small static
# files that are normally produced by early appliance setup so nginx can serve
# the original setup page on guest port 443.
set +e

PATH=/usr/sbin:/usr/bin:/sbin:/bin

share_dir=/usr/local/share/unifi-stubd-vm/http

write_web_config() {
    # UniFi Core expects these directories before it can generate dynamic nginx
    # snippets. In the VM lab we create them early so stock nginx can start even
    # before Core finishes its own setup path.
    mkdir -p \
        /data/unifi-core/config/http \
        /data/unifi-core/devices \
        /data/unifi-core/logs \
        /var/cache/nginx \
        /var/log/nginx

    # Generate a lab-only certificate if the firmware has not generated one
    # yet. The browser will not trust it, but nginx needs a keypair for port 443.
    if [ ! -f /data/unifi-core/config/unifi-core.key ] ||
        [ ! -f /data/unifi-core/config/unifi-core.crt ]; then
        openssl req -x509 -nodes -newkey rsa:2048 -days 3650 \
            -subj "/CN=UDM-SE-QEMU-LAB" \
            -keyout /data/unifi-core/config/unifi-core.key \
            -out /data/unifi-core/config/unifi-core.crt >/dev/null 2>&1 || true
        chmod 0600 /data/unifi-core/config/unifi-core.key 2>/dev/null || true
    fi

    # Keep the vendor setup site, but point upstreams at local lab endpoints and
    # insert a narrow guard for API paths that are not available in this VM mode.
    cp "$share_dir/local-certs.conf" /data/unifi-core/config/http/local-certs.conf 2>/dev/null || true
    rm -f /data/unifi-core/config/http/upstream-lab.conf
    cp "$share_dir/upstream-uos.conf" /data/unifi-core/config/http/upstream-uos.conf 2>/dev/null || true

    cp /etc/nginx/nginx.conf.disabled /etc/nginx/nginx.conf 2>/dev/null || true
    cp /usr/share/unifi-core/http/ssl-nist.conf /data/unifi-core/config/http/ssl-dynamic.conf 2>/dev/null || true
    cp /usr/share/unifi-core/http/site-setup.conf /data/unifi-core/config/http/site-setup.conf 2>/dev/null || true
    if ! grep -q 'unifi-stubd VM lab API guard' /data/unifi-core/config/http/site-setup.conf 2>/dev/null; then
        tmp_config=$(mktemp)
        awk -f "$share_dir/site-setup-api-guard.awk" \
            /data/unifi-core/config/http/site-setup.conf > "$tmp_config" &&
            cat "$tmp_config" > /data/unifi-core/config/http/site-setup.conf
        rm -f "$tmp_config"
    fi
    chown -R unifi-core:unifi-core /data/unifi-core 2>/dev/null || true
}

write_web_config
if nginx -t; then
    echo "unifi-stubd-vm-web-config: installed vendor setup nginx config for transparent LAN web access"
else
    echo "unifi-stubd-vm-web-config: nginx config test failed" >&2
fi
