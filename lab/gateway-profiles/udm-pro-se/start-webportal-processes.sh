#!/bin/bash
# Start the UDM Pro SE firmware simulation plus the minimal UniFi OS webportal.
set -euo pipefail

log_dir="${UNIFI_FW_SIM_WEB_LOG_DIR:-/tmp/udm-pro-se-webportal}"
postgres_password="${UNIFI_CORE_POSTGRES_PASSWORD:-unifi-core-lab-pass}"

mkdir -p "$log_dir"

write_ubnt_tools_wrapper() {
    if [[ -x /sbin/ubnt-tools && ! -e /sbin/ubnt-tools.real ]]; then
        mv /sbin/ubnt-tools /sbin/ubnt-tools.real
    fi

    cat >/sbin/ubnt-tools <<'SH'
#!/bin/sh
if [ "${1:-}" = "id" ]; then
    cat <<'EOF'
board.sysid=0xea2c
board.name=UDMPROSE
board.shortname=UDMPROSE
board.storename=UDMPROSE
board.subtype=
board.reboot=30
board.upgrade=150
board.cpu.id=410fd034-00000000
board.uuid=73194688-47d7-31ae-a40d-bf5dd963c999
board.bom=113-00000-01
board.hwrev=0x0001
board.serialno=02156D00EA2C
board.qrid=SIMULATED
EOF
    exit 0
fi
if [ -x /sbin/ubnt-tools.real ]; then
    exec /sbin/ubnt-tools.real "$@"
fi
echo "ubnt-tools.real is not available" >&2
exit 127
SH
    chmod 0755 /sbin/ubnt-tools
}

write_ubnt_systool_wrapper() {
    if [[ -x /sbin/ubnt-systool && ! -e /sbin/ubnt-systool.real ]]; then
        mv /sbin/ubnt-systool /sbin/ubnt-systool.real
    fi

    cat >/sbin/ubnt-systool <<'SH'
#!/bin/sh
case "${1:-}" in
    anonid)
        printf '%s\n' '7f2d8c16-7b60-4db9-84a1-02156d00ea2c'
        ;;
    anonidcontroller)
        printf '%s\n' '00000000-0000-0000-0000-000000000000'
        ;;
    cpuload)
        printf '0\n'
        ;;
    cputemp)
        printf '0\n'
        ;;
    hostname)
        if [ -n "${2:-}" ]; then
            printf '%s\n' "$2" >/tmp/udm-pro-se-lab-hostname
        fi
        ;;
    network-speed)
        if [ -n "${3:-}" ]; then
            printf 'mocked ubnt-systool %s\n' "$*" >&2
        else
            printf '{"speed":10000,"auto-nego":"on"}\n'
        fi
        ;;
    support)
        if [ -z "${2:-}" ]; then
            printf 'mocked ubnt-systool support requires a destination directory\n' >&2
            exit 1
        fi
        mkdir -p "$2"
        {
            printf 'UDM Pro SE firmware lab support placeholder\n'
            printf 'generated_by=ubnt-systool-wrapper\n'
            printf 'board=UDMPROSE\n'
        } >"$2/lab-system.txt"
        ubnt-tools id >"$2/ubnt-tools-id.txt" 2>/dev/null || true
        ;;
    sshd)
        if [ -n "${2:-}" ]; then
            printf '%s\n' "$2" >/tmp/udm-pro-se-lab-sshd-enabled
        elif [ "$(cat /tmp/udm-pro-se-lab-sshd-enabled 2>/dev/null || printf false)" = "true" ]; then
            printf 'enabled\n'
        else
            printf 'disabled\n'
        fi
        ;;
    sshpasswd)
        case "${2:-}" in
            get)
                cat /tmp/udm-pro-se-lab-sshpasswd 2>/dev/null || true
                ;;
            set)
                printf '%s\n' "${3:-}" >/tmp/udm-pro-se-lab-sshpasswd
                ;;
            *)
                printf 'unsupported mocked ubnt-systool command: %s\n' "$*" >&2
                exit 1
                ;;
        esac
        ;;
    chpasswd|fwupdate|network|poweroff|reboot|reset2defaults|synctime|timezone)
        printf 'mocked ubnt-systool %s\n' "$*" >&2
        ;;
    *)
        printf 'unsupported mocked ubnt-systool command: %s\n' "$*" >&2
        exit 1
        ;;
