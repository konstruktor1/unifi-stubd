// The facade claims only the canonical systemd bus name that UniFi Core asks
// for. Nothing here tries to emulate a full system manager.
const busName = "org.freedesktop.systemd1";

// UniFi Core calls methods on the manager object at this fixed systemd path.
const managerPath = "/org/freedesktop/systemd1";

// Keep the Node process alive after all DBus objects have been exported.
const keepAliveMs = 60000;

module.exports = {
  busName,
  keepAliveMs,
  managerPath,
};
