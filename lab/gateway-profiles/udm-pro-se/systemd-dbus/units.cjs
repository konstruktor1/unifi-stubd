// systemd encodes non-alphanumeric unit-name characters in object paths as
// `_XX` hex escapes. The lab facade only needs enough of that rule for service
// names observed in UniFi Core's application catalog.
function encodeUnitName(name) {
  return name.replace(/[^A-Za-z0-9]/g, (char) => {
    const hex = char.charCodeAt(0).toString(16).padStart(2, "0");
    return `_${hex}`;
  });
}

// uos-agent is the one service where UniFi Core reads structured JSON from
// StatusText. The fields below are the minimum stable shape needed by the setup
// watcher; they intentionally do not describe a real fabric runtime.
const uosAgentStatusText = JSON.stringify({
  version: "0.0.0",
  fabric: {
    configLastUpdated: 0,
  },
});

// Active units are the services the webportal lab actually starts or fakes well
// enough for UniFi Core to continue setup. The paths match systemd's canonical
// DBus object path encoding for each service name.
const serviceUnits = new Map([
  [
    "uos-agent.service",
    {
      path: "/org/freedesktop/systemd1/unit/uos_2dagent_2eservice",
      activeState: "active",
      statusText: uosAgentStatusText,
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

const inactiveServiceNames = [
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

// These applications are present in UniFi Core's app catalog but are not
// started in the webportal lab. Returning inactive systemd units lets Core
// inspect them without pretending those services are really available.
for (const name of inactiveServiceNames) {
  serviceUnits.set(name, {
    path: `/org/freedesktop/systemd1/unit/${encodeUnitName(name)}`,
    activeState: "inactive",
    statusText: "",
  });
}

const defaultUnit = {
  path: "/org/freedesktop/systemd1/unit/lab_2dinactive_2eservice",
  activeState: "inactive",
  statusText: "",
};

// Unknown services should resolve to an inactive placeholder instead of failing
// DBus method calls. That keeps missing optional apps visible in the logs while
// avoiding a false claim that the lab implements them.
function unitFor(name) {
  return serviceUnits.get(name) || defaultUnit;
}

// Export every known unit path plus the shared inactive fallback. DBus clients
// may read both org.freedesktop.systemd1.Unit and .Service properties from the
// same object path, so server.cjs exports both interfaces for each entry.
function unitsForExport() {
  return [
    ...serviceUnits.values(),
    {
      ...defaultUnit,
      statusText: uosAgentStatusText,
    },
  ];
}

module.exports = {
  serviceUnits,
  unitFor,
  unitsForExport,
};
