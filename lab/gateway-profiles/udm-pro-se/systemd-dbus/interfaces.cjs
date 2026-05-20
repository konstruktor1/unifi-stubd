const { ACCESS_READ, Interface } = require("./dbus.cjs");
const { unitFor } = require("./units.cjs");

class ManagerInterface extends Interface {
  constructor() {
    super("org.freedesktop.systemd1.Manager");
  }

  // UniFi Core subscribes before watching service state. There is no event
  // stream in this facade yet, so the method is accepted as a no-op.
  Subscribe() {}

  // Pair for Subscribe(). Keeping it explicit documents that unsubscribe calls
  // are expected and intentionally harmless.
  Unsubscribe() {}

  // LoadUnit is the only manager method currently needed by the webportal path.
  // It maps known services to deterministic object paths and unknown services to
  // the inactive lab placeholder from units.cjs.
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

  // ActiveState is the high-level state UniFi Core uses to decide whether a
  // service is running, inactive, or unavailable.
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

  // StatusText carries service-specific text. Only uos-agent currently needs a
  // structured JSON value; most lab services return an empty or abridged string.
  get StatusText() {
    return this.unit.statusText;
  }
}

ServiceInterface.configureMembers({
  properties: {
    StatusText: { signature: "s", access: ACCESS_READ },
  },
});

module.exports = {
  ManagerInterface,
  ServiceInterface,
  UnitInterface,
};