esac
exit 0
SH
    chmod 0755 /sbin/ubnt-systool
}

write_systemd_run_wrapper() {
    if [[ -x /usr/bin/systemd-run && ! -e /usr/bin/systemd-run.real ]]; then
        mv /usr/bin/systemd-run /usr/bin/systemd-run.real
    fi

    cat >/usr/bin/systemd-run <<'SH'
#!/bin/sh
# Minimal lab shim for unifi-core transient commands inside a non-systemd container.
while [ "$#" -gt 0 ]; do
    case "$1" in
        --wait|--pipe|--collect|--quiet|--same-dir)
            shift
            ;;
        --unit=*|--property=*|--description=*|--service-type=*)
            shift
            ;;
        --unit|--property|--description|--service-type|-p)
            shift 2
            ;;
        --)
            shift
            break
            ;;
        -*)
            shift
            ;;
        *)
            break
            ;;
    esac
done

if [ "$#" -eq 0 ]; then
    printf 'mocked systemd-run received no command\n' >&2
    exit 1
fi

exec "$@"
SH
    rm -f /usr/local/bin/systemd-run
    chmod 0755 /usr/bin/systemd-run
}

write_systemctl_wrapper() {
    if [[ -x /usr/bin/systemctl && ! -e /usr/bin/systemctl.real ]]; then
        mv /usr/bin/systemctl /usr/bin/systemctl.real
    fi
    if [[ -x /bin/systemctl && ! -e /bin/systemctl.real ]]; then
        mv /bin/systemctl /bin/systemctl.real
    fi

    cat >/usr/bin/systemctl <<'SH'
#!/bin/sh
# Lab-scoped systemctl shim for application lifecycle calls from UniFi Core.
# The container does not boot systemd as PID 1, so only the known UniFi service
# commands are accepted here; unsupported commands remain explicit failures.
log_dir="${UNIFI_FW_SIM_WEB_LOG_DIR:-/tmp/udm-pro-se-webportal}"
log_file="$log_dir/systemctl-wrapper.log"
cmd=""
services=""
start_now=0
quiet=0

mkdir -p "$log_dir"
printf '%s systemctl %s\n' "$(date -Iseconds)" "$*" >>"$log_file"

say() {
    [ "$quiet" -eq 1 ] || printf '%s\n' "$1"
}

start_network_stub() {
    if ss -ltn 2>/dev/null | grep -q '127[.]0[.]0[.]1:8081'; then
        return 0
    fi

    nohup /usr/bin/node24 /usr/local/lib/udm-pro-se-network-app-stub.cjs \
        >"$log_dir/network-app-stub.stdout" \
        2>"$log_dir/network-app-stub.stderr" &
}

is_network_active() {
    ss -ltn 2>/dev/null | grep -q '127[.]0[.]0[.]1:8081'
}

prepare_nginx_runtime() {
    mkdir -p /var/log/nginx /var/cache/nginx /data/unifi-core/logs
    chown -R nginx:nginx /var/cache/nginx 2>/dev/null || true
    chown -R unifi-core:unifi-core /data/unifi-core/logs 2>/dev/null || true
}

is_nginx_active() {
    [ -f /var/run/nginx.pid ] && kill -0 "$(cat /var/run/nginx.pid)" 2>/dev/null
}

start_or_reload_nginx() {
    prepare_nginx_runtime
    nginx -t >/dev/null || return 1

    if is_nginx_active; then
        nginx -s reload >/dev/null 2>&1 || nginx
        return 0
    fi

    nginx
}

normalize_service() {
    service="$1"
    service="${service%.service}"
    printf '%s\n' "$service"
}

for arg in "$@"; do
    case "$arg" in
        is-system-running|daemon-reload|reset-failed|enable|disable|start|stop|restart|try-restart|reload|is-active|is-enabled|status|kill)
            if [ -z "$cmd" ]; then
                cmd="$arg"
            fi
            ;;
        --now)
            start_now=1
            ;;
        -q|--quiet)
            quiet=1
            ;;
        -*)
            ;;
        *)
            services="$services $(normalize_service "$arg")"
            ;;
    esac
