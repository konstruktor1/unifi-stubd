/*
 * UDM Pro SE firmware simulation LD_PRELOAD shim.
 * Keeps selected firmware reads and host-management calls inside the lab.
 */

#define _GNU_SOURCE

#include <dlfcn.h>
#include <errno.h>
#include <fcntl.h>
#include <net/if.h>
#include <stdarg.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <strings.h>
#include <sys/stat.h>
#include <sys/socket.h>
#include <sys/un.h>
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
 * This LD_PRELOAD shim keeps firmware processes inside the lab container.
 * Selected hardware and sysctl reads are redirected into /mock, while a small
 * set of host-management commands is turned into no-ops.
 */
static char redirected_fds[4096];
static char mtd_fds[4096];

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

static struct switch_dev sim_sw_dev;
static struct switch_attr sim_global_attrs[7];
static struct switch_attr sim_port_attrs[14];
static struct switch_attr sim_vlan_attrs[3];
static struct switch_attr sim_extra_attrs[SIM_SW_EXTRA_ATTRS];
static struct switch_port sim_vlan_ports[SIM_SW_EDGE_PORTS + 1];
static struct switch_port_link sim_port_links[SIM_SW_TOTAL_PORTS];
static int sim_extra_attr_count;
static int sim_sw_initialized;

/* Treat unset, empty, 0, false, and no as disabled feature flags. */
static int env_flag(const char *name) {
    const char *value = getenv(name);
    return value != NULL && value[0] != '\0' &&
        strcmp(value, "0") != 0 &&
        strcasecmp(value, "false") != 0 &&
        strcasecmp(value, "no") != 0;
}

/* Return true when verbose shim logging is enabled. */
static int debug_enabled(void) {
    return env_flag("UBNTHAL_REDIRECT_DEBUG");
}

/* Return true when all intercepted calls should be logged. */
static int trace_all_enabled(void) {
    return env_flag("UBNTHAL_REDIRECT_TRACE_ALL");
}

/* Link a static swconfig attribute list in the order firmware scans it. */
static void init_sim_attr(struct switch_attr *attr, struct switch_attr *next, int atype, int id, int type,
    char *name, char *description) {
    attr->dev = &sim_sw_dev;
    attr->atype = atype;
    attr->id = id;
    attr->type = type;
    attr->name = name;
    attr->description = description;
    attr->next = next;
}

/* Return a conservative type for unknown attributes requested by the firmware. */
static int guess_sim_attr_type(int atype, const char *name) {
    if (name == NULL) {
        return SWITCH_TYPE_INT;
    }
    if (strcmp(name, "ports") == 0 || strcmp(name, "isolation") == 0 ||
        (atype == SWLIB_ATTR_GROUP_PORT && strcmp(name, "mirror") == 0)) {
        return SWITCH_TYPE_PORTS;
    }
    if (strcmp(name, "link") == 0) {
        return SWITCH_TYPE_LINK;
    }
    if (strstr(name, "mib") != NULL) {
        return SWITCH_TYPE_STRING;
    }
    if (strcmp(name, "reset") == 0 || strcmp(name, "apply") == 0) {
        return SWITCH_TYPE_NOVAL;
    }
    if (strstr(name, "mode") != NULL || strstr(name, "name") != NULL || strstr(name, "alias") != NULL) {
        return SWITCH_TYPE_STRING;
    }
    if (atype == SWLIB_ATTR_GROUP_VLAN && strcmp(name, "vid") != 0) {
        return SWITCH_TYPE_INT;
    }
    return SWITCH_TYPE_INT;
}

