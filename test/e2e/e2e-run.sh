#!/usr/bin/env bash
# Copyright (c) 2026 Ant Group Corporation.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -Eeuo pipefail

SANDBOXD_HOME="${SANDBOXD_HOME:-/home/akernel/sandboxd}"
SANDBOXD_ROOT="${SANDBOXD_ROOT:-${SANDBOXD_HOME}/root}"
SANDBOXD_STORE="${SANDBOXD_STORE:-${SANDBOXD_HOME}/store}"
CONFIG_DIR="${SANDBOXD_HOME}/config"
CONFIG_FILE="${SANDBOXD_HOME}/config.toml"
SOCKET="${SANDBOXD_SOCKET:-/run/sandboxd/sandboxd.sock}"
LOG_FILE="${SANDBOXD_LOG_FILE:-/var/log/sandboxd/e2e.log}"
ROOTFS="${E2E_ROOTFS:-/e2e/rootfs}"
HOST_MOUNT="${E2E_HOST_MOUNT:-/e2e/host-mount}"
WWW_ROOT="${E2E_WWW_ROOT:-/e2e/www}"
CGROUP_ROOT="${E2E_CGROUP_ROOT:-sandboxd-e2e}"
NETWORK_CIDR="${E2E_NETWORK_CIDR:-10.88.0.1/16}"
GATEWAY_IP="${E2E_GATEWAY_IP:-10.88.0.1}"
HTTP_PORT="${E2E_HTTP_PORT:-18080}"
BRIDGE_NAME="${E2E_BRIDGE_NAME:-sandbox0}"

SANDBOXD_PID=""
HTTPD_PID=""
SANDBOX_ID=""

log() {
    printf '[e2e] %s\n' "$*"
}

fail() {
    printf '[e2e][error] %s\n' "$*" >&2
    exit 1
}