done

case "$cmd" in
    is-system-running)
        say "running"
        exit 0
        ;;
    daemon-reload|reset-failed)
        exit 0
        ;;
esac

for service in $services; do
    case "$service" in
        unifi)
            case "$cmd" in
                enable|start|restart|try-restart|reload)
                    [ "$cmd" = "enable" ] && [ "$start_now" -ne 1 ] && exit 0
                    start_network_stub
                    exit 0
                    ;;
                is-active|status)
                    if is_network_active; then
                        say "active"
                        exit 0
                    fi
                    say "inactive"
                    exit 3
                    ;;
                is-enabled)
                    say "enabled"
                    exit 0
                    ;;
                stop|disable|kill)
                    exit 0
                    ;;
            esac
            ;;
        nginx)
            case "$cmd" in
                enable)
                    [ "$start_now" -eq 1 ] && start_or_reload_nginx
                    exit 0
                    ;;
                start|restart|try-restart|reload)
                    start_or_reload_nginx
                    exit $?
                    ;;
                is-active|status)
                    if is_nginx_active; then
                        say "active"
                        exit 0
                    fi
                    say "inactive"
                    exit 3
                    ;;
                is-enabled)
                    say "enabled"
                    exit 0
                    ;;
                stop|disable|kill)
                    nginx -s quit >/dev/null 2>&1 || true
                    exit 0
                    ;;
            esac
            ;;
        udapi-server)
            case "$cmd" in
                start|restart|try-restart|reload|enable)
                    exit 0
                    ;;
                is-active|status)
                    if [ -S /var/run/ubnt-udapi-server.sock ]; then
                        say "active"
                        exit 0
                    fi
                    say "inactive"
                    exit 3
                    ;;
                is-enabled)
                    say "enabled"
                    exit 0
                    ;;
                stop|disable|kill)
                    exit 0
                    ;;
            esac
            ;;
        unifi-protect|unifi-access|unifi-talk|unifi-talk-relay|unifi-connect|unifi-drive|unifi-innerspace|apollo|uid-agent|ulp-go)
            case "$cmd" in
                enable|start|restart|try-restart|reload|stop|disable|kill)
                    printf 'mocked systemctl %s %s\n' "$cmd" "$service" >>"$log_file"
                    exit 0
                    ;;
                is-active|status)
                    say "inactive"
                    exit 3
                    ;;
                is-enabled)
                    say "disabled"
                    exit 1
                    ;;
            esac
            ;;
    esac
done

case "$cmd" in
    is-active|status)
        say "inactive"
        exit 3
        ;;
    is-enabled)
        say "disabled"
        exit 1
        ;;
esac

printf 'unsupported mocked systemctl command: %s\n' "$*" >&2
exit 1
SH
    chmod 0755 /usr/bin/systemctl
    if [[ ! -e /bin/systemctl || ! /usr/bin/systemctl -ef /bin/systemctl ]]; then
        cp /usr/bin/systemctl /bin/systemctl
        chmod 0755 /bin/systemctl
    fi
}

