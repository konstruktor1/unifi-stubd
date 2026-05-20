const fs = require("node:fs");
const pathModule = require("node:path");

const { logPath } = require("./config.cjs");

// Append-only request logs are easier to inspect from support archives than
// stdout alone because several firmware processes share the same container.
function log(line) {
  // Create the directory lazily because tmpfs-backed log paths are rebuilt on
  // every container start.
  fs.mkdirSync(pathModule.dirname(logPath), { recursive: true });
  fs.appendFileSync(logPath, `${new Date().toISOString()} ${line}\n`);
}

module.exports = {
  log,
};
