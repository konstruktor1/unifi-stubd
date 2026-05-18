const { host, port, siteName } = require("./config.cjs");
const { log } = require("./logger.cjs");
const {
  appManifest,
  appStatus,
  controllerInfo,
  currentUser,
  gatewayDevice,
  networkConfig,
  okPayload,
  settings,
  site,
} = require("./payloads.cjs");

// Route only the Core-facing Network endpoints that the lab setup flow has been
// observed to call. Unknown routes are logged and answered with an empty ok
// envelope so the UI can continue while the missing surface is documented.
function route(req, body) {
  const url = new URL(req.url || "/", `http://${host}:${port}`);
  const path = url.pathname;

  // UniFi Core uses these two endpoints to decide whether the Network app is
  // installed, running, and compatible with the packaged frontend assets.
  if (req.method === "GET" && path === "/api/ucore/status") {
    return { status: 200, payload: appStatus() };
  }

  if (req.method === "GET" && path === "/api/ucore/manifest") {
    return { status: 200, payload: appManifest() };
  }

  if (req.method === "GET" && path === "/v2/api/site/default/features") {
    return { status: 200, payload: [] };
  }

  // The VPN wizard asks for suggested ports during setup. Returning common
  // defaults is enough to satisfy the validation path without opening ports.
  if (req.method === "GET" && path === "/v2/api/site/default/network/port-suggest") {
    const service = url.searchParams.get("service");
    return {
      status: 200,
      payload: {
        port: service === "openvpn" ? 1194 : 51820,
      },
    };
  }

  if (req.method === "GET" && path === "/v2/api/site/default/described-features") {
    return { status: 200, payload: [] };
  }

  if (
    req.method === "GET" &&
    path.startsWith("/v2/api/features/") &&
    path.endsWith("/exists")
  ) {
    return { status: 200, payload: { feature_exists: false } };
  }

  if (
    req.method === "GET" &&
    path.startsWith("/v2/api/site/default/features/") &&
    path.endsWith("/exists")
  ) {
    return { status: 200, payload: { feature_exists: false } };
  }

  if (req.method === "GET" && path === "/v2/api/info") {
    return { status: 200, payload: controllerInfo() };
  }

  // The frontend updates local user preferences through /api/self. Persisting
  // those values is outside the current lab scope, so both read and write return
  // the same deterministic owner account.
  if (req.method === "GET" && path === "/api/self") {
    return { status: 200, payload: okPayload([currentUser()]) };
  }

  if (req.method === "PUT" && path === "/api/self") {
    return { status: 200, payload: okPayload([currentUser()]) };
  }

  if (req.method === "GET" && path === "/v2/api/site/default/smart-subnet") {
    return {
      status: 200,
      payload: { code: "api.err.PreviousSubnetWasNotDetected" },
    };
  }

  // Device, client, health, and config endpoints provide the local gateway view
  // used after setup. They describe the simulated console only; client and WLAN
  // lists stay empty until the lab models those surfaces deliberately.
  if (req.method === "GET" && path === "/v2/api/site/default/apgroups") {
    return {
      status: 200,
      payload: [
        {
          _id: "default",
          attr_hidden_id: "default",
          attr_no_delete: true,
          device_macs: [],
          name: "All APs",
        },
      ],
    };
  }

  if (req.method === "GET" && path === "/api/s/default/stat/device") {
    return { status: 200, payload: okPayload([gatewayDevice()]) };
  }

  if (req.method === "GET" && path === "/api/s/default/stat/sta") {
    return { status: 200, payload: okPayload() };
  }

  if (req.method === "GET" && path === "/api/s/default/stat/health") {
    return {
      status: 200,
      payload: okPayload([
        {
          subsystem: "www",
          status: "ok",
        },
        {
          subsystem: "wan",
          status: "ok",
        },
        {
          subsystem: "lan",
          status: "ok",
        },
      ]),
    };
  }

  if (req.method === "GET" && path === "/api/s/default/rest/networkconf") {
    return { status: 200, payload: okPayload(networkConfig()) };
  }

  if (req.method === "GET" && path === "/api/s/default/rest/wlanconf") {
    return { status: 200, payload: okPayload() };
  }

  if (req.method === "GET" && path === "/api/s/default/get/setting") {
    return { status: 200, payload: okPayload(settings()) };
  }

  if (req.method === "GET" && path === "/api/s/default/self/sites") {
    return { status: 200, payload: okPayload([site()]) };
  }

  if (req.method === "GET" && path === "/api/s/default/rest/portconf") {
    return {
      status: 200,
      payload: okPayload([
        {
          _id: "default",
          name: "All",
          site_id: siteName,
        },
      ]),
    };
  }

  if (
    req.method === "GET" &&
    path === "/v2/api/site/default/system-log/remote-settings"
  ) {
    return {
      status: 200,
      payload: {
        contents: [],
        enabled: false,
        log_all_contents: false,
        this_controller: false,
      },
    };
  }

  if (req.method === "GET" && path === "/v2/api/site/default/settings/mgmt") {
    return {
      status: 200,
      payload: {
        direct_connect_supported: false,
        direct_connect_enabled: false,
      },
    };
  }

  if (req.method === "GET" && path === "/v2/api/site/default/shadowmode/status") {
    return {
      status: 200,
      payload: {
        shadow_mode_config_created: false,
        state: "UNKNOWN",
      },
    };
  }

  if (
    req.method === "POST" &&
    [
      "/api/set/setting/network_optimization",
      "/api/set/setting/country",
      "/api/set/setting/provider_capabilities",
      "/api/cmd/system",
      "/api/cmd/devmgr/setup-wan",
      "/api/cmd/devmgr/setup-firewall",
      "/api/rest/wlanconf",
      "/v2/api/site/default/smart-subnet",
      "/v2/api/site/default/dhcp/resolve-wan-subnet-conflict",
      "/v2/api/site/default/shadowmode/rollback",
      "/v2/api/site/default/settings/mgmt/direct_connect/enable",
      "/v2/api/site/default/settings/mgmt/direct_connect/disable",
    ].includes(path)
  ) {
    // Setup writes are acknowledged but not applied to the host/container. That
    // keeps the safety boundary explicit while allowing the webportal flow to
    // move past controller-side provisioning steps.
    return { status: 200, payload: okPayload() };
  }

  if (
    req.method === "DELETE" &&
    path === "/v2/api/magicsitetositevpn/configs"
  ) {
    return { status: 200, payload: okPayload() };
  }

  if (req.method === "GET" && path === "/v2/api/uisp/status") {
    return { status: 200, payload: { enabled: false, connected: false } };
  }

  // Keep unknown calls visible in the request log. Returning ok is intentional:
  // at this research stage a missing optional Network endpoint should be easy
  // to discover without breaking the whole UniFi OS setup surface.
  log(`unknown ${req.method} ${path} body=${body}`);
  return { status: 200, payload: okPayload() };
}

module.exports = {
  route,
};
