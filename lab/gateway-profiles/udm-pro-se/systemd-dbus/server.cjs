const { busName, keepAliveMs, managerPath } = require("./config.cjs");
const { dbus } = require("./dbus.cjs");
const {
  ManagerInterface,
  ServiceInterface,
  UnitInterface,
} = require("./interfaces.cjs");
const { unitsForExport } = require("./units.cjs");

async function start() {
  const bus = dbus.systemBus();
  bus.on("error", (error) => console.error("systemd dbus stub error", error));

  // Claim the name before exporting objects so UniFi Core never observes a
  // half-initialized manager path.
  const reply = await bus.requestName(busName);
  console.log(`requestName ${busName} -> ${reply}`);

  bus.export(managerPath, new ManagerInterface());
  for (const unit of unitsForExport()) {
    // systemd exposes multiple interfaces on the same unit object path. Export
    // both here because UniFi Core reads generic Unit state and Service text.
    bus.export(unit.path, new UnitInterface(unit));
    bus.export(unit.path, new ServiceInterface(unit));
  }

  console.log("systemd dbus stub ready");
  setInterval(() => {}, keepAliveMs);
}

module.exports = {
  start,
};
