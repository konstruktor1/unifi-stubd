const pathModule = require("node:path");

// Keep all tunables in one place so Docker, UTM-derived experiments, and local
// one-shot tests can point the facade at different ports or UI assets without
// editing the route and payload code.
const host = "127.0.0.1";

// The facade listens only on loopback because nginx/UniFi Core proxy to it from
// inside the firmware container. It is not intended as a host-exposed API.
const port = Number(process.env.UNIFI_FW_SIM_NETWORK_STUB_PORT || 8081);
const logPath =
  process.env.UNIFI_FW_SIM_NETWORK_STUB_LOG ||
  "/tmp/udm-pro-se-webportal/network-app-stub.log";

// UniFi Portal marks Network as "activating" when the application version does
// not match the SWAI UI version. Keep this as the semver base that pairs with
// the bundled UI manifest's 10.3.58.0 value.
const version = process.env.UNIFI_FW_SIM_NETWORK_VERSION || "10.3.58";
const packageVersion =
  process.env.UNIFI_FW_SIM_NETWORK_PACKAGE_VERSION || "10.3.58-34147-1";

const siteName = process.env.UNIFI_FW_SIM_NETWORK_SITE || "default";

// Keep the gateway identity stable across Docker rebuilds. The VM path can
// override this when comparing facade payloads with QEMU/UTM observations.
const gatewayMac =
  process.env.UNIFI_FW_SIM_NETWORK_GATEWAY_MAC || "02:15:6d:00:ea:2c";
const gatewayId = "lab-udm-pro-se";

// UI assets come from the extracted firmware rootfs. The facade only publishes
// their manifest location; it does not serve or modify the frontend bundle.
const uiDir =
  process.env.UNIFI_FW_SIM_NETWORK_UI_DIR ||
  "/usr/lib/unifi/webapps/ROOT/app-unifi";
const uiManifestPath = pathModule.join(uiDir, "manifest.json");

module.exports = {
  gatewayId,
  gatewayMac,
  host,
  logPath,
  packageVersion,
  port,
  siteName,
  uiDir,
  uiManifestPath,
  version,
};
