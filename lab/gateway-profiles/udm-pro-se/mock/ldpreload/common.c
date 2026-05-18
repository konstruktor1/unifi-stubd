#include "ubnthal_redirect.h"

/*
 * Shared feature flags for the shim.
 *
 * Every behavioral patch stays opt-in through an environment variable. This is
 * important because the same shared object is used in Docker and QEMU/UTM, but
 * not every compatibility patch is valid for both paths.
 */
static int env_flag(const char *name) {
    const char *value = getenv(name);
    return value != NULL && value[0] != '\0' &&
        strcmp(value, "0") != 0 &&
        strcasecmp(value, "false") != 0 &&
        strcasecmp(value, "no") != 0;
}

/* Return true when verbose shim logging is enabled. */
int debug_enabled(void) {
    return env_flag("UBNTHAL_REDIRECT_DEBUG");
}

/* Return true when all intercepted calls should be logged. */
int trace_all_enabled(void) {
    return env_flag("UBNTHAL_REDIRECT_TRACE_ALL");
}

/* Limit local-auth simulation to the UDAPI user resource in the QEMU VM. */
int udapi_user_check_enabled(void) {
    return env_flag("UNIFI_STUBD_VM_UDAPI_USER_CHECK");
}

/* Limit the Network setup-readiness patch to the QEMU VM lab. */
int network_status_ready_enabled(void) {
    return env_flag("UNIFI_STUBD_VM_NETWORK_STATUS_READY");
}

/* Limit the SFP-primary WAN role patch to the QEMU VM lab. */
int sfp_wan_primary_enabled(void) {
    return env_flag("UNIFI_STUBD_VM_SFP_WAN_PRIMARY");
}

/* Limit the UOS internet-state response patch to the QEMU VM lab. */
int system_internet_ready_enabled(void) {
    return env_flag("UNIFI_STUBD_VM_SYSTEM_INTERNET_READY");
}
