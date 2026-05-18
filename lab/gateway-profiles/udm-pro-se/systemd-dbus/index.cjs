#!/usr/bin/env node
// Minimal systemd DBus facade for the UDM Pro SE webportal lab.
//
// UniFi Core only needs to subscribe to systemd, load a known set of services,
// and read service state/status text during this lab startup path. Running a
// full systemd PID 1 inside the firmware container would make the simulation
// less explicit, so this facade exposes only that narrow DBus surface.

const { start } = require("./server.cjs");

start().catch((error) => {
  console.error(error);
  process.exit(1);
});
