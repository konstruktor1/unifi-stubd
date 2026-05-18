#include "ubnthal_redirect.h"

/*
 * open/stat/access interposition.
 *
 * Keep the path rewrite boundary here: callers still use their normal libc
 * APIs, but firmware hardware paths are transparently served from /mock.
 */

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

    const char *original = pathname;
    const char *effective = fs_redirect_path(pathname);
    int fd = (flags & O_CREAT) != 0 ? real_open(effective, flags, mode) : real_open(effective, flags);
    int saved_errno = errno;
    fs_remember_fd(fd, effective);
    fs_log_path_result("open", original, effective, fd, saved_errno);
    errno = saved_errno;
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

    const char *original = pathname;
    const char *effective = fs_redirect_path(pathname);
    int fd = (flags & O_CREAT) != 0 ? real_open64(effective, flags, mode) : real_open64(effective, flags);
    int saved_errno = errno;
    fs_remember_fd(fd, effective);
    fs_log_path_result("open64", original, effective, fd, saved_errno);
    errno = saved_errno;
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

    const char *original = pathname;
    const char *effective = fs_redirect_path(pathname);
    int fd = (flags & O_CREAT) != 0 ? real_openat(dirfd, effective, flags, mode) : real_openat(dirfd, effective, flags);
    int saved_errno = errno;
    fs_remember_fd(fd, effective);
    fs_log_path_result("openat", original, effective, fd, saved_errno);
    errno = saved_errno;
    return fd;
}

/* Redirect stdio file opens used by libc helpers and shell utilities. */
FILE *fopen(const char *pathname, const char *mode) {
    static FILE *(*real_fopen)(const char *, const char *) = NULL;
    if (real_fopen == NULL) {
        real_fopen = dlsym(RTLD_NEXT, "fopen");
    }
    const char *original = pathname;
    const char *effective = fs_redirect_path(pathname);
    FILE *file = real_fopen(effective, mode);
    int saved_errno = errno;
    if (file != NULL) {
        fs_remember_fd(fileno(file), effective);
    }
    fs_log_path_result("fopen", original, effective, file != NULL ? fileno(file) : -1, saved_errno);
    errno = saved_errno;
    return file;
}

/* Match fopen64 callers on firmware builds that request large-file symbols. */
FILE *fopen64(const char *pathname, const char *mode) {
    static FILE *(*real_fopen64)(const char *, const char *) = NULL;
    if (real_fopen64 == NULL) {
        real_fopen64 = dlsym(RTLD_NEXT, "fopen64");
    }
    const char *original = pathname;
    const char *effective = fs_redirect_path(pathname);
    FILE *file = real_fopen64(effective, mode);
    int saved_errno = errno;
    if (file != NULL) {
        fs_remember_fd(fileno(file), effective);
    }
    fs_log_path_result("fopen64", original, effective, file != NULL ? fileno(file) : -1, saved_errno);
    errno = saved_errno;
    return file;
}

/* Redirect existence checks before the firmware decides which path to open. */
int access(const char *pathname, int mode) {
    static int (*real_access)(const char *, int) = NULL;
    if (real_access == NULL) {
        real_access = dlsym(RTLD_NEXT, "access");
    }
    const char *original = pathname;
    const char *effective = fs_redirect_path(pathname);
    int result = real_access(effective, mode);
    int saved_errno = errno;
    fs_log_path_result("access", original, effective, result, saved_errno);
    errno = saved_errno;
    return result;
}

/* Redirect metadata checks for mocked files. */
int stat(const char *pathname, struct stat *statbuf) {
    static int (*real_stat)(const char *, struct stat *) = NULL;
    if (real_stat == NULL) {
        real_stat = dlsym(RTLD_NEXT, "stat");
    }
    const char *original = pathname;
    const char *effective = fs_redirect_path(pathname);
    int result = real_stat(effective, statbuf);
    int saved_errno = errno;
    fs_log_path_result("stat", original, effective, result, saved_errno);
    errno = saved_errno;
    return result;
}

/* Keep symlink-aware metadata checks consistent with stat(). */
int lstat(const char *pathname, struct stat *statbuf) {
    static int (*real_lstat)(const char *, struct stat *) = NULL;
    if (real_lstat == NULL) {
        real_lstat = dlsym(RTLD_NEXT, "lstat");
    }
    const char *original = pathname;
    const char *effective = fs_redirect_path(pathname);
    int result = real_lstat(effective, statbuf);
    int saved_errno = errno;
    fs_log_path_result("lstat", original, effective, result, saved_errno);
    errno = saved_errno;
    return result;
}
