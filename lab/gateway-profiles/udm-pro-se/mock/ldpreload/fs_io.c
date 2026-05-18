#include "ubnthal_redirect.h"

/*
 * Byte-stream and ioctl interposition.
 *
 * Response patching stays separate from path rewriting: this module only sees
 * file descriptors and buffers. The descriptor tracking comes from fs_paths.c.
 */

/* Patch only the lab-local UDAPI/auth/readiness responses while keeping servers real. */
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

/* sendmsg() is the other common Unix-socket response path. */
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

/* Debug redirected low-level reads without changing the returned bytes. */
ssize_t read(int fd, void *buf, size_t count) {
    static ssize_t (*real_read)(int, void *, size_t) = NULL;
    if (real_read == NULL) {
        real_read = dlsym(RTLD_NEXT, "read");
    }

    ssize_t n = real_read(fd, buf, count);
    if (n > 0 && fs_is_redirected_fd(fd) && debug_enabled()) {
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
    if (n > 0 && fs_is_redirected_fd(fd) && debug_enabled()) {
        size_t bytes = n * size;
        fprintf(stderr, "ubnthal_redirect: fread fd=%d bytes=%zu: ", fd, bytes);
        fwrite(ptr, 1, bytes, stderr);
        fputc('\n', stderr);
    }
    return n;
}

/* Log ioctl calls to identify the switch-driver boundary. */
int ioctl(int fd, unsigned long request, ...) {
    static int (*real_ioctl)(int, unsigned long, ...) = NULL;
    void *arg = NULL;
    const unsigned long mock_mtd_otpselect = 0x80044d0dUL;

    va_list ap;
    va_start(ap, request);
    arg = va_arg(ap, void *);
    va_end(ap);

    if (real_ioctl == NULL) {
        real_ioctl = dlsym(RTLD_NEXT, "ioctl");
    }

    if (fs_is_mtd_fd(fd) && request == mock_mtd_otpselect) {
        if (debug_enabled()) {
            fprintf(stderr, "ubnthal_redirect: simulated MTD_OTPSELECT fd=%d request=0x%lx\n", fd, request);
        }
        return 0;
    }

    int result = real_ioctl(fd, request, arg);
    int saved_errno = errno;
    if (debug_enabled()) {
        fprintf(stderr, "ubnthal_redirect: ioctl fd=%d request=0x%lx result=%d errno=%d\n",
            fd, request, result, result < 0 ? saved_errno : 0);
    }
    errno = saved_errno;
    return result;
}