/* Seed one deterministic RTL8370-like switch for the UDM Pro SE board profile. */
static void init_sim_switch(void) {
    if (sim_sw_initialized) {
        return;
    }

    memset(&sim_sw_dev, 0, sizeof(sim_sw_dev));
    sim_sw_dev.id = 0;
    strncpy(sim_sw_dev.dev_name, "switch0", sizeof(sim_sw_dev.dev_name) - 1);
    sim_sw_dev.name = "RTL8370";
    sim_sw_dev.alias = "switch0";
    sim_sw_dev.ports = SIM_SW_TOTAL_PORTS;
    sim_sw_dev.vlans = SIM_SW_MAX_VLANS;
    sim_sw_dev.cpu_port = SIM_SW_CPU_PORT;

    init_sim_attr(&sim_global_attrs[0], &sim_global_attrs[1], SWLIB_ATTR_GROUP_GLOBAL, 1, SWITCH_TYPE_INT,
        "enable_vlan", "simulated VLAN enable");
    init_sim_attr(&sim_global_attrs[1], &sim_global_attrs[2], SWLIB_ATTR_GROUP_GLOBAL, 2, SWITCH_TYPE_NOVAL,
        "apply", "simulated apply");
    init_sim_attr(&sim_global_attrs[2], &sim_global_attrs[3], SWLIB_ATTR_GROUP_GLOBAL, 3, SWITCH_TYPE_NOVAL,
        "reset", "simulated reset");
    init_sim_attr(&sim_global_attrs[3], &sim_global_attrs[4], SWLIB_ATTR_GROUP_GLOBAL, 4, SWITCH_TYPE_INT,
        "enable_mirror_rx", "simulated mirror RX");
    init_sim_attr(&sim_global_attrs[4], &sim_global_attrs[5], SWLIB_ATTR_GROUP_GLOBAL, 5, SWITCH_TYPE_INT,
        "enable_mirror_tx", "simulated mirror TX");
    init_sim_attr(&sim_global_attrs[5], &sim_global_attrs[6], SWLIB_ATTR_GROUP_GLOBAL, 6, SWITCH_TYPE_INT,
        "mirror_monitor_port", "simulated mirror monitor port");
    init_sim_attr(&sim_global_attrs[6], NULL, SWLIB_ATTR_GROUP_GLOBAL, 7, SWITCH_TYPE_INT,
        "mirror_source_port", "simulated mirror source port");

    init_sim_attr(&sim_port_attrs[0], &sim_port_attrs[1], SWLIB_ATTR_GROUP_PORT, 100, SWITCH_TYPE_LINK,
        "link", "simulated link status");
    init_sim_attr(&sim_port_attrs[1], &sim_port_attrs[2], SWLIB_ATTR_GROUP_PORT, 101, SWITCH_TYPE_INT,
        "pvid", "simulated port VLAN ID");
    init_sim_attr(&sim_port_attrs[2], &sim_port_attrs[3], SWLIB_ATTR_GROUP_PORT, 102, SWITCH_TYPE_INT,
        "enable", "simulated port enable");
    init_sim_attr(&sim_port_attrs[3], &sim_port_attrs[4], SWLIB_ATTR_GROUP_PORT, 103, SWITCH_TYPE_INT,
        "speed", "simulated port speed");
    init_sim_attr(&sim_port_attrs[4], &sim_port_attrs[5], SWLIB_ATTR_GROUP_PORT, 104, SWITCH_TYPE_INT,
        "duplex", "simulated port duplex");
    init_sim_attr(&sim_port_attrs[5], &sim_port_attrs[6], SWLIB_ATTR_GROUP_PORT, 105, SWITCH_TYPE_INT,
        "autoneg", "simulated port autonegotiation");
    init_sim_attr(&sim_port_attrs[6], &sim_port_attrs[7], SWLIB_ATTR_GROUP_PORT, 106, SWITCH_TYPE_INT,
        "txflow", "simulated TX flow control");
    init_sim_attr(&sim_port_attrs[7], &sim_port_attrs[8], SWLIB_ATTR_GROUP_PORT, 107, SWITCH_TYPE_INT,
        "rxflow", "simulated RX flow control");
    init_sim_attr(&sim_port_attrs[8], &sim_port_attrs[9], SWLIB_ATTR_GROUP_PORT, 108, SWITCH_TYPE_STRING,
        "poe_mode", "simulated PoE mode");
    init_sim_attr(&sim_port_attrs[9], &sim_port_attrs[10], SWLIB_ATTR_GROUP_PORT, 109, SWITCH_TYPE_INT,
        "poe_power", "simulated PoE power");
    init_sim_attr(&sim_port_attrs[10], &sim_port_attrs[11], SWLIB_ATTR_GROUP_PORT, 110, SWITCH_TYPE_INT,
        "poe", "simulated PoE state");
    init_sim_attr(&sim_port_attrs[11], &sim_port_attrs[12], SWLIB_ATTR_GROUP_PORT, 111, SWITCH_TYPE_PORTS,
        "isolation", "simulated port isolation list");
    init_sim_attr(&sim_port_attrs[12], &sim_port_attrs[13], SWLIB_ATTR_GROUP_PORT, 112, SWITCH_TYPE_PORTS,
        "mirror", "simulated port mirror list");
    init_sim_attr(&sim_port_attrs[13], NULL, SWLIB_ATTR_GROUP_PORT, 113, SWITCH_TYPE_INT,
        "isolate", "simulated port isolation");

    init_sim_attr(&sim_vlan_attrs[0], &sim_vlan_attrs[1], SWLIB_ATTR_GROUP_VLAN, 200, SWITCH_TYPE_PORTS,
        "ports", "simulated VLAN port list");
    init_sim_attr(&sim_vlan_attrs[1], &sim_vlan_attrs[2], SWLIB_ATTR_GROUP_VLAN, 201, SWITCH_TYPE_INT,
        "vid", "simulated VLAN ID");
    init_sim_attr(&sim_vlan_attrs[2], NULL, SWLIB_ATTR_GROUP_VLAN, 202, SWITCH_TYPE_INT,
        "fid", "simulated VLAN FID");

    sim_sw_dev.ops = sim_global_attrs;
    sim_sw_dev.port_ops = sim_port_attrs;
    sim_sw_dev.vlan_ops = sim_vlan_attrs;

    for (unsigned int i = 0; i < SIM_SW_EDGE_PORTS; i++) {
        sim_vlan_ports[i].id = i;
        sim_vlan_ports[i].flags = 0;
    }
    sim_vlan_ports[SIM_SW_EDGE_PORTS].id = SIM_SW_CPU_PORT;
    sim_vlan_ports[SIM_SW_EDGE_PORTS].flags = SWLIB_PORT_FLAG_TAGGED;

    for (int i = 0; i < SIM_SW_TOTAL_PORTS; i++) {
        sim_port_links[i].link = -1;
        sim_port_links[i].duplex = -1;
        sim_port_links[i].aneg = -1;
        sim_port_links[i].tx_flow = -1;
        sim_port_links[i].rx_flow = -1;
        sim_port_links[i].speed = i == SIM_SW_CPU_PORT ? 10000 : 1000;
        sim_port_links[i].eee = 0;
    }

    sim_sw_initialized = 1;
}

