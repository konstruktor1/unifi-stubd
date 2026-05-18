#!/usr/bin/env node
// Minimal UniFi Network API facade for the UDM Pro SE webportal lab.
//
// The real Network application is not started in this firmware profile. UniFi
// Core still calls a small set of Network endpoints while finishing setup, so
// this stub records those calls and returns deterministic no-op responses.

const http = require("node:http");
const crypto = require("node:crypto");
const fs = require("node:fs");
const pathModule = require("node:path");

const host = "127.0.0.1";
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
const gatewayMac =
  process.env.UNIFI_FW_SIM_NETWORK_GATEWAY_MAC || "02:15:6d:00:ea:2c";
const gatewayId = "lab-udm-pro-se";
const uiDir =
  process.env.UNIFI_FW_SIM_NETWORK_UI_DIR ||
  "/usr/lib/unifi/webapps/ROOT/app-unifi";
const uiManifestPath = pathModule.join(uiDir, "manifest.json");

function log(line) {
  fs.mkdirSync(pathModule.dirname(logPath), { recursive: true });
  fs.appendFileSync(logPath, `${new Date().toISOString()} ${line}\n`);
}

function readUiManifest() {
  try {
    return JSON.parse(fs.readFileSync(uiManifestPath, "utf8"));
  } catch (error) {
    log(`failed to read ${uiManifestPath}: ${error}`);
    return {};
  }
}

function readBody(req) {
  return new Promise((resolve, reject) => {
    let data = "";
    req.setEncoding("utf8");
    req.on("data", (chunk) => {
      data += chunk;
    });
    req.on("end", () => resolve(data));
    req.on("error", reject);
  });
}

function sendJson(res, status, payload) {
  const body = JSON.stringify(payload);
  res.writeHead(status, {
    "content-type": "application/json",
    "content-length": Buffer.byteLength(body),
  });
  res.end(body);
}

