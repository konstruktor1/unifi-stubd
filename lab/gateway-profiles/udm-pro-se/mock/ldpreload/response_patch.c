#include "ubnthal_redirect.h"

/*
 * Byte-preserving response patches for narrow lab readiness gaps.
 *
 * The patch strings intentionally keep the same byte length as the originals.
 * That lets us adjust selected JSON booleans in write/send paths without
 * reframing HTTP responses or Unix-socket payloads.
 */
static int byte_span_contains(const char *buf, size_t len, const char *needle) {
    size_t needle_len = strlen(needle);
    if (needle_len == 0 || len < needle_len) {
        return 0;
    }
    for (size_t i = 0; i <= len - needle_len; i++) {
        if (memcmp(buf + i, needle, needle_len) == 0) {
            return 1;
        }
    }
    return 0;
}

static int byte_span_replace(char *buf, size_t len, const char *from, const char *to) {
    size_t from_len = strlen(from);
    size_t to_len = strlen(to);
    int replaced = 0;

    /*
     * Never change the payload length here. These buffers may already carry an
     * HTTP Content-Length or a Unix-socket frame length calculated by firmware
     * code we do not own.
     */
    if (from_len == 0 || from_len != to_len || len < from_len) {
        return 0;
    }
    for (size_t i = 0; i <= len - from_len; i++) {
        if (memcmp(buf + i, from, from_len) == 0) {
            memcpy(buf + i, to, to_len);
            replaced = 1;
            i += from_len - 1;
        }
    }
    return replaced;
}

static int patch_udapi_user_check_response(char *buf, size_t len) {
    int patched = 0;

    /*
     * UniFi Core asks UDAPI whether the local root credential is usable during
     * setup. In the lab we only flip the narrow A12 failure response, leaving
     * unrelated UDAPI errors visible.
     */
    if (!udapi_user_check_enabled() ||
        !byte_span_contains(buf, len, "/user/check") ||
        !byte_span_contains(buf, len, "A12")) {
        return 0;
    }

    patched |= byte_span_replace(buf, len, "\"error\":1", "\"error\":0");
    patched |= byte_span_replace(buf, len, "\"error\": 1", "\"error\": 0");
    patched |= byte_span_replace(buf, len, "\"statusCode\":500", "\"statusCode\":200");
    patched |= byte_span_replace(buf, len, "\"statusCode\": 500", "\"statusCode\": 200");

    if (patched && debug_enabled()) {
        fprintf(stderr, "ubnthal_redirect: patched UDAPI /user/check A12 response\n");
    }
    return patched;
}

static int patch_network_status_response(char *buf, size_t len) {
    int patched = 0;

    /*
     * The Network application gates setup on a self-readiness response. The VM
     * reference uses this to observe the next setup stages while the native
     * self-provisioning path is still being researched.
     */
    if (!network_status_ready_enabled() ||
        !byte_span_contains(buf, len, "\"isReadyForSetup\"")) {
        return 0;
    }

    patched |= byte_span_replace(buf, len, "\"udmConnected\":false", "\"udmConnected\":true ");
    patched |= byte_span_replace(buf, len, "\"udmConnected\": false", "\"udmConnected\": true ");
    patched |= byte_span_replace(buf, len, "\"isReadyForSetup\":false", "\"isReadyForSetup\":true ");
    patched |= byte_span_replace(buf, len, "\"isReadyForSetup\": false", "\"isReadyForSetup\": true ");
    patched |= byte_span_replace(
        buf,
        len,
        "\"udmProvisionCompleted\":false",
        "\"udmProvisionCompleted\":true "
    );
    patched |= byte_span_replace(
        buf,
        len,
        "\"udmProvisionCompleted\": false",
        "\"udmProvisionCompleted\": true "
    );

    if (patched && debug_enabled()) {
        fprintf(stderr, "ubnthal_redirect: patched Network /api/ucore/status readiness response\n");
    }
    return patched;
}

static int patch_interfaces_response(char *buf, size_t len) {
    int patched = 0;

    /*
     * The VM maps UTM Network 0 to the SFP+ WAN role. This preserves that role
     * in UDAPI interface summaries without needing a kernel switch driver.
     */
    if (!sfp_wan_primary_enabled()) {
        return 0;
    }

    patched |= byte_span_replace(buf, len, "\"comment\": \"WAN\"", "\"comment\": \"LAN\"");
    patched |= byte_span_replace(buf, len, "\"comment\": \"WAN2\"", "\"comment\": \"WAN1\"");

    if (patched && debug_enabled()) {
        fprintf(stderr, "ubnthal_redirect: patched UDAPI SFP WAN roles\n");
    }
    return patched;
}

static int patch_system_internet_response(char *buf, size_t len) {
    int patched = 0;

    if (!system_internet_ready_enabled() ||
        !byte_span_contains(buf, len, "\"hasInternet\"")) {
        return 0;
    }

    patched |= byte_span_replace(buf, len, "\"hasInternet\":false", "\"hasInternet\":true ");
    patched |= byte_span_replace(buf, len, "\"hasInternet\": false", "\"hasInternet\": true ");

    if (patched && debug_enabled()) {
        fprintf(stderr, "ubnthal_redirect: patched UOS system internet response\n");
    }
    return patched;
}

int response_patch_needed(const void *buf, size_t len) {
    if (buf == NULL || len == 0) {
        return 0;
    }

    if (udapi_user_check_enabled() &&
        byte_span_contains((const char *)buf, len, "/user/check") &&
        byte_span_contains((const char *)buf, len, "A12")) {
        return 1;
    }

    if (sfp_wan_primary_enabled() &&
        (byte_span_contains((const char *)buf, len, "\"comment\": \"WAN\"") ||
            byte_span_contains((const char *)buf, len, "\"comment\": \"WAN2\""))) {
        return 1;
    }

    if (system_internet_ready_enabled() &&
        (byte_span_contains((const char *)buf, len, "\"hasInternet\":false") ||
            byte_span_contains((const char *)buf, len, "\"hasInternet\": false"))) {
        return 1;
    }

    return network_status_ready_enabled() &&
        (byte_span_contains((const char *)buf, len, "\"isReadyForSetup\":false") ||
            byte_span_contains((const char *)buf, len, "\"isReadyForSetup\": false"));
}

int patch_lab_response(char *buf, size_t len) {
    int patched = 0;

    patched |= patch_udapi_user_check_response(buf, len);
    patched |= patch_network_status_response(buf, len);
    patched |= patch_interfaces_response(buf, len);
    patched |= patch_system_internet_response(buf, len);
    return patched;
}

ssize_t write_patched_buffer(
    ssize_t (*writer)(int, const void *, size_t),
    int fd,
    const void *buf,
    size_t count
) {
    if (!response_patch_needed(buf, count)) {
        return writer(fd, buf, count);
    }

    char *copy = malloc(count);
    if (copy == NULL) {
        return writer(fd, buf, count);
    }
    memcpy(copy, buf, count);
    patch_lab_response(copy, count);
    ssize_t result = writer(fd, copy, count);
    free(copy);
    return result;
}