/* Find a named swconfig attribute in a linked static list. */
static struct switch_attr *find_sim_attr(struct switch_attr *attr, const char *name) {
    while (attr != NULL) {
        if (name != NULL && attr->name != NULL && strcmp(attr->name, name) == 0) {
            return attr;
        }
        attr = attr->next;
    }
    return NULL;
}

/* Return a stable generic attribute for firmware lookups not seen yet. */
static struct switch_attr *make_extra_sim_attr(int atype, const char *name) {
    for (int i = 0; i < sim_extra_attr_count; i++) {
        if (sim_extra_attrs[i].atype == atype && sim_extra_attrs[i].name != NULL &&
            name != NULL && strcmp(sim_extra_attrs[i].name, name) == 0) {
            return &sim_extra_attrs[i];
        }
    }

    if (sim_extra_attr_count >= SIM_SW_EXTRA_ATTRS) {
        return NULL;
    }

    struct switch_attr *attr = &sim_extra_attrs[sim_extra_attr_count++];
    memset(attr, 0, sizeof(*attr));
    attr->dev = &sim_sw_dev;
    attr->atype = atype;
    attr->id = 1000 + sim_extra_attr_count;
    attr->type = guess_sim_attr_type(atype, name);
    attr->name = name != NULL ? strdup(name) : NULL;
    attr->description = "simulated dynamic attribute";

