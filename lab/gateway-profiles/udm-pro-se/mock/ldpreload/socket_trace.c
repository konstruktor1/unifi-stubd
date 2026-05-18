#include "ubnthal_redirect.h"

/*
 * Socket boundary tracing.
 *
 * This does not change socket behavior. It records Unix-socket and netlink
 * activity so later firmware gaps can be traced without widening the mock.
 */

/* Log socket creation because switch access can be netlink or UNIX-socket based. */
int socket(int domain, int type, int protocol) {
    static int (*real_socket)(int, int, int) = NULL;
    if (real_socket == NULL) {
        real_socket = dlsym(RTLD_NEXT, "socket");
    }

    int fd = real_socket(domain, type, protocol);
    int saved_errno = errno;
    if (debug_enabled() && (trace_all_enabled() || fd < 0 || domain == AF_UNIX || domain == AF_NETLINK)) {
        fprintf(stderr, "ubnthal_redirect: socket domain=%d type=%d protocol=%d result=%d errno=%d\n",
            domain, type, protocol, fd, fd < 0 ? saved_errno : 0);
    }
    errno = saved_errno;
    return fd;
}

/* Log connection targets for UNIX sockets and failed control-plane sockets. */
int connect(int sockfd, const struct sockaddr *addr, socklen_t addrlen) {
    static int (*real_connect)(int, const struct sockaddr *, socklen_t) = NULL;
    if (real_connect == NULL) {
        real_connect = dlsym(RTLD_NEXT, "connect");
    }

    int result = real_connect(sockfd, addr, addrlen);
    int saved_errno = errno;
    if (debug_enabled() && addr != NULL && (trace_all_enabled() || result < 0 || addr->sa_family == AF_UNIX)) {
        if (addr->sa_family == AF_UNIX) {
            const struct sockaddr_un *un = (const struct sockaddr_un *)addr;
            fprintf(stderr, "ubnthal_redirect: connect fd=%d family=AF_UNIX path=\"%s\" result=%d errno=%d\n",
                sockfd, un->sun_path, result, result < 0 ? saved_errno : 0);
        } else {
            fprintf(stderr, "ubnthal_redirect: connect fd=%d family=%d result=%d errno=%d\n",
                sockfd, addr->sa_family, result, result < 0 ? saved_errno : 0);
        }
    }
    errno = saved_errno;
    return result;
}
