// UniFi Core ships the DBus client library inside its own Node dependency tree.
// Use that copy so the facade runs in the imported firmware rootfs without
// installing project-managed npm packages.
const dbus = require("/usr/share/unifi-core/app/node_modules/@jellybrick/dbus-next");

const { Interface, ACCESS_READ } = dbus.interface;

// Re-export the small DBus surface used by the facade so the interface modules
// do not need to know where the firmware stores its bundled Node dependency.
module.exports = {
  ACCESS_READ,
  Interface,
  dbus,
};