cleanup_cgroups() {
    local subsystem
    for subsystem in /sys/fs/cgroup/*; do
        if [ -d "${subsystem}/${CGROUP_ROOT}" ]; then
            find "${subsystem}/${CGROUP_ROOT}" -depth -type d -print 2>/dev/null | while read -r dir; do
                rmdir "${dir}" 2>/dev/null || true
            done
        fi
    done
}

cleanup() {
    local status=$?
    set +e

    if [ -n "${SANDBOX_ID}" ]; then
        /usr/local/bin/sbox --address "${SOCKET}" --timeout 20s delete "${SANDBOX_ID}" >/dev/null 2>&1
    fi
    if [ -n "${HTTPD_PID}" ]; then
        kill "${HTTPD_PID}" >/dev/null 2>&1
        wait "${HTTPD_PID}" >/dev/null 2>&1
    fi
    if [ -n "${SANDBOXD_PID}" ]; then
        kill "${SANDBOXD_PID}" >/dev/null 2>&1
        wait "${SANDBOXD_PID}" >/dev/null 2>&1
    fi
    cleanup_cgroups

    if [ "${status}" -ne 0 ] && [ -f "${LOG_FILE}" ]; then
        log "sandboxd log tail"
        tail -200 "${LOG_FILE}" >&2
    fi
}
trap cleanup EXIT

preflight() {
    [ "$(id -u)" = "0" ] || fail "e2e container must run as root"

    local bin
    for bin in sandboxd sbox runsc ip iptables busybox; do
        command -v "${bin}" >/dev/null 2>&1 || fail "missing command: ${bin}"
    done

    [ -d /sys/fs/cgroup/memory ] || fail "cgroup v1 memory hierarchy is required at /sys/fs/cgroup/memory"
    local probe="/sys/fs/cgroup/memory/${CGROUP_ROOT}-probe-$$"
    mkdir "${probe}" || fail "cannot create memory cgroup; run container with --privileged --cgroupns=host"
    rmdir "${probe}" || true

    iptables -t nat -L >/dev/null || fail "iptables nat table is not usable"
}

write_config() {
    mkdir -p "${CONFIG_DIR}" "${SANDBOXD_ROOT}" "${SANDBOXD_STORE}" "$(dirname "${SOCKET}")" "$(dirname "${LOG_FILE}")"

    cat > "${CONFIG_DIR}/oss.json" <<'EOF'
{
  "oss": {
    "access_key_id": "",
    "access_key_secret": "",
    "bucket": "",
    "connect_timeout": 300,
    "endpoint": "",
    "object_prefix": "/",
    "proxy": {
      "check_interval": 60,
      "fallback": true,
      "ping_url": "",
      "url": ""
    },
    "retry_limit": 3,
    "scheme": "http",
    "timeout": 300
  },
  "type": "oss"
}
EOF

    cat > "${CONFIG_DIR}/registry.json" <<'EOF'
{
  "registry": {
    "auth": "",
    "blob_url_scheme": "http",
    "connect_timeout": 300,
    "host": "",
    "proxy": {
      "check_interval": 60,
      "fallback": true,
      "ping_url": "",
      "url": ""
    },
    "repo": "",
    "scheme": "https",
    "skip_verify": true,
    "timeout": 300
  },
  "type": "registry"
}
EOF

    cat > "${CONFIG_DIR}/oss_auths.json" <<'EOF'
{}
EOF

    cat > "${CONFIG_DIR}/registry_auths.json" <<'EOF'
{"auths": {}}
EOF

    cat > "${CONFIG_FILE}" <<EOF
rootDir = "${SANDBOXD_ROOT}"
storeDir = "${SANDBOXD_STORE}"

[plugin.network]
ip_range = "${NETWORK_CIDR}"
nat_backend = "iptables"

[plugin.resource]
cgroup_cache_size = 1
interface_cache_size = 1
cgroup_root_name = "/${CGROUP_ROOT}"
max_instance_num = 8
recycle_policy = "destroy"

[plugin.runtime]
image_lib_dir = "/e2e/images"
overlay_tmpfs_size = "64M"

[plugin.runtime.basic_spec]
runsc = ""

[plugin.runtime.runtime_binary]
runsc = "/usr/local/bin/runsc"

[plugin.image]
root = "${SANDBOXD_HOME}/image_manager"
distill_fs_bin = "/usr/local/bin/distill_fs"
oss_template = "${CONFIG_DIR}/oss.json"
nydus_template = "${CONFIG_DIR}/registry.json"
nydus_suffix = "_nydus_v3"
oss_auths_path = "${CONFIG_DIR}/oss_auths.json"
registry_auths_path = "${CONFIG_DIR}/registry_auths.json"
cgroup_memory_limit = "0"
EOF
}

prepare_rootfs() {
    rm -rf "${ROOTFS}" "${HOST_MOUNT}" "${WWW_ROOT}"
    mkdir -p \
        "${ROOTFS}/bin" \
        "${ROOTFS}/dev" \
        "${ROOTFS}/etc" \
        "${ROOTFS}/mnt/host" \
        "${ROOTFS}/proc" \
        "${ROOTFS}/sys" \
        "${ROOTFS}/tmp" \
        "${ROOTFS}/usr/bin" \
        "${ROOTFS}/var" \
        "${HOST_MOUNT}" \
        "${WWW_ROOT}"
    chmod 1777 "${ROOTFS}/tmp"

    cp /bin/busybox "${ROOTFS}/bin/busybox"
    while read -r applet; do
        if [ "${applet}" = "busybox" ]; then
            continue
        fi
        ln -sf /bin/busybox "${ROOTFS}/bin/${applet}"
    done < <(/bin/busybox --list)

    cat > "${ROOTFS}/etc/hosts" <<'EOF'
127.0.0.1 localhost
EOF
    cat > "${ROOTFS}/etc/resolv.conf" <<'EOF'
nameserver 8.8.8.8
EOF

    echo "host-mount-ok" > "${HOST_MOUNT}/input.txt"
    echo "sandboxd-network-ok" > "${WWW_ROOT}/health.txt"
}

start_sandboxd() {
    log "starting sandboxd"
    /usr/local/bin/sandboxd \
        --root "${SANDBOXD_HOME}" \
        --config "${CONFIG_FILE}" \
        --socket "${SOCKET}" \
        --log-level debug \
        --log-file "${LOG_FILE}" \
        --http-address "127.0.0.1:23001" &
    SANDBOXD_PID=$!

    local i
    for i in $(seq 1 60); do
        if ! kill -0 "${SANDBOXD_PID}" >/dev/null 2>&1; then
            fail "sandboxd exited during startup"
        fi
        if [ -S "${SOCKET}" ] && /usr/local/bin/sbox --address "${SOCKET}" --timeout 5s check >/dev/null 2>&1; then
            log "sandboxd socket ready"
            return 0
        fi
        sleep 1
    done
    fail "sandboxd did not become ready"
}

start_gateway_httpd() {
    local i
    for i in $(seq 1 20); do
        if ip addr show "${BRIDGE_NAME}" >/dev/null 2>&1; then
            break
        fi
        sleep 1
    done
    ip addr show "${BRIDGE_NAME}" >/dev/null 2>&1 || fail "sandbox bridge ${BRIDGE_NAME} was not created"

    /bin/busybox httpd -f -p "${GATEWAY_IP}:${HTTP_PORT}" -h "${WWW_ROOT}" &
    HTTPD_PID=$!
}

sbox_cmd() {
    /usr/local/bin/sbox --address "${SOCKET}" --timeout 30s "$@"
}

assert_eq() {
    local got="$1"
    local want="$2"
    local name="$3"
    if [ "${got}" != "${want}" ]; then
        fail "${name}: got ${got@Q}, want ${want@Q}"
    fi
}

run_checks() {
    log "starting sandbox"
    SANDBOX_ID="$(sbox_cmd start \
        --quiet \
        --runtime runsc \
        --sandbox-id sbox-e2e-runtime \
        --rootfs "${ROOTFS}" \
        --cwd / \
        --env E2E_MARKER=start-env-ok \
        --mount "${HOST_MOUNT}:/mnt/host" \
        --cpu-millicores 100 \
        --memory-mb 128 \
        /bin/sh -c 'echo "$E2E_MARKER" > /tmp/start-env && sleep 300')"
    [ -n "${SANDBOX_ID}" ] || fail "start returned empty sandbox id"
    log "sandbox started: ${SANDBOX_ID}"

    local list_line
    list_line="$(sbox_cmd list | grep "${SANDBOX_ID}")" || fail "sandbox not found in list"
    echo "${list_line}" | grep -q "SANDBOX_STATE_RUNNING" || fail "sandbox is not running: ${list_line}"

    local got
    got="$(sbox_cmd exec "${SANDBOX_ID}" /bin/cat /tmp/start-env)"
    assert_eq "${got}" "start-env-ok" "start env"

    got="$(sbox_cmd exec "${SANDBOX_ID}" /bin/sh -c 'echo writable-ok > /tmp/e2e-write && cat /tmp/e2e-write')"
    assert_eq "${got}" "writable-ok" "writable overlay"

    got="$(sbox_cmd exec "${SANDBOX_ID}" /bin/cat /mnt/host/input.txt)"
    assert_eq "${got}" "host-mount-ok" "host bind mount read"

    sbox_cmd exec "${SANDBOX_ID}" /bin/sh -c 'echo from-sandbox > /mnt/host/from-sandbox.txt'
    got="$(cat "${HOST_MOUNT}/from-sandbox.txt")"
    assert_eq "${got}" "from-sandbox" "host bind mount write"

    got="$(sbox_cmd exec "${SANDBOX_ID}" /bin/wget -qO- "http://${GATEWAY_IP}:${HTTP_PORT}/health.txt")"
    assert_eq "${got}" "sandboxd-network-ok" "sandbox network"

    sbox_cmd stats "${SANDBOX_ID}" | grep -q "Memory Usage" || fail "stats output missing memory usage"

    log "deleting sandbox"
    sbox_cmd delete "${SANDBOX_ID}"
    local deleted_id="${SANDBOX_ID}"
    SANDBOX_ID=""
    if sbox_cmd inspect "${deleted_id}" >/tmp/sbox-inspect-after-delete.log 2>&1; then
        cat /tmp/sbox-inspect-after-delete.log >&2
        fail "sandbox still inspectable after delete"
    fi
}

main() {
    preflight
    cleanup_cgroups
    write_config
    prepare_rootfs
    start_sandboxd
    start_gateway_httpd
    run_checks
    log "e2e passed"
}

main "$@"
