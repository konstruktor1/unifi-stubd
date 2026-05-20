#!/bin/bash
# Service startup helpers for the Docker webportal profile.

ensure_lab_lan_bridge() {
    local bridge="${UNIFI_FW_SIM_LAN_BRIDGE:-br0}"
    local bridge_ip="${UNIFI_FW_SIM_LAN_BRIDGE_IP:-192.168.1.1/24}"
    local bridge_mac="${UNIFI_FW_SIM_MAC:-02:15:6d:00:ea:2c}"
    local lan_port="${UNIFI_FW_SIM_LAN_DUMMY_IFACE:-rtl8370-lan1}"
    local lan_mac="${UNIFI_FW_SIM_LAN_DUMMY_MAC:-02:15:6d:00:ea:31}"

    # UniFi Core expects a local LAN bridge even in the reduced container. The
    # dummy port gives the application a stable interface to inspect.
    if ! ip link show "$bridge" >/dev/null 2>&1; then
        ip link add name "$bridge" type bridge
    fi

    ip link set dev "$bridge" address "$bridge_mac" 2>/dev/null || true
    ip addr replace "$bridge_ip" dev "$bridge"

    if ! ip link show "$lan_port" >/dev/null 2>&1; then
        ip link add name "$lan_port" type dummy
    fi

    ip link set dev "$lan_port" address "$lan_mac" 2>/dev/null || true
    ip link set dev "$lan_port" master "$bridge"
    ip link set dev "$lan_port" up
    ip link set dev "$bridge" up
}

start_postgres() {
    mkdir -p /var/run/postgresql /var/log/postgresql
    chown postgres:postgres /var/run/postgresql /var/log/postgresql

    if pg_lsclusters 2>/dev/null | awk 'NR > 1 && $1 == "14" && $2 == "main" { found = 1 } END { exit !found }'; then
        pg_ctlcluster 14 main start >/dev/null 2>&1 || true
    else
        pg_createcluster --locale C.UTF-8 --encoding UTF8 14 main --start \
            >"$log_dir/postgres-create.log" \
            2>"$log_dir/postgres-create.err"
    fi

    # The firmware packages assume local PostgreSQL roles already exist. Create
    # the smallest schema surface needed by unifi-core and ulp-go.
    runuser -u postgres -- psql -tc "SELECT 1 FROM pg_roles WHERE rolname = 'unifi-core'" | grep -q 1 \
        || runuser -u postgres -- createuser "unifi-core" -d
    runuser -u postgres -- psql -c "ALTER USER \"unifi-core\" WITH PASSWORD '$postgres_password'" >/dev/null
    runuser -u postgres -- psql -tc "SELECT 1 FROM pg_database WHERE datname = 'unifi-core'" | grep -q 1 \
        || runuser -u postgres -- createdb -O "unifi-core" "unifi-core"

    runuser -u postgres -- psql -tc "SELECT 1 FROM pg_roles WHERE rolname = 'ulp-go'" | grep -q 1 \
        || runuser -u postgres -- createuser "ulp-go"
    for database in "ulp-go" "ulp-go-syslog"; do
        runuser -u postgres -- psql -tc "SELECT 1 FROM pg_database WHERE datname = '$database'" | grep -q 1 \
            || runuser -u postgres -- createdb -O "ulp-go" "$database"
        runuser -u postgres -- psql -c "GRANT ALL PRIVILEGES ON DATABASE \"$database\" TO \"ulp-go\"" >/dev/null
    done
}

write_unifi_core_override() {
    local target="/data/unifi-core/config/overrides/default.yaml"

    mkdir -p /data/unifi-core/config/overrides
    POSTGRES_PASSWORD="$postgres_password" \
        awk '
            {
                line = $0
                token = "@POSTGRES_PASSWORD@"
                while ((idx = index(line, token)) > 0) {
                    line = substr(line, 1, idx - 1) ENVIRON["POSTGRES_PASSWORD"] substr(line, idx + length(token))
                }
                print line
            }
        ' \
        "$template_dir/unifi-core-default.yaml.in" > "$target"
    chown -R unifi-core:unifi-core /data/unifi-core
}

