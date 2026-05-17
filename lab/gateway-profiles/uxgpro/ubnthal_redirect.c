/*
 * UXG-Pro firmware simulation LD_PRELOAD shim.
 * Keeps selected firmware reads and host-management calls inside the lab.
 */

#define _GNU_SOURCE

#include <dlfcn.h>
#include <fcntl.h>
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <unistd.h>

/*
 * This LD_PRELOAD shim keeps firmware processes inside the lab container.
 * Selected hardware and sysctl reads are redirected into /mock, while a small
 * set of host-management commands is turned into no-ops.
 */
static char redirected_fds[4096];

/* Return true for paths that already point at the mock tree. */
static int is_mock_path(const char *path) {
    const char *target = "/mock/ubnthal/";
    const char *proc_target = "/mock/proc/";

    return path != NULL &&
        (strncmp(path, target, strlen(target)) == 0 ||
         strncmp(path, proc_target, strlen(proc_target)) == 0);
}

/* Track redirected descriptors so debug reads can be attributed later. */
static void remember_fd(int fd, const char *path) {
    if (fd >= 0 && fd < (int)sizeof(redirected_fds) && is_mock_path(path)) {
        redirected_fds[fd] = 1;
    }
}

/* Map firmware hardware/sysctl reads to deterministic mock files. */
static const char *redirect_path(const char *path) {
    static __thread char redirected[512];
    const char *prefix = "/proc/ubnthal/";
    const char *target = "/mock/ubnthal/";
    const char *proc_prefix = "/proc/sys/";
    const char *proc_target = "/mock/proc/sys/";

    if (path == NULL) {
        return path;
    }

    if (strncmp(path, prefix, strlen(prefix)) == 0) {
        snprintf(redirected, sizeof(redirected), "%s%s", target, path + strlen(prefix));
    } else if (strncmp(path, proc_prefix, strlen(proc_prefix)) == 0) {
        snprintf(redirected, sizeof(redirected), "%s%s", proc_target, path + strlen(proc_prefix));
    } else {
        return path;
    }

    if (getenv("UBNTHAL_REDIRECT_DEBUG") != NULL) {
        fprintf(stderr, "ubnthal_redirect: %s -> %s\n", path, redirected);
    }
    return redirected;
}

/* Interpose open-style calls so firmware reads see the mock hardware tree. */
int open(const char *pathname, int flags, ...) {
    static int (*real_open)(const char *, int, ...) = NULL;
    mode_t mode = 0;

    if ((flags & O_CREAT) != 0) {
        va_list ap;
        va_start(ap, flags);
        mode = (mode_t)va_arg(ap, int);
        va_end(ap);
    }

    if (real_open == NULL) {
        real_open = dlsym(RTLD_NEXT, "open");
    }

    pathname = redirect_path(pathname);
    int fd = (flags & O_CREAT) != 0 ? real_open(pathname, flags, mode) : real_open(pathname, flags);
    remember_fd(fd, pathname);
    return fd;
}

/* Some firmware binaries call the large-file variant directly. */
int open64(const char *pathname, int flags, ...) {
    static int (*real_open64)(const char *, int, ...) = NULL;
    mode_t mode = 0;

    if ((flags & O_CREAT) != 0) {
        va_list ap;
        va_start(ap, flags);
        mode = (mode_t)va_arg(ap, int);
        va_end(ap);
    }

    if (real_open64 == NULL) {
        real_open64 = dlsym(RTLD_NEXT, "open64");
    }

    pathname = redirect_path(pathname);
    int fd = (flags & O_CREAT) != 0 ? real_open64(pathname, flags, mode) : real_open64(pathname, flags);
    remember_fd(fd, pathname);
    return fd;
}

/* Preserve relative openat behavior while redirecting absolute mock targets. */
int openat(int dirfd, const char *pathname, int flags, ...) {
    static int (*real_openat)(int, const char *, int, ...) = NULL;
    mode_t mode = 0;

    if ((flags & O_CREAT) != 0) {
        va_list ap;
        va_start(ap, flags);
        mode = (mode_t)va_arg(ap, int);
        va_end(ap);
    }

    if (real_openat == NULL) {
        real_openat = dlsym(RTLD_NEXT, "openat");
    }

    pathname = redirect_path(pathname);
    int fd = (flags & O_CREAT) != 0 ? real_openat(dirfd, pathname, flags, mode) : real_openat(dirfd, pathname, flags);
    remember_fd(fd, pathname);
    return fd;
}

