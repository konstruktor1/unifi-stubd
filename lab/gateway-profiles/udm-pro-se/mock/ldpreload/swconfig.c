#include "ubnthal_redirect.h"

/*
 * Minimal RTL8370-style swconfig ABI.
 *
 * The UDM Pro SE userspace asks libsw.so for switch, VLAN, port, PoE, mirror,
 * isolation, link, and MIB data. QEMU does not emulate the real ASIC, so this
 * module returns deterministic lab values and accepts writes without mutating
 * host networking.
 */
static struct switch_dev sim_sw_dev;
static struct switch_attr sim_global_attrs[7];
static struct switch_attr sim_port_attrs[14];
static struct switch_attr sim_vlan_attrs[3];
static struct switch_attr sim_extra_attrs[SIM_SW_EXTRA_ATTRS];
static struct switch_port sim_vlan_ports[SIM_SW_EDGE_PORTS + 1];
static struct switch_port_link sim_port_links[SIM_SW_TOTAL_PORTS];
static int sim_extra_attr_count;
static int sim_sw_initialized;

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

    /*
     * The attribute set mirrors the names the firmware asks for during
     * network-init and UDAPI status generation. Values are intentionally stable:
     * the goal is to let userspace progress, not to simulate a mutable switch.
     */
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

    /*
     * Unknown links are represented as -1 until a getter asks for them. The CPU
     * port reports 10G, edge ports report 1G, matching the conservative status
     * expected by the current UDM userspace path.
     */
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
    /*
     * Cache dynamic attributes by name instead of allocating a new object per
     * lookup. Some firmware loops compare returned pointers and expect them to
     * remain valid for the process lifetime.
     */
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
