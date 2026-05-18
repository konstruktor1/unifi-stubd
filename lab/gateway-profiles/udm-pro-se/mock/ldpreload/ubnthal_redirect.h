#ifndef UBNTHAL_REDIRECT_H
#define UBNTHAL_REDIRECT_H

#ifndef _GNU_SOURCE
#define _GNU_SOURCE
#endif

#include <dlfcn.h>
#include <errno.h>
#include <fcntl.h>
#include <net/if.h>
#include <pwd.h>
#include <shadow.h>
#include <stdarg.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <strings.h>
#include <sys/stat.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <sys/uio.h>
#include <unistd.h>

#ifndef AF_NETLINK
#define AF_NETLINK (-1)
#endif

#ifndef IFNAMSIZ
#define IFNAMSIZ 16
#endif

#define SIM_SW_EDGE_PORTS 8
#define SIM_SW_CPU_PORT 9
#define SIM_SW_TOTAL_PORTS 10
#define SIM_SW_MAX_VLANS 4095
#define SIM_SW_EXTRA_ATTRS 256

/*
 * Minimal ABI-compatible declarations for the OpenWrt swconfig userspace
 * library. The UDM Pro SE server links libsw.so, so these symbols let the lab
 * replace the missing kernel switch driver with deterministic data.
 */
enum swlib_attr_group {
    SWLIB_ATTR_GROUP_GLOBAL,
    SWLIB_ATTR_GROUP_VLAN,
    SWLIB_ATTR_GROUP_PORT,
};

enum switch_val_type {
    SWITCH_TYPE_UNSPEC,
    SWITCH_TYPE_INT,
    SWITCH_TYPE_STRING,
    SWITCH_TYPE_PORTS,
    SWITCH_TYPE_LINK,
    SWITCH_TYPE_NOVAL,
};

enum swlib_port_flags {
    SWLIB_PORT_FLAG_TAGGED = (1 << 0),
};

struct switch_dev;
struct switch_attr;
struct switch_port;
struct switch_port_link;
struct switch_val;

struct switch_dev {
    int id;
    char dev_name[IFNAMSIZ];
    char *name;
    char *alias;
    int ports;
    int vlans;
    int cpu_port;
    struct switch_attr *ops;
    struct switch_attr *port_ops;
    struct switch_attr *vlan_ops;
    struct switch_portmap *maps;
    struct switch_dev *next;
    void *priv;
};

struct switch_val {
    struct switch_attr *attr;
    int len;
    int err;
    int port_vlan;
    union {
        char *s;
        int i;
        struct switch_port *ports;
        struct switch_port_link *link;
    } value;
};

struct switch_attr {
    struct switch_dev *dev;
    int atype;
    int id;
    int type;
    char *name;
    char *description;
    struct switch_attr *next;
};

struct switch_port {
    unsigned int id;
    unsigned int flags;
};

struct switch_portmap {
    unsigned int virt;
    char *segment;
};

struct switch_port_link {
    int link:1;
    int duplex:1;
    int aneg:1;
    int tx_flow:1;
    int rx_flow:1;
    int speed;
    uint32_t eee;
};

/* Environment-controlled feature gates shared by every interposition module. */
int debug_enabled(void);
int trace_all_enabled(void);
int udapi_user_check_enabled(void);
int network_status_ready_enabled(void);
int sfp_wan_primary_enabled(void);
int system_internet_ready_enabled(void);

/* Filesystem redirect helpers keep path policy centralized in fs_paths.c. */
const char *fs_redirect_path(const char *path);
void fs_log_path_result(const char *op, const char *original, const char *effective, long result, int saved_errno);
void fs_remember_fd(int fd, const char *path);
int fs_is_redirected_fd(int fd);
int fs_is_mtd_fd(int fd);

/* Response patch helpers are byte-preserving and shared by write/send hooks. */
int response_patch_needed(const void *buf, size_t len);
int patch_lab_response(char *buf, size_t len);
ssize_t write_patched_buffer(
    ssize_t (*writer)(int, const void *, size_t),
    int fd,
    const void *buf,
    size_t count
);

#endif