write_lab_sudoers() {
    # UniFi Core enables applications through sudo. The stock firmware sudoers
    # allows /bin/systemctl, while the container writes the shim to /usr/bin too.
    # Keep this narrow: the shim itself only accepts reviewed lab service names.
    cat >/etc/sudoers.d/unifi-core-lab <<'SUDO'
unifi-core ALL=SETENV: NOPASSWD: /usr/bin/systemctl, /bin/systemctl
SUDO
    chmod 0440 /etc/sudoers.d/unifi-core-lab
}

write_timedatectl_wrapper() {
    if [[ -x /usr/bin/timedatectl && ! -e /usr/bin/timedatectl.real ]]; then
        mv /usr/bin/timedatectl /usr/bin/timedatectl.real
    fi

    cat >/usr/bin/timedatectl <<'SH'
#!/bin/sh
# Minimal timedatectl output for UniFi Core inside the non-systemd lab.
case "${1:-}" in
    show)
        printf 'yes\n'
        printf 'yes\n'
        exit 0
        ;;
    set-timezone)
        printf '%s\n' "${2:-UTC}" >/tmp/udm-pro-se-lab-timezone
        exit 0
        ;;
    *)
        printf 'Local time: lab\n'
        printf 'Universal time: lab\n'
        printf 'RTC time: lab\n'
        printf 'Time zone: %s\n' "$(cat /tmp/udm-pro-se-lab-timezone 2>/dev/null || printf Europe/Zurich)"
        printf 'System clock synchronized: yes\n'
        printf 'NTP service: active\n'
        exit 0
        ;;
esac
SH
    chmod 0755 /usr/bin/timedatectl
}

write_tar_wrapper() {
    cat >/usr/local/bin/tar <<'SH'
#!/bin/sh
# Lab support bundles can include logs that change while tar is reading them.
# Accept exit code 1 only when a readable archive was still produced.
log_file="${UNIFI_FW_SIM_WEB_LOG_DIR:-/tmp/udm-pro-se-webportal}/tar-wrapper.log"
err_file="${UNIFI_FW_SIM_WEB_LOG_DIR:-/tmp/udm-pro-se-webportal}/tar-wrapper.err"
archive=""
next_is_archive=0

mkdir -p "$(dirname "$log_file")"
printf 'tar %s\n' "$*" >>"$log_file"

for arg in "$@"; do
    if [ "$next_is_archive" = 1 ]; then
        archive="$arg"
        next_is_archive=0
        continue
    fi
    case "$arg" in
        -f|-*f)
            next_is_archive=1
            ;;
    esac
done

/bin/tar "$@" 2>>"$err_file"
status=$?
if [ "$status" -eq 1 ] && [ -n "$archive" ] && [ -s "$archive" ]; then
    if /bin/tar -tzf "$archive" >/dev/null 2>>"$err_file"; then
        printf 'accepted tar exit 1 for readable archive %s\n' "$archive" >>"$log_file"
        exit 0
    fi
fi
exit "$status"
SH
    chmod 0755 /usr/local/bin/tar
}

write_udapi_lab_wrappers() {
    # UniFi Core decides whether the console has internet by asking UDAPI for a
    # plugged WAN. The reduced container only has Docker eth0, so these wrappers
    # expose that interface as the lab WAN while delegating unsupported commands
    # to the original firmware binaries.
    local tool

    for tool in mca-ctrl mca-dump ubios-udapi-client; do
        if [[ -x "/usr/bin/$tool" && ! -e "/usr/bin/$tool.real" ]]; then
            if ! grep -q "udm-pro-se-udapi-lab-shim" "/usr/bin/$tool" 2>/dev/null; then
                mv "/usr/bin/$tool" "/usr/bin/$tool.real"
            fi
        fi

        cat >"/usr/bin/$tool" <<SH
#!/bin/sh
exec /usr/bin/node24 /usr/local/lib/udm-pro-se-udapi-lab-shim.cjs $tool "\$@"
SH
        chmod 0755 "/usr/bin/$tool"
    done
}