prepare_support_bundle_paths() {
    mkdir -p \
        /data/ulp-go/log \
        /data/uid/log \
        /data/unifi/logs \
        /srv/unifi-protect/logs \
        /data/unifi-access/log \
        /var/log/unifi-talk \
        /var/log/freeswitch \
        /var/log/unifi-talk-relay \
        /data/unifi-connect/log \
        /srv/unifi-connect/log \
        /data/unifi-drive/logs \
        /data/unifi-innerspace/log \
        /data/apollo/logs \
        /data/unifi-core/supportFile \
        /srv/unifi-core/supportFile
    : >/var/log/freeswitch/freeswitch.log
    chown -R unifi-core:unifi-core /data/unifi-core /srv/unifi-core
    chown -R ulp-go:ulp-go /data/ulp-go 2>/dev/null || true
}

start_dbus() {
    mkdir -p /var/run/dbus
    if [[ ! -S /var/run/dbus/system_bus_socket ]]; then
        dbus-daemon --system --fork --nopidfile
    fi
}

start_systemd_stub() {
    if ! pgrep -f "$systemd_dbus_stub" >/dev/null 2>&1; then
        /usr/bin/node24 "$systemd_dbus_stub" \
            >"$log_dir/systemd-dbus-stub.log" \
            2>"$log_dir/systemd-dbus-stub.err" &
    fi
}

start_network_app_stub() {
    if ! pgrep -f "$network_app_stub" >/dev/null 2>&1; then
        /usr/bin/node24 "$network_app_stub" \
            >"$log_dir/network-app-stub.stdout" \
            2>"$log_dir/network-app-stub.stderr" &
    fi
}

start_ulp_go() {
    if pgrep -x ulp-go-app >/dev/null 2>&1; then
        return
    fi

    mkdir -p /data/ulp-go/log /data/ulp-go/ws /data/ulp-go/tmp /run/ulp-go /srv/ulp-go
    chown -R ulp-go:ulp-go /data/ulp-go /run/ulp-go /srv/ulp-go

    /usr/lib/ulp-go/scripts/service/start.sh \
        >"$log_dir/ulp-go-start.log" \
        2>"$log_dir/ulp-go-start.err" &

    for _ in {1..45}; do
        if ss -ltn | grep -q '127[.]0[.]0[.]1:9080'; then
            return
        fi
        sleep 1
    done

    echo "warning: ulp-go did not expose 127.0.0.1:9080 within the wait window" >&2
}

start_nginx() {
    if [[ -f /var/run/nginx.pid ]] && kill -0 "$(cat /var/run/nginx.pid)" 2>/dev/null; then
        return
    fi

    mkdir -p /var/log/nginx /var/cache/nginx /data/unifi-core/logs
    chown -R nginx:nginx /var/cache/nginx
    chown -R unifi-core:unifi-core /data/unifi-core/logs

    nginx >"$log_dir/nginx-start.log" 2>"$log_dir/nginx-start.err" || {
        cat "$log_dir/nginx-start.err" >&2
        return 1
    }
}

start_unifi_core() {
    if pgrep -f "/usr/share/unifi-core/app/service.js" >/dev/null 2>&1; then
        return
    fi

    USER=unifi-core GROUP=unifi-core /usr/share/unifi-core/app/hooks/pre-start \
        >"$log_dir/unifi-core-pre.log" \
        2>"$log_dir/unifi-core-pre.err"

    start_nginx

    # Start the original Node service directly. The surrounding wrappers provide
    # the systemd, HTTP, and host-tool behavior it normally receives on hardware.
    (
        set -a
        # shellcheck disable=SC1091
        . /etc/default/unifi-core
        set +a
        cd /usr/share/unifi-core/app
        exec runuser -u unifi-core -- \
            /usr/bin/node24 \
                --expose-gc \
                --max-old-space-size=300 \
                --openssl-legacy-provider \
                --no-network-family-autoselection \
                --dns-result-order=ipv4first \
                /usr/share/unifi-core/app/service.js
    ) >"$log_dir/unifi-core.log" 2>"$log_dir/unifi-core.err" &

    for _ in {1..60}; do
        if [[ -S /data/unifi-core/config/http/uos-http.sock ]]; then
            write_lab_http_overrides
            enable_lab_http_preview
            block_host_mutating_http_endpoints
            return
        fi
        sleep 1
    done

    echo "warning: unifi-core did not create uos-http.sock within the wait window" >&2
}
