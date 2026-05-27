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
#include <sys/socket.h>
#include <sys/stat.h>
#include <sys/uio.h>
#include <unistd.h>

/*
 * This LD_PRELOAD shim keeps firmware processes inside the lab container.
 * Selected hardware and sysctl reads are redirected into /mock, while a small
 * set of host-management commands is turned into no-ops.
 */
static char redirected_fds[4096];

static int patch_udapi_user_check_enabled(void) {
    const char *value = getenv("UXGPRO_SIM_PATCH_UDAPI_USER_CHECK");
    return value == NULL || strcmp(value, "0") != 0;
}

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

static int patch_lab_response(char *buf, size_t len) {
    int patched = 0;

    if (!patch_udapi_user_check_enabled() ||
        !byte_span_contains(buf, len, "/user/check") ||
        !byte_span_contains(buf, len, "A12")) {
        return 0;
    }

    patched |= byte_span_replace(buf, len, "\"error\":1", "\"error\":0");
    patched |= byte_span_replace(buf, len, "\"error\": 1", "\"error\": 0");
    patched |= byte_span_replace(buf, len, "\"statusCode\":500", "\"statusCode\":200");
    patched |= byte_span_replace(buf, len, "\"statusCode\": 500", "\"statusCode\": 200");

    if (patched && getenv("UBNTHAL_REDIRECT_DEBUG") != NULL) {
        fprintf(stderr, "ubnthal_redirect: patched UDAPI /user/check A12 response\n");
    }
    return patched;
}

static int response_patch_needed(const void *buf, size_t len) {
    return buf != NULL && len > 0 &&
        patch_udapi_user_check_enabled() &&
        byte_span_contains((const char *)buf, len, "/user/check") &&
        byte_span_contains((const char *)buf, len, "A12");
}

static ssize_t write_patched_buffer(
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
ssize_t write(int fd, const void *buf, size_t count) {
    static ssize_t (*real_write)(int, const void *, size_t) = NULL;
    if (real_write == NULL) {
        real_write = dlsym(RTLD_NEXT, "write");
    }
    return write_patched_buffer(real_write, fd, buf, count);
}

/* Some socket writers use send() instead of write(). */
ssize_t send(int sockfd, const void *buf, size_t len, int flags) {
    static ssize_t (*real_send)(int, const void *, size_t, int) = NULL;
    if (real_send == NULL) {
        real_send = dlsym(RTLD_NEXT, "send");
    }

    if (!response_patch_needed(buf, len)) {
        return real_send(sockfd, buf, len, flags);
    }

    char *copy = malloc(len);
    if (copy == NULL) {
        return real_send(sockfd, buf, len, flags);
    }
    memcpy(copy, buf, len);
    patch_lab_response(copy, len);
    ssize_t result = real_send(sockfd, copy, len, flags);
    free(copy);
    return result;
}

/* sendmsg() is another common Unix-socket response path. */
ssize_t sendmsg(int sockfd, const struct msghdr *msg, int flags) {
    static ssize_t (*real_sendmsg)(int, const struct msghdr *, int) = NULL;
    if (real_sendmsg == NULL) {
        real_sendmsg = dlsym(RTLD_NEXT, "sendmsg");
    }

    if (msg == NULL || msg->msg_iov == NULL || msg->msg_iovlen == 0) {
        return real_sendmsg(sockfd, msg, flags);
    }

    int patched_index = -1;
    for (size_t i = 0; i < msg->msg_iovlen; i++) {
        if (response_patch_needed(msg->msg_iov[i].iov_base, msg->msg_iov[i].iov_len)) {
            patched_index = (int)i;
            break;
        }
    }

    if (patched_index < 0) {
        return real_sendmsg(sockfd, msg, flags);
    }

    struct msghdr msg_copy = *msg;
    struct iovec *copy_iov = calloc(msg->msg_iovlen, sizeof(*copy_iov));
    if (copy_iov == NULL) {
        return real_sendmsg(sockfd, msg, flags);
    }
    memcpy(copy_iov, msg->msg_iov, msg->msg_iovlen * sizeof(*copy_iov));

    char *copy = malloc(msg->msg_iov[patched_index].iov_len);
    if (copy == NULL) {
        free(copy_iov);
        return real_sendmsg(sockfd, msg, flags);
    }
    memcpy(copy, msg->msg_iov[patched_index].iov_base, msg->msg_iov[patched_index].iov_len);
    patch_lab_response(copy, msg->msg_iov[patched_index].iov_len);
    copy_iov[patched_index].iov_base = copy;
    msg_copy.msg_iov = copy_iov;

    ssize_t result = real_sendmsg(sockfd, &msg_copy, flags);
    free(copy);
    free(copy_iov);
    return result;
}

/* Boost/asio may batch a single response payload through writev(). */
ssize_t writev(int fd, const struct iovec *iov, int iovcnt) {
    static ssize_t (*real_writev)(int, const struct iovec *, int) = NULL;
    if (real_writev == NULL) {
        real_writev = dlsym(RTLD_NEXT, "writev");
    }

    if (iov == NULL || iovcnt <= 0) {
        return real_writev(fd, iov, iovcnt);
    }

    int patched_index = -1;
    for (int i = 0; i < iovcnt; i++) {
        if (response_patch_needed(iov[i].iov_base, iov[i].iov_len)) {
            patched_index = i;
            break;
        }
    }

    if (patched_index < 0) {
        return real_writev(fd, iov, iovcnt);
    }

    struct iovec *copy_iov = calloc((size_t)iovcnt, sizeof(*copy_iov));
    if (copy_iov == NULL) {
        return real_writev(fd, iov, iovcnt);
    }
    memcpy(copy_iov, iov, (size_t)iovcnt * sizeof(*copy_iov));

    char *copy = malloc(iov[patched_index].iov_len);
    if (copy == NULL) {
        free(copy_iov);
        return real_writev(fd, iov, iovcnt);
    }
    memcpy(copy, iov[patched_index].iov_base, iov[patched_index].iov_len);
    patch_lab_response(copy, iov[patched_index].iov_len);
    copy_iov[patched_index].iov_base = copy;

    ssize_t result = real_writev(fd, copy_iov, iovcnt);
    free(copy);
    free(copy_iov);
    return result;
}

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