ensure_lab_lan_bridge() {
    # UniFi Core reads LAN IPs from Node's os.networkInterfaces(), not from
    # UDAPI. A bridge without a carrier is hidden there, so attach one dummy
    # RTL8370-style LAN port to make br0 visible as the simulated gateway LAN.
    local bridge="${UNIFI_FW_SIM_LAN_BRIDGE:-br0}"
    local bridge_ip="${UNIFI_FW_SIM_LAN_BRIDGE_IP:-192.168.1.1/24}"
    local bridge_mac="${UNIFI_FW_SIM_MAC:-02:15:6d:00:ea:2c}"
    local lan_port="${UNIFI_FW_SIM_LAN_DUMMY_IFACE:-rtl8370-lan1}"
    local lan_mac="${UNIFI_FW_SIM_LAN_DUMMY_MAC:-02:15:6d:00:ea:31}"

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
    mkdir -p /data/unifi-core/config/overrides
    cat >/data/unifi-core/config/overrides/default.yaml <<YAML
overrideConsoleFeatures:
  waitForUFN: false
postgres:
  password: "$postgres_password"
YAML
    chown -R unifi-core:unifi-core /data/unifi-core
}

prepare_support_bundle_paths() {
    # unifi-core packages every configured application support path, even when
    # the matching application is not installed in this minimal firmware lab.
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
    if ! pgrep -f "udm-pro-se-systemd-dbus-stub.cjs" >/dev/null 2>&1; then
        /usr/bin/node24 /usr/local/lib/udm-pro-se-systemd-dbus-stub.cjs \
            >"$log_dir/systemd-dbus-stub.log" \
            2>"$log_dir/systemd-dbus-stub.err" &
    fi
}

start_network_app_stub() {
    if ! pgrep -f "udm-pro-se-network-app-stub.cjs" >/dev/null 2>&1; then
        /usr/bin/node24 /usr/local/lib/udm-pro-se-network-app-stub.cjs \
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

block_host_mutating_http_endpoints() {
    local site_config="/data/unifi-core/config/http/site-setup.conf"
    local tmp_config

    if [[ ! -f "$site_config" ]] || grep -q "UDM Pro SE lab reset guard" "$site_config"; then
        return
    fi

    tmp_config="$(mktemp)"
    awk '
        /^[[:space:]]*# UniFi OS public API/ && !inserted {
            print "    # UDM Pro SE lab reset guard."
            print "    # These endpoints would normally reboot or reset the console."
            print "    # The firmware lab blocks them because host-mutating actions are no-ops here."
            print "    location = /api/setup/reset { return 403; }"
            print "    location = /api/system/reboot { return 403; }"
            print "    location = /api/system/reset { return 403; }"
            inserted = 1
        }
        { print }
    ' "$site_config" >"$tmp_config"
    cat "$tmp_config" >"$site_config"
    rm -f "$tmp_config"

    nginx -s reload >/dev/null 2>&1 || true
}

