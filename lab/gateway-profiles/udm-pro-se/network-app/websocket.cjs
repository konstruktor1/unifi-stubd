const crypto = require("node:crypto");

const { log } = require("./logger.cjs");

// The UI opens websocket connections while bootstrapping, but this lab profile
// does not stream live device telemetry yet. Completing the handshake is enough
// to keep the frontend from treating Network as unreachable.
function acceptWebSocket(req, socket) {
  const key = req.headers["sec-websocket-key"];
  if (!key) {
    // A missing key means the request is not a valid websocket upgrade. Fail it
    // plainly instead of routing it through the JSON API handler.
    socket.end("HTTP/1.1 400 Bad Request\r\n\r\n");
    return;
  }

  // RFC 6455 requires this GUID when deriving the Sec-WebSocket-Accept value.
  // Keeping the handshake standards-compliant is enough for the UI bootstrap.
  const accept = crypto
    .createHash("sha1")
    .update(`${key}258EAFA5-E914-47DA-95CA-C5AB0DC85B11`)
    .digest("base64");

  socket.write(
    [
      "HTTP/1.1 101 Switching Protocols",
      "Upgrade: websocket",
      "Connection: Upgrade",
      `Sec-WebSocket-Accept: ${accept}`,
      "\r\n",
    ].join("\r\n"),
  );
  log(`WS ${req.url || "/"} open`);
  socket.on("close", () => log(`WS ${req.url || "/"} close`));
  socket.on("error", (error) => log(`WS ${req.url || "/"} error=${error}`));
}

module.exports = {
  acceptWebSocket,
};