function acceptWebSocket(req, socket) {
  const key = req.headers["sec-websocket-key"];
  if (!key) {
    socket.end("HTTP/1.1 400 Bad Request\r\n\r\n");
    return;
  }

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

function okPayload(data = []) {
  return {
    meta: {
      rc: "ok",
    },
    data,
  };
}

function currentUser() {
  return {
    _id: "admin",
    admin_id: "admin",
    avatar_url: "",
    email: "admin@example.invalid",
    email_alert_enabled: false,
    first_name: "admin",
    html_email_enabled: false,
    is_owner: true,
    is_professional_installer: false,
    is_super: true,
    last_name: "",
    last_site_name: siteName,
    name: "admin",
    org_role: "owner",
    push_alert_enabled: false,
    requires_new_password: false,
    role: "admin",
    site_id: siteName,
    super_admin: true,
    ubic_name: "admin",
    ubic_uuid: "admin",
    uid: "admin",
    ui_settings: {
      preferredLanguage: "en",
      isAppDark: true,
      timeFormat: "h:mm a",
      use24HourTime: false,
    },
  };
}

function site() {
  return {
    _id: siteName,
    attr_hidden_id: "default",
    attr_no_delete: true,
    desc: "Default",
    device_count: 1,
    gw_mac: gatewayMac,
    is_active: true,
    name: siteName,
    permissions: ["admin"],
    role: "admin",
    timezone: "Europe/Zurich",
  };
}

function networkInfo() {
  return {
    wifiExperienceScore: 100,
    clientCount: 0,
    wiredClients: 0,
    wirelessClients: 0,
    guestClients: 0,
    isReadyForSetup: true,
    wanStatus: "connected",
    gatewayConfigVersion: "lab",
    portProfiles: [
      {
        id: "default",
        name: "All",
      },
    ],
    portTable: [
      {
        fullDuplex: true,
        ifname: "eth0",
        name: "WAN",
        poeCaps: 0,
        poePower: 0,
        speed: 10000,
        throughputRx: 0,
        throughputTx: 0,
        type: "wan",
        up: true,
        usageRx: 0,
        usageTx: 0,
        profileId: "default",
      },
      {
        fullDuplex: true,
        ifname: "rtl8370-lan1",
        name: "LAN 1",
        poeCaps: 0,
        poePower: 0,
        speed: 1000,
        throughputRx: 0,
        throughputTx: 0,
        type: "lan",
        up: false,
        usageRx: 0,
        usageTx: 0,
        profileId: "default",
      },
    ],
    interfaces: [
      {
        name: "eth0",
        ip: "172.17.0.2",
      },
    ],
  };
}

function gatewayDevice() {
  return {
    _id: gatewayId,
    adopted: true,
    board_rev: 1,
    cfgversion: "lab",
    connected_at: Math.floor(Date.now() / 1000),
    displayable_version: packageVersion,
    ethernet_table: [
      {
        mac: gatewayMac,
        name: "eth0",
        num_port: 1,
      },
    ],
    gateway_mac: gatewayMac,
    has_fan: false,
    inform_ip: "127.0.0.1",
    ip: "172.17.0.2",
    isolated: false,
    kernel_version: "lab",
    known_cfgversion: "lab",
    last_seen: Math.floor(Date.now() / 1000),
    mac: gatewayMac,
    model: "UDMPROSE",
    model_in_eol: false,
    model_in_lts: false,
    name: "UDM Pro SE Firmware Lab",
    network_table: [
      {
        _id: "lan",
        attr_no_delete: true,
        ip: "192.0.2.1",
        name: "Default",
        networkgroup: "LAN",
        purpose: "corporate",
        subnet: "192.0.2.0/24",
      },
    ],
    port_table: networkInfo().portTable.map((port, index) => ({
      ...port,
      port_idx: index,
      portconf_id: "default",
      media: port.type === "wan" ? "SFP+" : "GE",
      speed_caps: port.type === "wan" ? 10000 : 1000,
    })),
    required_version: packageVersion,
    serial: "02156D00EA2C",
    site_id: siteName,
    state: 1,
    type: "udm",
    upgradable: false,
    uptime: 3600,
    version: packageVersion,
    version_incompatible: false,
    wan1: {
      ifname: "eth0",
      ip: "172.17.0.2",
      netmask: "255.255.0.0",
      type: "dhcp",
      up: true,
    },
  };
}

function networkConfig() {
  return [
    {
      _id: "lan",
      attr_hidden_id: "default",
      attr_no_delete: true,
      dhcpd_enabled: false,
      domain_name: "localdomain",
      enabled: true,
      ip_subnet: "192.0.2.1/24",
      name: "Default",
      networkgroup: "LAN",
      purpose: "corporate",
      site_id: siteName,
      vlan_enabled: false,
    },
    {
      _id: "wan",
      attr_no_delete: true,
      enabled: true,
      name: "WAN",
      networkgroup: "WAN",
      purpose: "wan",
      site_id: siteName,
      wan_type: "dhcp",
    },
  ];
}

function settings() {
  return [
    {
      _id: "country",
      key: "country",
      site_id: siteName,
      value: "840",
    },
    {
      _id: "locale",
      key: "locale",
      site_id: siteName,
      timezone: "Europe/Zurich",
    },
    {
      _id: "mgmt",
      direct_connect_enabled: false,
      direct_connect_supported: false,
      key: "mgmt",
      site_id: siteName,
    },
    {
      _id: "network_optimization",
      enabled: false,
      key: "network_optimization",
      site_id: siteName,
    },
    {
      _id: "super_mgmt",
      key: "super_mgmt",
      site_id: siteName,
      x_ssh_enabled: false,
    },
  ];
}

function controllerInfo() {
  return {
    applicationVersion: packageVersion,
    build: "34147",
    buildNumber: 34147,
    controller_uuid: "00000000-0000-0000-0000-02156d00ea2c",
    debug_system: "warn",
    hostname: "udm-pro-se-lab",
    isCloudConsole: true,
    isSetup: true,
    isUniFiOS: true,
    name: "UDM Pro SE Firmware Lab",
    self: currentUser(),
    serverName: "UDM Pro SE Firmware Lab",
    sites: [site()],
    system: {
      cloud_access_enabled: false,
      device_id: gatewayMac,
      gateway_mac: gatewayMac,
      has_gateway: true,
      has_webrtc_support: false,
      host_meta: {
        model: "UDMPROSE",
        shortname: "UDMPROSE",
      },
      hostname: "udm-pro-se-lab",
      inform_port: 8080,
      logging: "warn",
      multiple_sites_supported: false,
      name: "UDM Pro SE Firmware Lab",
      network_server: {
        platform_type: "linux",
      },
      super_permissions: true,
      timezone: "Europe/Zurich",
      unifi_console: {
        fqdn: "127.0.0.1",
        type: "UDMPROSE",
        version: "5.0.25",
      },
      uptime: 3600,
      version: packageVersion,
    },
    timezone: "Europe/Zurich",
    update_available: false,
    version: packageVersion,
  };
}

function appManifest() {
  const uiManifest = readUiManifest();

  return {
    apps: [
      {
        name: "network",
        type: "controller",
        port,
        uiDir,
        uiIndex:
          uiManifest.uiIndex ||
          uiManifest.legacyUiIndex ||
          "hybrid-swai-10.3.58.0-gf6e3b703c.js",
        uiVersion: uiManifest.uiVersion || "10.3.58.0",
        version,
        isConfigured: true,
        isInstalled: true,
        isRunning: true,
        installable: true,
        updatable: true,
        swaiVersion: 2,
        uiCdn: [],
        prefetch: Array.isArray(uiManifest.prefetch) ? uiManifest.prefetch : [],
        flags: ["skipConnectionBlocker"],
        features: {
          devicesLog: true,
          stackable: false,
        },
        info: networkInfo(),
      },
    ],
  };
}

function appStatus() {
  const manifest = appManifest().apps[0];

  return {
    port,
    version,
    versionRaw: version,
    required: true,
    preinstall: false,
    hidden: false,
    installable: true,
    updatable: true,
    isInstalled: true,
    isConfigured: true,
    isRunning: true,
    state: "active",
    status: "ok",
    statusMessage: "",
    installState: "installed",
    controllerStatus: "READY",
    swaiVersion: manifest.swaiVersion,
    uiCdn: manifest.uiCdn,
    uiVersion: manifest.uiVersion,
    uiIndex: manifest.uiIndex,
    prefetch: manifest.prefetch,
    flags: manifest.flags,
    features: manifest.features,
    ui: {
      baseUrl: "/network/",
      publicPath: "/app-assets/network/",
      cdnPublicPaths: [],
      prefetch: manifest.prefetch,
      entrypoint: manifest.uiIndex,
      swaiVersion: manifest.swaiVersion,
      apiPrefix: "/proxy/network/",
      flags: manifest.flags,
    },
    info: networkInfo(),
  };
}

function route(req, body) {
  const url = new URL(req.url || "/", `http://${host}:${port}`);
  const path = url.pathname;

  if (req.method === "GET" && path === "/api/ucore/status") {
    return { status: 200, payload: appStatus() };
  }

  if (req.method === "GET" && path === "/api/ucore/manifest") {
    return { status: 200, payload: appManifest() };
  }

  if (req.method === "GET" && path === "/v2/api/site/default/features") {
    return { status: 200, payload: [] };
  }

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

  log(`unknown ${req.method} ${path} body=${body}`);
  return { status: 200, payload: okPayload() };
}

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

server.on("upgrade", acceptWebSocket);

server.listen(port, host, () => {
  log(`listening http://${host}:${port}`);
});
