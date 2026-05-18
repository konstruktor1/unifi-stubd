#!/usr/bin/env node
// Lab UDAPI and management-data shim for the UDM Pro SE webportal profile.
//
// The reduced firmware container has Docker connectivity on eth0, but the
// stock UDAPI cache does not expose a UDM-style WAN port. UniFi Core treats
// that as "no physical connection". This shim maps the container's external
// interface to a deterministic WAN view and keeps all other commands delegated
// to the original firmware binaries.

const childProcess = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");

const tool = process.argv[2] || "";
const args = process.argv.slice(3);

// Prefer Docker's real outward-facing interface so UniFi Core can pass its
// internet checks, but fall back to documentation-safe data for offline tests.
function firstExternalInterface() {
  const interfaces = os.networkInterfaces();

  for (const [name, entries = []] of Object.entries(interfaces)) {
    const ipv4 = entries.find((entry) => entry.family === "IPv4" && !entry.internal);
    if (ipv4) {
      return {
        name,
        address: ipv4.address,
        cidr: ipv4.cidr || `${ipv4.address}/${prefixFromNetmask(ipv4.netmask)}`,
        mac: normalizeMac(ipv4.mac),
      };
    }
  }

  return {
    name: "eth0",
    address: "192.0.2.10",
    cidr: "192.0.2.10/24",
    mac: "02:15:6d:00:ea:2c",
  };
}

// Docker can report an all-zero MAC for synthetic interfaces. UniFi payloads
// expect a stable hardware identifier, so normalize that case to the lab MAC.
function normalizeMac(mac) {
  if (mac && mac !== "00:00:00:00:00:00") {
    return mac.toLowerCase();
  }
  return "02:15:6d:00:ea:2c";
}

// Node reports IPv4 netmasks separately on some platforms. Convert them into
// CIDR prefixes because UDAPI interface payloads use CIDR strings.
function prefixFromNetmask(netmask) {
  if (!netmask) {
    return 24;
  }
  return netmask
    .split(".")
    .map((octet) => Number(octet).toString(2).padStart(8, "0"))
    .join("")
    .replace(/0+$/, "").length;
}

// The webportal reads interface statistics, but the lab does not yet model
// traffic counters. Return explicit zeros instead of omitting fields.
function emptyStats() {
  return {
    dropped: 0,
    errors: 0,
    poePower: 0,
    rxBroadcast: 0,
    rxBytes: 0,
    rxDropped: 0,
    rxErrors: 0,
    rxFlowCtrl: 0,
    rxJumbo: 0,
    rxMulticast: 0,
    rxPPS: 0,
    rxPackets: 0,
    rxRate: 0,
    txBroadcast: 0,
    txBytes: 0,
    txDropped: 0,
    txErrors: 0,
    txFlowCtrl: 0,
    txJumbo: 0,
    txMulticast: 0,
    txPPS: 0,
    txPackets: 0,
    txRate: 0,
  };
}

// Shared link-state shape for WAN and LAN. The `comment` field is the human
// label UniFi Core shows in several local status decisions.
function baseStatus(comment, plugged) {
  return {
    arpProxy: false,
    comment,
    currentEnabled: true,
    currentFlowControl: null,
    currentMTU: 1500,
    currentSpeed: plugged ? "10000full" : null,
    enabled: true,
    mtu: 1500,
    plugged,
    speed: "auto",
    statistics: emptyStats(),
  };
}

