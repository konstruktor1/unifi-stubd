#!/usr/bin/env node
// Minimal UniFi Network API facade for the UDM Pro SE webportal lab.
//
// The real Network application is not started in this firmware profile. UniFi
// Core still calls a small set of Network endpoints while finishing setup, so
// this facade records those calls and returns deterministic no-op responses.

const http = require("node:http");

const { host, port } = require("./config.cjs");
const { readBody, sendJson } = require("./http.cjs");
const { log } = require("./logger.cjs");
const { route } = require("./routes.cjs");
const { acceptWebSocket } = require("./websocket.cjs");

// Keep request handling centralized: index.cjs owns transport concerns, while
// routes.cjs decides which deterministic payload belongs to a UniFi endpoint.
const server = http.createServer(async (req, res) => {
  try {
    const body = await readBody(req);
    log(`${req.method} ${req.url || "/"} body=${body}`);
    const { status, payload } = route(req, body);
    sendJson(res, status, payload);
  } catch (error) {
    log(`error ${error && error.stack ? error.stack : String(error)}`);
    sendJson(res, 500, { errors: [String(error)] });
  }
});

// The frontend opens websocket channels even though this lab does not publish
// live Network events yet. Accepting upgrades avoids false "app unreachable"
// states while keeping event simulation isolated in websocket.cjs.
server.on("upgrade", acceptWebSocket);

// Listen after all handlers are attached so the startup log means the facade is
// ready for both HTTP and websocket bootstrap calls.
server.listen(port, host, () => {
  log(`listening http://${host}:${port}`);
});