    if (debug_enabled()) {
        fprintf(stderr, "ubnthal_redirect: swlib dynamic attr atype=%d name=\"%s\" type=%d\n",
            atype, name != NULL ? name : "(null)", attr->type);
    }
    return attr;
}

/* Return a deterministic integer value for the requested switch attribute. */
static int sim_attr_int_value(struct switch_attr *attr, int port_vlan) {
    if (attr == NULL || attr->name == NULL) {
        return 0;
    }
    if (strcmp(attr->name, "enable_vlan") == 0 || strcmp(attr->name, "enable") == 0 ||
        strcmp(attr->name, "duplex") == 0 || strcmp(attr->name, "autoneg") == 0 ||
        strcmp(attr->name, "txflow") == 0 || strcmp(attr->name, "rxflow") == 0) {
        return 1;
    }
    if (strcmp(attr->name, "speed") == 0) {
        return port_vlan == SIM_SW_CPU_PORT ? 10000 : 1000;
    }
    if (strcmp(attr->name, "pvid") == 0) {
        return port_vlan >= 0 && port_vlan < SIM_SW_EDGE_PORTS ? 4094 - port_vlan : 1;
    }
    if (strcmp(attr->name, "vid") == 0) {
        return port_vlan > 0 ? port_vlan : 1;
    }
    return 0;
}

/* Emulate libsw.so connection to the absent kernel switch driver. */
struct switch_dev *swlib_connect(const char *name) {
    init_sim_switch();
    if (name != NULL && name[0] != '\0') {
        snprintf(sim_sw_dev.dev_name, sizeof(sim_sw_dev.dev_name), "%s", name);
    }
    if (debug_enabled()) {
        fprintf(stderr, "ubnthal_redirect: swlib_connect name=\"%s\" -> %s ports=%d cpu=%d\n",
            name != NULL ? name : "(null)", sim_sw_dev.dev_name, sim_sw_dev.ports, sim_sw_dev.cpu_port);
    }
    return &sim_sw_dev;
}

/* The mock device is already populated, so scanning is a successful no-op. */
int swlib_scan(struct switch_dev *dev) {
    init_sim_switch();
    if (debug_enabled()) {
        fprintf(stderr, "ubnthal_redirect: swlib_scan dev=%p\n", (void *)dev);
    }
    return dev == NULL ? -1 : 0;
}

/* Resolve known switch, port, and VLAN attributes for the firmware. */
struct switch_attr *swlib_lookup_attr(struct switch_dev *dev, enum swlib_attr_group atype, const char *name) {
    init_sim_switch();
    struct switch_attr *attr = NULL;

    if (dev == NULL) {
        return NULL;
    }
    if (atype == SWLIB_ATTR_GROUP_GLOBAL) {
        attr = find_sim_attr(dev->ops, name);
    } else if (atype == SWLIB_ATTR_GROUP_PORT) {
        attr = find_sim_attr(dev->port_ops, name);
    } else if (atype == SWLIB_ATTR_GROUP_VLAN) {
        attr = find_sim_attr(dev->vlan_ops, name);
    }
    if (attr == NULL) {
        attr = make_extra_sim_attr(atype, name);
    }
    if (debug_enabled()) {
        fprintf(stderr, "ubnthal_redirect: swlib_lookup_attr atype=%d name=\"%s\" -> %p type=%d\n",
            atype, name != NULL ? name : "(null)", (void *)attr, attr != NULL ? attr->type : -1);
    }
    return attr;
}