// Expose one WAN mapped to Docker's external interface plus one deterministic
// LAN bridge. This is enough for Core's physical-link and local-console checks.
function labInterfaces() {
  const wan = firstExternalInterface();

  return [
    {
      addresses: [
        {
          cidr: wan.cidr,
          eui64: false,
          inUse: true,
          origin: "dhcp",
          type: "dynamic",
          version: "v4",
        },
      ],
      identification: {
        id: wan.name,
        mac: wan.mac,
        type: "ethernet",
      },
      ipv4: {
        cos: 0,
        dhcpOptions: [],
      },
      ipv6: {
        cos: 0,
        dhcp6Options: [],
        dhcp6OptionsOverrides: [],
        dhcp6PDStatus: [],
      },
      macTable: [],
      status: baseStatus("WAN", true),
    },
    {
      addresses: [
        {
          cidr: "192.168.1.1/24",
          eui64: false,
          inUse: true,
          origin: "static",
          type: "static",
          version: "v4",
        },
      ],
      bridge: {
        id: 0,
        interfaces: [],
      },
      identification: {
        id: "br0",
        mac: "02:15:6d:00:ea:2c",
        type: "bridge",
      },
      ipv4: {
        cos: 0,
        dhcpOptions: [],
      },
      ipv6: {
        cos: 0,
        dhcp6Options: [],
        dhcp6OptionsOverrides: [],
        dhcp6PDStatus: [],
      },
      macTable: [],
      status: baseStatus("LAN", true),
    },
  ];
}

// Core treats missing DNS data as an offline console. Static public resolvers
// are used here only as lab metadata; the shim does not change system DNS.
function labSystem() {
  const wan = firstExternalInterface();

  return {
    dnsServers: [
      {
        address: "1.1.1.1",
        associatedInterface: null,
        interface: wan.name,
        type: "static",
      },
      {
        address: "8.8.8.8",
        associatedInterface: null,
        interface: wan.name,
        type: "static",
      },
    ],
  };
}

// mca-dump geo data feeds ISP/location display paths in the setup UI. Keep it
// clearly labeled as local lab metadata.
function labMcaDump() {
  return {
    geo_info: {
      WAN: {
        city: "Local Lab",
        country_name: "Lab Network",
        isp_name: "Docker Bridge Internet",
      },
    },
  };
}

function printJson(value) {
  process.stdout.write(`${JSON.stringify(value, null, 2)}\n`);
}

// mca-ctrl uses `-t read-udapi-cache -s /interfaces` for cached interface
// reads. Match only that exact query and delegate every other mca-ctrl command.
function isReadUdapiInterfaces() {
  const tIndex = args.indexOf("-t");
  const sIndex = args.indexOf("-s");
  return (
    tIndex !== -1 &&
    args[tIndex + 1] === "read-udapi-cache" &&
    sIndex !== -1 &&
    args[sIndex + 1] === "/interfaces"
  );
}

// ubios-udapi-client accepts several argument forms. The last absolute path is
// the resource path for the GET calls this shim handles.
function clientPath() {
  const paths = args.filter((arg) => arg.startsWith("/"));
  return paths[paths.length - 1] || "";
}

// Restrict ubios-udapi-client handling to read-only GET operations. Mutating
// UDAPI commands stay delegated to the real firmware binary.
function shouldHandleClientGet() {
  return args.some((arg) => arg.toLowerCase() === "get");
}

// Anything outside the narrow lab surface is passed through to the original
// firmware tool, which start-webportal-processes.sh renames to *.real.
function runRealTool() {
  const realPath = `/usr/bin/${tool}.real`;
  if (!tool || !fs.existsSync(realPath)) {
    process.stderr.write(`unsupported lab shim command: ${tool} ${args.join(" ")}\n`);
    process.exit(127);
  }

  const result = childProcess.spawnSync(realPath, args, { stdio: "inherit" });
  if (result.signal) {
    process.kill(process.pid, result.signal);
    return;
  }
  process.exit(result.status ?? 1);
}

// Dispatch only the observed read paths needed by UniFi Core setup. Each branch
// exits immediately so unsupported calls cannot accidentally receive mock data.
if (tool === "mca-ctrl" && isReadUdapiInterfaces()) {
  printJson(labInterfaces());
  process.exit(0);
}

if (tool === "ubios-udapi-client" && shouldHandleClientGet()) {
  switch (clientPath()) {
    case "/interfaces":
      printJson(labInterfaces());
      process.exit(0);
      break;
    case "/system":
      printJson(labSystem());
      process.exit(0);
      break;
    default:
      break;
  }
}

if (tool === "mca-dump") {
  printJson(labMcaDump());
  process.exit(0);
}

runRealTool();
