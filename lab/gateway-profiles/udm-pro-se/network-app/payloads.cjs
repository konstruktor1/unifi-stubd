const fs = require("node:fs");

const {
  gatewayId,
  gatewayMac,
  packageVersion,
  port,
  siteName,
  uiDir,
  uiManifestPath,
  version,
} = require("./config.cjs");
const { log } = require("./logger.cjs");

function readUiManifest() {
  try {
    // The bundled Network UI manifest provides the real SWAI entrypoint name.
    // Falling back below keeps the lab usable when only a partial rootfs is
    // available for local inspection.
    return JSON.parse(fs.readFileSync(uiManifestPath, "utf8"));
  } catch (error) {
    log(`failed to read ${uiManifestPath}: ${error}`);
    return {};
  }
}

// UniFi's legacy Network API wraps most route payloads in a meta/data envelope.
// Keeping that shape here avoids teaching the route table about response
// formatting and keeps endpoint fixtures close to the observed controller data.
function okPayload(data = []) {
  return {
    meta: {
      rc: "ok",
    },
    data,
  };
}

function currentUser() {
  // The setup UI mostly needs an owner-like user object so it can persist local
  // preferences and route into the default site without cloud identity calls.
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
  // Keep one deterministic default site. Multi-site behavior is not part of the
  // UDM Pro SE setup path and would add controller semantics this lab does not
  // currently model.
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
  // This is the compact readiness summary exposed through UniFi Core's app
  // manifest/status calls. It tells the webportal that Network is installed,
  // has a gateway, and has one connected WAN without deriving host state.
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

// The setup UI expects the console to appear as an already adopted gateway in
// its own local Network application. These values stay deterministic so changes
// in Docker or QEMU networking do not leak host-specific state into fixtures.
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
    port_table: networkInfo().portTable.map((portRow, index) => ({
      ...portRow,
      port_idx: index,
      portconf_id: "default",
      media: portRow.type === "wan" ? "SFP+" : "GE",
      speed_caps: portRow.type === "wan" ? 10000 : 1000,
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
  // The UI asks for Network's LAN/WAN configuration even during offline setup.
  // Return documentation-safe addresses and keep DHCP disabled so the facade
  // never implies it can provision the host or the container network.
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
  // Settings here are only the keys the observed setup flow reads or writes.
  // Security-sensitive toggles such as SSH and direct-connect stay disabled.
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
  // /v2/api/info is a broad bootstrap payload. Keep it internally consistent
  // with appStatus/appManifest so the frontend does not think Network and UniFi
  // OS disagree about the console identity.
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

  // UniFi Core discovers the Network application through this app manifest.
  // The version and UI values deliberately match the firmware bundle so Core
  // reports READY instead of an update or activation mismatch.
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

  // The setup page reads this status repeatedly while waiting for Network. All
  // readiness fields are explicit so a future regression is visible in one
  // payload instead of being hidden behind default frontend behavior.
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

module.exports = {
  appManifest,
  appStatus,
  controllerInfo,
  currentUser,
  gatewayDevice,
  networkConfig,
  networkInfo,
  okPayload,
  settings,
  site,
};