write_lab_http_overrides() {
    local lab_config="/data/unifi-core/config/http/shared-runnable-lab.conf"

    cat >"$lab_config" <<'NGINX'
# UDM Pro SE firmware lab compatibility endpoints.
# The local portal polls this route for console grouping metadata. UniFi Core
# returns 405 in the reduced lab profile, which can leave the browser shell on a
# blank dark page while the real backend is otherwise healthy.
location = /api/device/groups {
    include /usr/share/unifi-core/http/cors.conf;
    include /usr/share/unifi-core/http/security.conf;

    default_type application/json;
    return 200 '{"groups":[],"devices":[]}';
}

# The local portal asks the Users runnable for optional Access and user-asset
# metadata even when those applications are not part of this minimal lab stack.
# Return empty capability documents instead of surfacing 502s from absent
# sidecar services.
location = /proxy/users/user-assets/api/v1/info {
    include /usr/share/unifi-core/http/cors.conf;
    include /usr/share/unifi-core/http/security.conf;
    include /usr/share/unifi-core/http/auth.conf;

    default_type application/json;
    return 200 '{"enabled":false}';
}

location = /proxy/users/access/api/v2/access/feature {
    include /usr/share/unifi-core/http/cors.conf;
    include /usr/share/unifi-core/http/security.conf;
    include /usr/share/unifi-core/http/auth.conf;

    default_type application/json;
    return 200 '{"enabled":false,"features":{}}';
}

location = /proxy/users/access/api/v2/access/info {
    include /usr/share/unifi-core/http/cors.conf;
    include /usr/share/unifi-core/http/security.conf;
    include /usr/share/unifi-core/http/auth.conf;

    default_type application/json;
    return 200 '{"enabled":false}';
}

location = /proxy/users/access/api/v2/settings {
    include /usr/share/unifi-core/http/cors.conf;
    include /usr/share/unifi-core/http/security.conf;
    include /usr/share/unifi-core/http/auth.conf;

    default_type application/json;
    return 200 '{}';
}

location = /proxy/users/directory/api/v1/admin/ldap/config/base {
    include /usr/share/unifi-core/http/cors.conf;
    include /usr/share/unifi-core/http/security.conf;
    include /usr/share/unifi-core/http/auth.conf;

    default_type application/json;
    return 200 '{}';
}

location = /proxy/users/api/v2/org {
    include /usr/share/unifi-core/http/cors.conf;
    include /usr/share/unifi-core/http/security.conf;
    include /usr/share/unifi-core/http/auth.conf;

    default_type application/json;
    return 200 '{}';
}
NGINX

    chown unifi-core:unifi-core "$lab_config"
    nginx -s reload >/dev/null 2>&1 || true
}

enable_lab_http_preview() {
    local site_config="/data/unifi-core/config/http/site-local-ip.conf"
    local tmp_config

    if [[ ! -f "$site_config" ]] || grep -q "UDM Pro SE lab HTTP preview" "$site_config"; then
        return
    fi

    tmp_config="$(mktemp)"
    awk '
        /^[[:space:]]*return 301 https:\/\/\$host\$request_uri;/ {
            print "    # UDM Pro SE lab HTTP preview."
            print "    # Docker maps host 127.0.0.1:9080 to container port 80. A"
            print "    # stock redirect points the browser at host port 443, which"
            print "    # is not the mapped HTTPS port in this lab."
            print "    include /usr/share/unifi-core/http/errors.conf;"
            print "    include /usr/share/unifi-core/http/shared-server-defaults.conf;"
            print "    include /usr/share/unifi-core/http/shared-post-setup-server.conf;"
            next
        }
        { print }
    ' "$site_config" >"$tmp_config"
    cat "$tmp_config" >"$site_config"
    rm -f "$tmp_config"

    nginx -t >/dev/null && nginx -s reload >/dev/null 2>&1 || true
}

start_unifi_core() {
    if pgrep -f "/usr/share/unifi-core/app/service.js" >/dev/null 2>&1; then
        return
    fi

    USER=unifi-core GROUP=unifi-core /usr/share/unifi-core/app/hooks/pre-start \
        >"$log_dir/unifi-core-pre.log" \
        2>"$log_dir/unifi-core-pre.err"

    start_nginx

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

write_ubnt_tools_wrapper
write_ubnt_systool_wrapper
write_systemd_run_wrapper
write_systemctl_wrapper
write_lab_sudoers
write_timedatectl_wrapper
write_tar_wrapper
write_udapi_lab_wrappers
ensure_lab_lan_bridge
start_postgres
write_unifi_core_override
prepare_support_bundle_paths
start_dbus
start_systemd_stub
start_network_app_stub
start_ulp_go
start_unifi_core

exec /usr/local/bin/udm-pro-se-sim-start