/* Return stable switch values so status generation can run without hardware. */
int swlib_get_attr(struct switch_dev *dev, struct switch_attr *attr, struct switch_val *val) {
    init_sim_switch();
    if (dev == NULL || attr == NULL || val == NULL) {
        errno = EINVAL;
        return -1;
    }

    val->attr = attr;
    val->err = 0;

    switch (attr->type) {
    case SWITCH_TYPE_STRING:
        if (attr->name != NULL && strstr(attr->name, "mib") != NULL) {
            val->value.s = strdup("rxBytes: 0\ntxBytes: 0\nrxPackets: 0\ntxPackets: 0\nrxErrors: 0\ntxErrors: 0\n");
        } else {
            val->value.s = strdup(strcmp(attr->name, "poe_mode") == 0 ? "off" : "simulated");
        }
        if (val->value.s == NULL) {
            errno = ENOMEM;
            return -1;
        }
        break;
    case SWITCH_TYPE_PORTS: {
        struct switch_port *ports = calloc(SIM_SW_EDGE_PORTS + 1, sizeof(*ports));
        if (ports == NULL) {
            errno = ENOMEM;
            return -1;
        }
        memcpy(ports, sim_vlan_ports, sizeof(sim_vlan_ports));
        val->len = SIM_SW_EDGE_PORTS + 1;
        val->value.ports = ports;
        break;
    }
    case SWITCH_TYPE_LINK: {
        int port = val->port_vlan;
        if (port < 0 || port >= SIM_SW_TOTAL_PORTS) {
            port = 0;
        }
        struct switch_port_link *link = calloc(1, sizeof(*link));
        if (link == NULL) {
            errno = ENOMEM;
            return -1;
        }
        *link = sim_port_links[port];
        val->value.link = link;
        break;
    }
    case SWITCH_TYPE_NOVAL:
        val->value.i = 0;
        break;
    case SWITCH_TYPE_INT:
    default:
        val->value.i = sim_attr_int_value(attr, val->port_vlan);
        break;
    }

    if (debug_enabled()) {
        fprintf(stderr, "ubnthal_redirect: swlib_get_attr atype=%d name=\"%s\" port_vlan=%d type=%d\n",
            attr->atype, attr->name != NULL ? attr->name : "(null)", val->port_vlan, attr->type);
    }
    return 0;
}

/* Accept configuration writes in the lab, but do not mutate host networking. */
int swlib_set_attr(struct switch_dev *dev, struct switch_attr *attr, struct switch_val *val) {
    (void)dev;
    if (debug_enabled()) {
        fprintf(stderr, "ubnthal_redirect: swlib_set_attr name=\"%s\" port_vlan=%d\n",
            attr != NULL && attr->name != NULL ? attr->name : "(null)", val != NULL ? val->port_vlan : -1);
    }
    if (val != NULL) {
        val->err = 0;
    }
    return 0;
}

/* Accept string-based writes such as swconfig would, keeping them lab-local. */
int swlib_set_attr_string(struct switch_dev *dev, struct switch_attr *attr, int port_vlan, const char *str) {
    (void)dev;
    if (debug_enabled()) {
        fprintf(stderr, "ubnthal_redirect: swlib_set_attr_string name=\"%s\" port_vlan=%d value=\"%s\"\n",
            attr != NULL && attr->name != NULL ? attr->name : "(null)", port_vlan, str != NULL ? str : "(null)");
    }
    return 0;
}

/* swlib_free is a no-op because the mock uses static process-local storage. */
void swlib_free(struct switch_dev *dev) {
    if (debug_enabled()) {
        fprintf(stderr, "ubnthal_redirect: swlib_free dev=%p\n", (void *)dev);
    }
}

/* Free a possible switch list; the mock exposes a single static switch. */
void swlib_free_all(struct switch_dev *dev) {
    if (debug_enabled()) {
        fprintf(stderr, "ubnthal_redirect: swlib_free_all dev=%p\n", (void *)dev);
    }
}

/* Print a compact switch inventory for binaries that use swconfig helpers. */
void swlib_list(void) {
    init_sim_switch();
    printf("Found: %s - %s\n", sim_sw_dev.dev_name, sim_sw_dev.name);
}

