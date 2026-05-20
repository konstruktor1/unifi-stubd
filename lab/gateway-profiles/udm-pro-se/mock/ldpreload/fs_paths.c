#include "ubnthal_redirect.h"

/*
 * Path redirection and descriptor tracking.
 *
 * Firmware reads for board identity, EEPROM/MTD, selected sysfs classes, and
 * appliance persistence are redirected into /mock. The descriptor bitmaps let
 * other interposition modules attribute later read/ioctl calls to those mocked
 * paths without sharing filesystem policy there.
 */
static char redirected_fds[4096];
static char mtd_fds[4096];

static int is_mock_path(const char *path) {
    const char *target = "/mock/ubnthal/";
    const char *proc_target = "/mock/proc/";
    const char *mtd_target = "/mock/mtd/";
    const char *sys_target = "/mock/sys/";
    const char *persistent_target = "/mock/persistent/";

    return path != NULL &&
        (strncmp(path, target, strlen(target)) == 0 ||
         strncmp(path, proc_target, strlen(proc_target)) == 0 ||
         strncmp(path, mtd_target, strlen(mtd_target)) == 0 ||
         strncmp(path, sys_target, strlen(sys_target)) == 0 ||
         strncmp(path, persistent_target, strlen(persistent_target)) == 0);
}

/* Return true for mocked MTD devices used by the board EEPROM path. */
static int is_mock_mtd_path(const char *path) {
    return path != NULL &&
        (strcmp(path, "/mock/mtd/mtd5") == 0 ||
         strcmp(path, "/mock/mtd/mtdblock5") == 0);
}

/* Return true only for sysfs classes backed by deterministic mock files. */
static int is_mocked_sys_class_path(const char *path) {
    static const char *prefixes[] = {
        "/sys/class/hwmon/",
        "/sys/class/mtd/",
        "/sys/class/thermal/",
        NULL,
    };

    if (path == NULL) {
        return 0;
    }

    for (size_t i = 0; prefixes[i] != NULL; i++) {
        if (strncmp(path, prefixes[i], strlen(prefixes[i])) == 0) {
            return 1;
        }
    }
    return 0;
}

/* Return true for hardware and runtime paths that matter to the switch stack. */
static int is_interesting_path(const char *path) {
    static const char *prefixes[] = {
        "/proc/ubnthal/",
        "/mock/ubnthal/",
        "/proc/sys/",
        "/mock/proc/",
        "/mock/mtd/",
        "/mock/sys/",
        "/mock/persistent/",
        "/dev/",
        "/sys/",
        "/run/",
        "/var/run/",
        "/data/udapi-config/",
        NULL,
    };

    if (path == NULL) {
        return 0;
    }

    for (size_t i = 0; prefixes[i] != NULL; i++) {
        if (strncmp(path, prefixes[i], strlen(prefixes[i])) == 0) {
            return 1;
        }
    }
    return 0;
}

/* Emit a compact path syscall trace while preserving errno for the firmware. */
void fs_log_path_result(const char *op, const char *original, const char *effective, long result, int saved_errno) {
    if (!debug_enabled()) {
        return;
    }

    if (!trace_all_enabled() && result >= 0 && !is_interesting_path(original) && !is_interesting_path(effective)) {
        return;
    }

    if (original != NULL && effective != NULL && strcmp(original, effective) != 0) {
        fprintf(stderr, "ubnthal_redirect: %s path=\"%s\" effective=\"%s\" result=%ld errno=%d\n",
            op, original, effective, result, result < 0 ? saved_errno : 0);
    } else {
        fprintf(stderr, "ubnthal_redirect: %s path=\"%s\" result=%ld errno=%d\n",
            op, effective != NULL ? effective : "(null)", result, result < 0 ? saved_errno : 0);
    }
}

/* Track redirected descriptors so debug reads and MTD ioctls can be handled. */
void fs_remember_fd(int fd, const char *path) {
    if (fd >= 0 && fd < (int)sizeof(redirected_fds) && is_mock_path(path)) {
        redirected_fds[fd] = 1;
    }
    if (fd >= 0 && fd < (int)sizeof(mtd_fds) && is_mock_mtd_path(path)) {
        mtd_fds[fd] = 1;
    }
}

int fs_is_redirected_fd(int fd) {
    return fd >= 0 && fd < (int)sizeof(redirected_fds) && redirected_fds[fd];
}

int fs_is_mtd_fd(int fd) {
    return fd >= 0 && fd < (int)sizeof(mtd_fds) && mtd_fds[fd];
}

/* Map firmware hardware/sysctl reads to deterministic mock files. */
const char *fs_redirect_path(const char *path) {
    static __thread char redirected[512];
    const char *prefix = "/proc/ubnthal/";
    const char *target = "/mock/ubnthal/";
    const char *proc_prefix = "/proc/sys/";
    const char *proc_target = "/mock/proc/sys/";
    const char *sys_prefix = "/sys/class/";
    const char *sys_target = "/mock/sys/class/";
    const char *persistent_prefix = "/etc/persistent/";
    const char *persistent_target = "/mock/persistent/";

    if (path == NULL) {
        return path;
    }

    if (strncmp(path, prefix, strlen(prefix)) == 0) {
        snprintf(redirected, sizeof(redirected), "%s%s", target, path + strlen(prefix));
    } else if (strncmp(path, proc_prefix, strlen(proc_prefix)) == 0) {
        snprintf(redirected, sizeof(redirected), "%s%s", proc_target, path + strlen(proc_prefix));
    } else if (strcmp(path, "/proc/mtd") == 0) {
        snprintf(redirected, sizeof(redirected), "%s", "/mock/mtd/proc_mtd");
    } else if (strcmp(path, "/dev/mtdblock5") == 0) {
        snprintf(redirected, sizeof(redirected), "%s", "/mock/mtd/mtdblock5");
    } else if (strcmp(path, "/dev/mtd5") == 0) {
        snprintf(redirected, sizeof(redirected), "%s", "/mock/mtd/mtd5");
    } else if (strcmp(path, "/etc/board.info") == 0) {
        snprintf(redirected, sizeof(redirected), "%s", "/mock/ubnthal/board");
    } else if (is_mocked_sys_class_path(path)) {
        snprintf(redirected, sizeof(redirected), "%s%s", sys_target, path + strlen(sys_prefix));
    } else if (strncmp(path, persistent_prefix, strlen(persistent_prefix)) == 0) {
        snprintf(redirected, sizeof(redirected), "%s%s", persistent_target, path + strlen(persistent_prefix));
    } else {
        return path;
    }

    if (debug_enabled()) {
        fprintf(stderr, "ubnthal_redirect: %s -> %s\n", path, redirected);
    }
    return redirected;
}
