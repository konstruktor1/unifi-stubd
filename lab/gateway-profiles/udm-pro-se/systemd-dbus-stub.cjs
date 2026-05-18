#!/usr/bin/env node
// Minimal systemd DBus facade for the UDM Pro SE webportal lab.
//
// UniFi Core only needs to subscribe to systemd, load uos-agent.service, and
// read the service state/status text during this lab startup path. Running a
// full systemd PID 1 inside the firmware container would make the simulation
// much less explicit, so this stub exposes only that narrow DBus surface.

const dbus = require("/usr/share/unifi-core/app/node_modules/@jellybrick/dbus-next");

const { Interface, ACCESS_READ } = dbus.interface;

const SERVICE_UNITS = new Map([
  [
    "uos-agent.service",
    {
      path: "/org/freedesktop/systemd1/unit/uos_2dagent_2eservice",
      activeState: "active",
      statusText: JSON.stringify({
        version: "0.0.0",
        fabric: {
          configLastUpdated: 0,
        },
      }),
    },
  ],
  [
    "unifi.service",
    {
      path: "/org/freedesktop/systemd1/unit/unifi_2eservice",
      activeState: "active",
      statusText: JSON.stringify({ abridged: true }),
    },
  ],
  [
    "udapi-server.service",
    {
      path: "/org/freedesktop/systemd1/unit/udapi_2dserver_2eservice",
      activeState: "active",
      statusText: JSON.stringify({ abridged: true }),
    },
  ],
  [
    "ulp-go.service",
    {
      path: "/org/freedesktop/systemd1/unit/ulp_2dgo_2eservice",
      activeState: "active",
      statusText: JSON.stringify({ abridged: true }),
    },
  ],
]);

const INACTIVE_SERVICE_NAMES = [
  "unifi-protect.service",
  "unifi-access.service",
  "unifi-talk.service",
  "unifi-talk-relay.service",
  "unifi-connect.service",
  "unifi-drive.service",
  "unifi-innerspace.service",
  "apollo.service",
  "uid-agent.service",
];

for (const name of INACTIVE_SERVICE_NAMES) {
  SERVICE_UNITS.set(name, {
    path: `/org/freedesktop/systemd1/unit/${encodeUnitName(name)}`,
    activeState: "inactive",
    statusText: "",
  });
}

const DEFAULT_UNIT = {
  path: "/org/freedesktop/systemd1/unit/lab_2dinactive_2eservice",
  activeState: "inactive",
  statusText: "",
};

function unitFor(name) {
  return SERVICE_UNITS.get(name) || DEFAULT_UNIT;
}

function encodeUnitName(name) {
  return name.replace(/[^A-Za-z0-9]/g, (char) => {
    const hex = char.charCodeAt(0).toString(16).padStart(2, "0");
    return `_${hex}`;
  });
}

const UOS_AGENT_STATUS_TEXT = JSON.stringify({
  version: "0.0.0",
  fabric: {
    configLastUpdated: 0,
  },
});

class ManagerInterface extends Interface {
  constructor() {
    super("org.freedesktop.systemd1.Manager");
  }

  Subscribe() {}

  Unsubscribe() {}

  LoadUnit(name) {
    const unit = unitFor(name);
    console.log(`LoadUnit ${name} -> ${unit.path}`);
    return unit.path;
  }
}

ManagerInterface.configureMembers({
  methods: {
    Subscribe: { inSignature: "", outSignature: "" },
    Unsubscribe: { inSignature: "", outSignature: "" },
    LoadUnit: { inSignature: "s", outSignature: "o" },
  },
  signals: {
    UnitFilesChanged: { signature: "" },
  },
});

class UnitInterface extends Interface {
  constructor(unit) {
    super("org.freedesktop.systemd1.Unit");
    this.unit = unit;
  }

  get ActiveState() {
    return this.unit.activeState;
  }
}

UnitInterface.configureMembers({
  properties: {
    ActiveState: { signature: "s", access: ACCESS_READ },
  },
});

class ServiceInterface extends Interface {
  constructor(unit) {
    super("org.freedesktop.systemd1.Service");
    this.unit = unit;
  }

  get StatusText() {
    return this.unit.statusText;
  }
}

ServiceInterface.configureMembers({
  properties: {
    StatusText: { signature: "s", access: ACCESS_READ },
  },
});

(async () => {
  const bus = dbus.systemBus();
  bus.on("error", (error) => console.error("systemd dbus stub error", error));

  const reply = await bus.requestName("org.freedesktop.systemd1");
  console.log(`requestName org.freedesktop.systemd1 -> ${reply}`);

  bus.export("/org/freedesktop/systemd1", new ManagerInterface());
  for (const unit of [
    ...SERVICE_UNITS.values(),
    {
      ...DEFAULT_UNIT,
      statusText: UOS_AGENT_STATUS_TEXT,
    },
  ]) {
    bus.export(unit.path, new UnitInterface(unit));
    bus.export(unit.path, new ServiceInterface(unit));
  }

  console.log("systemd dbus stub ready");
  setInterval(() => {}, 60000);
})().catch((error) => {
  console.error(error);
  process.exit(1);
});