/* Redirect stdio file opens used by libc helpers and shell utilities. */
FILE *fopen(const char *pathname, const char *mode) {
    static FILE *(*real_fopen)(const char *, const char *) = NULL;
    if (real_fopen == NULL) {
        real_fopen = dlsym(RTLD_NEXT, "fopen");
    }
    pathname = redirect_path(pathname);
    FILE *file = real_fopen(pathname, mode);
    if (file != NULL) {
        remember_fd(fileno(file), pathname);
    }
    return file;
}

/* Match fopen64 callers on firmware builds that request large-file symbols. */
FILE *fopen64(const char *pathname, const char *mode) {
    static FILE *(*real_fopen64)(const char *, const char *) = NULL;
    if (real_fopen64 == NULL) {
        real_fopen64 = dlsym(RTLD_NEXT, "fopen64");
    }
    pathname = redirect_path(pathname);
    FILE *file = real_fopen64(pathname, mode);
    if (file != NULL) {
        remember_fd(fileno(file), pathname);
    }
    return file;
}

/* Redirect existence checks before the firmware decides which path to open. */
int access(const char *pathname, int mode) {
    static int (*real_access)(const char *, int) = NULL;
    if (real_access == NULL) {
        real_access = dlsym(RTLD_NEXT, "access");
    }
    return real_access(redirect_path(pathname), mode);
}

/* Redirect metadata checks for mocked files. */
int stat(const char *pathname, struct stat *statbuf) {
    static int (*real_stat)(const char *, struct stat *) = NULL;
    if (real_stat == NULL) {
        real_stat = dlsym(RTLD_NEXT, "stat");
    }
    return real_stat(redirect_path(pathname), statbuf);
}

/* Keep symlink-aware metadata checks consistent with stat(). */
int lstat(const char *pathname, struct stat *statbuf) {
    static int (*real_lstat)(const char *, struct stat *) = NULL;
    if (real_lstat == NULL) {
        real_lstat = dlsym(RTLD_NEXT, "lstat");
    }
    return real_lstat(redirect_path(pathname), statbuf);
}

/* Debug redirected low-level reads without changing the returned bytes. */
ssize_t read(int fd, void *buf, size_t count) {
    static ssize_t (*real_read)(int, void *, size_t) = NULL;
    if (real_read == NULL) {
        real_read = dlsym(RTLD_NEXT, "read");
    }

    ssize_t n = real_read(fd, buf, count);
    if (n > 0 && fd >= 0 && fd < (int)sizeof(redirected_fds) && redirected_fds[fd] &&
        getenv("UBNTHAL_REDIRECT_DEBUG") != NULL) {
        fprintf(stderr, "ubnthal_redirect: read fd=%d bytes=%zd: ", fd, n);
        fwrite(buf, 1, (size_t)n, stderr);
        fputc('\n', stderr);
    }
    return n;
}

/* Debug redirected stdio reads without changing the returned bytes. */
size_t fread(void *ptr, size_t size, size_t nmemb, FILE *stream) {
    static size_t (*real_fread)(void *, size_t, size_t, FILE *) = NULL;
    if (real_fread == NULL) {
        real_fread = dlsym(RTLD_NEXT, "fread");
    }

    size_t n = real_fread(ptr, size, nmemb, stream);
    int fd = fileno(stream);
    if (n > 0 && fd >= 0 && fd < (int)sizeof(redirected_fds) && redirected_fds[fd] &&
        getenv("UBNTHAL_REDIRECT_DEBUG") != NULL) {
        size_t bytes = n * size;
        fprintf(stderr, "ubnthal_redirect: fread fd=%d bytes=%zu: ", fd, bytes);
        fwrite(ptr, 1, bytes, stderr);
        fputc('\n', stderr);
    }
    return n;
}

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
        if (getenv("UBNTHAL_REDIRECT_DEBUG") != NULL) {
            fprintf(stderr, "ubnthal_redirect: simulated system(\"%s\")\n", command);
        }
        return 0;
    }

    return real_system(command);
}

/* No-op direct execs of the same host-management commands. */
int execve(const char *pathname, char *const argv[], char *const envp[]) {
    static int (*real_execve)(const char *, char *const[], char *const[]) = NULL;
    if (real_execve == NULL) {
        real_execve = dlsym(RTLD_NEXT, "execve");
    }

    if (pathname != NULL &&
        (strcmp(pathname, "/bin/systemctl") == 0 ||
         strcmp(pathname, "/sbin/modprobe") == 0 ||
         strcmp(pathname, "/usr/sbin/modprobe") == 0)) {
        if (getenv("UBNTHAL_REDIRECT_DEBUG") != NULL) {
            fprintf(stderr, "ubnthal_redirect: simulated execve(\"%s\")\n", pathname);
        }
        _exit(0);
    }

    return real_execve(pathname, argv, envp);
}