/* Print deterministic port mappings for diagnostics. */
void swlib_print_portmap(struct switch_dev *dev, char *segment) {
    (void)segment;
    init_sim_switch();
    if (dev == NULL) {
        return;
    }
    for (int i = 0; i < SIM_SW_EDGE_PORTS; i++) {
        printf("port%d:\t%s.%d\n", i, dev->dev_name, i);
    }
    printf("port%d:\tcpu\n", SIM_SW_CPU_PORT);
}

/* Return true for paths that already point at the mock tree. */
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
static void log_path_result(const char *op, const char *original, const char *effective, long result, int saved_errno) {
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

/* Track redirected descriptors so debug reads can be attributed later. */
static void remember_fd(int fd, const char *path) {
    if (fd >= 0 && fd < (int)sizeof(redirected_fds) && is_mock_path(path)) {
        redirected_fds[fd] = 1;
    }
    if (fd >= 0 && fd < (int)sizeof(mtd_fds) && is_mock_mtd_path(path)) {
        mtd_fds[fd] = 1;
    }
}

/* Map firmware hardware/sysctl reads to deterministic mock files. */
static const char *redirect_path(const char *path) {
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
    } else if (strncmp(path, sys_prefix, strlen(sys_prefix)) == 0) {
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
    const char *effective = redirect_path(pathname);
    int fd = (flags & O_CREAT) != 0 ? real_open(effective, flags, mode) : real_open(effective, flags);
    int saved_errno = errno;
    remember_fd(fd, effective);
    log_path_result("open", original, effective, fd, saved_errno);
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
    const char *effective = redirect_path(pathname);
    int fd = (flags & O_CREAT) != 0 ? real_open64(effective, flags, mode) : real_open64(effective, flags);
    int saved_errno = errno;
    remember_fd(fd, effective);
    log_path_result("open64", original, effective, fd, saved_errno);
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
    const char *effective = redirect_path(pathname);
    int fd = (flags & O_CREAT) != 0 ? real_openat(dirfd, effective, flags, mode) : real_openat(dirfd, effective, flags);
    int saved_errno = errno;
    remember_fd(fd, effective);
    log_path_result("openat", original, effective, fd, saved_errno);
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
    const char *effective = redirect_path(pathname);
    FILE *file = real_fopen(effective, mode);
    int saved_errno = errno;
    if (file != NULL) {
        remember_fd(fileno(file), effective);
    }
    log_path_result("fopen", original, effective, file != NULL ? fileno(file) : -1, saved_errno);
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
    const char *effective = redirect_path(pathname);
    FILE *file = real_fopen64(effective, mode);
    int saved_errno = errno;
    if (file != NULL) {
        remember_fd(fileno(file), effective);
    }
    log_path_result("fopen64", original, effective, file != NULL ? fileno(file) : -1, saved_errno);
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
    const char *effective = redirect_path(pathname);
    int result = real_access(effective, mode);
    int saved_errno = errno;
    log_path_result("access", original, effective, result, saved_errno);
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
    const char *effective = redirect_path(pathname);
    int result = real_stat(effective, statbuf);
    int saved_errno = errno;
    log_path_result("stat", original, effective, result, saved_errno);
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
    const char *effective = redirect_path(pathname);
    int result = real_lstat(effective, statbuf);
    int saved_errno = errno;
    log_path_result("lstat", original, effective, result, saved_errno);
    errno = saved_errno;
    return result;
}

/* Debug redirected low-level reads without changing the returned bytes. */
ssize_t read(int fd, void *buf, size_t count) {
    static ssize_t (*real_read)(int, void *, size_t) = NULL;
    if (real_read == NULL) {
        real_read = dlsym(RTLD_NEXT, "read");
    }

    ssize_t n = real_read(fd, buf, count);
    if (n > 0 && fd >= 0 && fd < (int)sizeof(redirected_fds) && redirected_fds[fd] &&
        debug_enabled()) {
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
        debug_enabled()) {
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

    if (fd >= 0 && fd < (int)sizeof(mtd_fds) && mtd_fds[fd] && request == mock_mtd_otpselect) {
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
