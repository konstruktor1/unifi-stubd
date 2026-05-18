#include "ubnthal_redirect.h"

/*
 * Host-management command containment.
 *
 * These calls can appear during firmware startup, but allowing a container or
 * VM lab process to control systemd, modprobe, reboot, or similar host paths is
 * outside the project safety boundary.
 */

/* Avoid container-side system management actions that are unsafe or useless. */
int system(const char *command) {
    static int (*real_system)(const char *) = NULL;
    if (real_system == NULL) {
        real_system = dlsym(RTLD_NEXT, "system");
    }

    if (command != NULL &&
        (strstr(command, "/bin/systemctl") != NULL ||
         strstr(command, "systemctl ") != NULL ||
         strstr(command, "modprobe ") != NULL)) {
        if (debug_enabled()) {
            fprintf(stderr, "ubnthal_redirect: simulated system(\"%s\")\n", command);
        }
        return 0;
    }

    return real_system(command);
}

/* Return true for direct exec targets that must stay outside the lab. */
static int is_blocked_exec_path(const char *pathname) {
    return pathname != NULL &&
        (strcmp(pathname, "/bin/systemctl") == 0 ||
         strcmp(pathname, "/sbin/modprobe") == 0 ||
         strcmp(pathname, "/usr/sbin/modprobe") == 0);
}

/* No-op direct execs of the same host-management commands. */
int execve(const char *pathname, char *const argv[], char *const envp[]) {
    static int (*real_execve)(const char *, char *const[], char *const[]) = NULL;
    if (real_execve == NULL) {
        real_execve = dlsym(RTLD_NEXT, "execve");
    }

    if (is_blocked_exec_path(pathname)) {
        if (debug_enabled()) {
            fprintf(stderr, "ubnthal_redirect: simulated execve(\"%s\")\n", pathname);
        }
        _exit(0);
    }

    return real_execve(pathname, argv, envp);
}
