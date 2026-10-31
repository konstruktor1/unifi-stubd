import { Buffer } from "node:buffer";
import { constants } from "node:fs";
import { open } from "node:fs/promises";
import { setTimeout as sleep } from "node:timers/promises";
import { TextDecoder } from "node:util";
import { chromium } from "playwright-core";
import controllerModes from "../controller-adoption/controller_modes.json" with { type: "json" };
import {
  fetchPublicFingerprint,
  publicFingerprintURL,
} from "./public-data.mjs";

const controllerMode = process.env.UNIFI_STUB_CONTROLLER_MODE ?? "classic";
if (!Object.hasOwn(controllerModes, controllerMode))
  throw new Error("controller mode must be classic or unifi-os");
const configuredURL = process.env.UNIFI_STUB_CONTROLLER_URL;
const authFile = process.env.UNIFI_STUB_CONTROLLER_AUTH_FILE;
if (controllerMode === "unifi-os" && (!configuredURL || !authFile))
  throw new Error("unifi-os requires a controller URL and private auth file");
if (controllerMode === "classic" && authFile)
  throw new Error("auth file is only supported for unifi-os");
const baseURL = controllerOrigin(
  configuredURL ?? "https://unifi-controller:8443",
);
const routes = controllerModes[controllerMode];
const loginURL = controllerPath(routes.login_path, baseURL);
const uiURL = controllerPath(routes.ui_path, baseURL);
const deviceIP = process.env.UNIFI_STUB_UI_DEVICE_IP;
const speedMbps = Number(process.env.UNIFI_STUB_UI_SPEED_MBPS);
if (!deviceIP || !Number.isInteger(speedMbps) || speedMbps <= 0) {
  throw new Error("expected a device IP and positive observed speed");
}

function controllerOrigin(value) {
  try {
    const url = new URL(value);
    if (
      /[\s\\?#]/.test(value) ||
      url.protocol !== "https:" ||
      url.username ||
      url.password ||
      url.pathname !== "/" ||
      url.search ||
      url.hash ||
      (url.port && Number(url.port) < 1)
    )
      throw new Error();
    return url.origin;
  } catch {
    throw new Error(
      "controller URL must be an https origin without credentials or path",
    );
  }
}

function controllerPath(value, origin) {
  if (
    typeof value !== "string" ||
    !value.startsWith("/") ||
    value.startsWith("//")
  )
    throw new Error("invalid controller mode paths");
  const url = new URL(value, origin);
  if (url.origin !== origin || url.search || url.hash)
    throw new Error("invalid controller mode paths");
  return url;
}

async function readCredentials(path) {
  let file;
  try {
    file = await open(
      path,
      constants.O_RDONLY | constants.O_NOFOLLOW | constants.O_NONBLOCK,
    );
    const info = await file.stat();
    if (
      !info.isFile() ||
      (info.mode & 0o7777) !== 0o600 ||
      info.size < 1 ||
      info.size > 65536
    )
      throw new Error();
    const data = Buffer.alloc(65537);
    let length = 0;
    while (length < data.length) {
      const { bytesRead } = await file.read(
        data,
        length,
        data.length - length,
        null,
      );
      if (bytesRead === 0) break;
      length += bytesRead;
    }
    if (length < 1 || length > 65536) throw new Error();
    const auth = JSON.parse(
      new TextDecoder("utf-8", { fatal: true }).decode(
        data.subarray(0, length),
      ),
    );
    if (
      !auth ||
      typeof auth !== "object" ||
      Array.isArray(auth) ||
      Object.keys(auth).length !== 2 ||
      typeof auth.username !== "string" ||
      auth.username.length === 0 ||
      typeof auth.password !== "string" ||
      auth.password.length === 0
    )
      throw new Error();
    return auth;
  } catch {
    throw new Error(
      "auth file must be regular, private 0600 and contain only username/password JSON within 64 KiB",
    );
  } finally {
    await file?.close();
  }
}

function loginRetryDelay(value) {
  const seconds = Number(value);
  if (value?.trim() && Number.isFinite(seconds) && seconds >= 0)
    return seconds * 1000;
  const date = Date.parse(value ?? "");
  return Number.isNaN(date) ? 60_000 : Math.max(0, date - Date.now());
}

async function loginController(context, auth) {
  let remainingBackoff = 300_000;
  for (let attempt = 0; attempt < 6; attempt += 1) {
    let response;
    try {
      response = await context.request.post(loginURL.href, {
        data: { ...auth, remember: false },
        maxRedirects: 0,
      });
    } catch {
      throw new Error("controller login request failed");
    }
    if (response.status() !== 429) return response;
    const delay = loginRetryDelay(response.headers()["retry-after"]);
    await response.dispose();
    if (attempt === 5 || delay > remainingBackoff) break;
    remainingBackoff -= delay;
    await sleep(delay);
  }
  throw new Error("controller login returned 429 after bounded retries");
}

function assertUIURL(value) {
  const url = new URL(value);
  const path =
    controllerMode === "unifi-os"
      ? new URL(".", uiURL).pathname
      : uiURL.pathname;
  const prefix = path.endsWith("/") ? path : `${path}/`;
  if (
    url.origin !== baseURL ||
    (url.pathname !== path && !url.pathname.startsWith(prefix))
  )
    throw new Error(
      "controller UI left the expected origin or application path",
    );
}

function diagnosticText(value, secrets, limit = 8192) {
  let text = value.replace(/\b(?:https?|wss?):\/\/[^\s"'<>]+/gi, (value) => {
    try {
      return new URL(value).pathname;
    } catch {
      return "<url>";
    }
  });
  text = text.replace(/(\/[^\s"'<>?#]*)[?#][^\s"'<>]*/g, "$1");
  for (const secret of secrets)
    if (secret) text = text.replaceAll(secret, "<redacted>");
  return new TextDecoder().decode(Buffer.from(text).subarray(0, limit), {
    stream: true,
  });
}

async function writeUIDiagnostics(page, secrets, events) {
  try {
    const url = new URL(page.url());
    const diagnostics = {
      title: diagnosticText(await page.title(), secrets, 1024),
      pathname:
        url.origin === baseURL
          ? diagnosticText(url.pathname, secrets, 1024)
          : "<outside-origin>",
      frames: [],
      events: events.map((event) => ({
        ...event,
        pathname: diagnosticText(event.pathname, secrets, 1024),
        code:
          typeof event.code === "string"
            ? diagnosticText(event.code, secrets, 128)
            : event.code,
        ...(event.message && {
          message: diagnosticText(event.message, secrets, 1024),
        }),
      })),
    };
    for (const frame of page.frames()) {
      if (diagnostics.frames.length === 4) break;
      const frameURL = new URL(frame.url());
      if (frameURL.origin !== baseURL) continue;
      try {
        const state = await frame.locator("body").evaluate(
          (body, origin) => {
            const location = body.ownerDocument.location;
            return location.origin === origin
              ? { pathname: location.pathname, bodyText: body.innerText }
              : null;
          },
          baseURL,
          { timeout: 2000 },
        );
        if (state)
          diagnostics.frames.push({
            pathname: diagnosticText(state.pathname, secrets, 1024),
            bodyText: diagnosticText(state.bodyText, secrets),
          });
      } catch {
        diagnostics.frames.push({
          pathname: diagnosticText(frameURL.pathname, secrets, 1024),
          bodyText: "<unavailable>",
        });
      }
    }
    process.stderr.write(
      `controller UI diagnostics: ${JSON.stringify(diagnostics)}\n`,
    );
  } catch {
    // A diagnostic failure must preserve the original UI failure.
  }
}

function fieldValue(text, label) {
  const lines = text.split(/\r?\n/).map((line) => line.trim());
  const index = lines.indexOf(label);
  return index < 0 ? undefined : lines.slice(index + 1).find(Boolean);
}

function positivePackets(value) {
  return Number(value?.match(/^([0-9]+)/)?.[1]) > 0 ? value : undefined;
}

async function setPortColumn(columns, label, selected) {
  const checkbox = columns.getByRole("checkbox", { name: label, exact: true });
  if ((await checkbox.isChecked({ timeout: 5_000 })) === selected) return;
  await checkbox
    .locator("xpath=ancestor::label[@data-uic-component='Checkbox.Root']")
    .click({ timeout: 5_000 });
  const deadline = Date.now() + 5_000;
  while (Date.now() <= deadline) {
    if ((await checkbox.isChecked({ timeout: 1_000 })) === selected) return;
    await sleep(100);
  }
  throw new Error(
    `Port Manager column ${label} did not reach selected=${selected}`,
  );
}

async function openPortColumns(page, label) {
  const columns = page.locator(
    '[data-waypoint-id="network.ports.filters.columns"]',
  );
  const deadline = Date.now() + 15_000;
  let lastClickError;
  while (Date.now() <= deadline) {
    if (await columns.isVisible()) return columns;
    const footer = page.getByTestId("left-filter-panel-footer");
    // Overflow hides the original content actions with opacity:0 and renders
    // a separate footer. Playwright still considers that hidden copy visible.
    const panel = (await footer.isVisible())
      ? footer
      : page.getByTestId("left-filter-panel-content");
    const button = panel.getByRole("button", { name: label, exact: true });
    try {
      if (
        (await button.isVisible()) &&
        (await button.isEnabled({ timeout: 1_000 }))
      )
        await button.click({ timeout: 1_000 });
    } catch (error) {
      if (error.name !== "TimeoutError") throw error;
      lastClickError = error;
    }
    await sleep(100);
  }
  throw new Error("Port Manager column dialog did not open within 15 seconds", {
    cause: lastClickError,
  });
}

async function openPortMetrics(page) {
  const controls = routes.ui_port_metrics;
  await page
    .getByRole("button", { name: controls.open_label, exact: true })
    .click();
  assertUIURL(page.url());
  const table = page.locator('[data-waypoint-id="network.ports.table"]');
  await table.waitFor({ state: "visible", timeout: 90_000 });
  const columns = await openPortColumns(page, controls.columns_label);
  // A narrow native column selection keeps the required cells rendered despite
  // the Port Manager's horizontal column virtualization.
  await setPortColumn(columns, controls.all_label, true);
  await setPortColumn(columns, controls.all_label, false);
  for (const label of Object.values(controls.column_labels))
    await setPortColumn(columns, label, true);
  await columns
    .getByRole("button", { name: controls.done_label, exact: true })
    .click();
  return table;
}

async function readPortMetrics(table, expectedSpeed) {
  const speedCells = table
    .locator('td[data-column-id="speed"]')
    .filter({ hasText: expectedSpeed });
  const count = await speedCells.count();
  if (count > 1)
    throw new Error("Port Manager has ambiguous observed-speed rows");
  if (count === 0) return undefined;
  const row = speedCells.locator("..");
  const values = {};
  for (const column of Object.keys(routes.ui_port_metrics.column_labels)) {
    const cell = row.locator(`td[data-column-id="${column}"]`);
    await cell.scrollIntoViewIfNeeded({ timeout: 2_000 });
    if (!(await cell.isVisible())) return undefined;
    values[column] = (await cell.innerText()).trim();
  }
  for (const column of Object.keys(values).filter(
    (column) => column !== "speed",
  ))
    if (
      !/^[0-9]+(?:[.,\s][0-9]+)*(?:\s*[KMGTPE]?i?B)?$/i.test(values[column]) ||
      !/[1-9]/.test(values[column])
    )
      return undefined;
  return {
    speed: values.speed.match(expectedSpeed)?.[0],
    rxPackets: values.rx_packets,
    rxBytes: values.rx_sum,
    txPackets: values.tx_packets,
    txBytes: values.tx_sum,
  };
}

const expectedSpeed =
  speedMbps >= 1000 && speedMbps % 1000 === 0
    ? new RegExp(
        `\\b${speedMbps / 1000}(?:[.,]0)?\\s*(?:Gbps|Gbit/s|GbE)\\b`,
        "i",
      )
    : new RegExp(`\\b${speedMbps}\\s*(?:Mbps|Mbit/s|MbE)\\b`, "i");

const browser = await chromium.launch({
  args: ["--disable-dev-shm-usage", "--no-sandbox"],
});
try {
  const context = await browser.newContext({
    baseURL,
    ignoreHTTPSErrors: true,
    serviceWorkers: "block",
    viewport: { width: 1440, height: 1000 },
  });
  await context.grantPermissions(["local-network-access"], { origin: baseURL });
  let networkFailure;
  let fallbackCSRF;
  const diagnosticSecrets = [];
  const diagnosticEvents = [];
  const diagnosticHosts = new Set([
    new URL(baseURL).host,
    new URL(baseURL).hostname,
  ]);
  const recordDiagnostic = (kind, url, code, message) => {
    diagnosticHosts.add(url.host);
    diagnosticHosts.add(url.hostname);
    diagnosticEvents.push({
      kind,
      pathname: url.pathname,
      code,
      originAllowed:
        url.origin === baseURL ||
        url.href === publicFingerprintURL ||
        (url.protocol === "wss:" &&
          url.host === new URL(baseURL).host &&
          !url.username &&
          !url.password),
      ...(message && { message }),
    });
    if (diagnosticEvents.length > 20) diagnosticEvents.shift();
  };
  await context.route("**/*", async (route) => {
    const url = new URL(route.request().url());
    if (
      url.protocol === "data:" ||
      (url.protocol === "blob:" && url.origin === baseURL)
    )
      return route.continue();
    if (
      url.href === publicFingerprintURL &&
      route.request().method() === "GET"
    ) {
      try {
        const body = await fetchPublicFingerprint();
        await route.fulfill({
          status: 200,
          contentType: "application/json",
          headers: { "access-control-allow-origin": baseURL },
          body,
        });
      } catch {
        recordDiagnostic("public-data-error", url, "fingerprint-rejected");
        await route.abort();
      }
      return;
    }
    if (url.origin !== baseURL) {
      recordDiagnostic("blocked-request", url, "origin-policy");
      return route.abort();
    }
    let response;
    try {
      // Browser routing does not safely constrain an automatically followed
      // redirect chain. No redirect response is released to the browser.
      const headers = await route.request().allHeaders();
      if (!Object.hasOwn(headers, "x-csrf-token") && fallbackCSRF)
        headers["x-csrf-token"] = fallbackCSRF;
      const requestCSRF = headers["x-csrf-token"];
      if (requestCSRF && !diagnosticSecrets.includes(requestCSRF))
        diagnosticSecrets.push(requestCSRF);
      response = await route.fetch({ maxRedirects: 0, headers });
      const responseHeaders = response.headers();
      const responseCSRF =
        responseHeaders["x-updated-csrf-token"]?.trim() ||
        responseHeaders["x-csrf-token"]?.trim();
      if (responseCSRF) {
        fallbackCSRF = responseCSRF;
        if (!diagnosticSecrets.includes(responseCSRF))
          diagnosticSecrets.push(responseCSRF);
      }
      if (response.status() >= 400)
        recordDiagnostic("http-error", url, response.status());
      if (
        response.status() >= 300 &&
        response.status() < 400 &&
        response.status() !== 304
      ) {
        recordDiagnostic("http-error", url, response.status());
        networkFailure =
          "controller returned an HTTP redirect during UI verification";
        return await route.abort();
      }
      await route.fulfill({ response });
    } catch {
      recordDiagnostic("request-failed", url, "fetch-failed");
      networkFailure = "controller UI network request failed";
      await route.abort();
    } finally {
      await response?.dispose();
    }
  });
  await context.routeWebSocket("**/*", async (route) => {
    const url = new URL(route.url());
    if (
      url.protocol !== "wss:" ||
      url.host !== new URL(baseURL).host ||
      url.username ||
      url.password
    ) {
      recordDiagnostic("blocked-request", url, "websocket-policy");
      return route.close();
    }
    route.connectToServer();
  });

  const auth =
    controllerMode === "unifi-os"
      ? await readCredentials(authFile)
      : { username: "admin", password: "admin" };
  diagnosticSecrets.push(auth.username, auth.password);
  // APIRequestContext bypasses browser routes; bind the credential destination
  // separately and prevent every automatic redirect before sending it.
  if (loginURL.origin !== baseURL)
    throw new Error("controller login changed origin");
  const login = await loginController(context, auth);
  try {
    if (login.status() !== 200)
      throw new Error(`controller login returned ${login.status()}`);
    if (controllerMode === "unifi-os") {
      const csrf = login.headers()["x-csrf-token"]?.trim();
      const cookies = await context.cookies(baseURL);
      if (
        !csrf ||
        !cookies.some(
          (cookie) => cookie.name === "TOKEN" && cookie.value.trim(),
        )
      )
        throw new Error(
          "unifi-os login did not establish a TOKEN and CSRF session",
        );
      diagnosticSecrets.push(csrf, ...cookies.map((cookie) => cookie.value));
      fallbackCSRF = csrf;
    }
  } finally {
    await login.dispose();
  }

  const page = await context.newPage();
  page.on("websocket", (socket) => {
    const url = new URL(socket.url());
    recordDiagnostic("websocket-created", url, "created");
    socket.on("socketerror", (message) => {
      recordDiagnostic(
        "websocket-error",
        url,
        "socket-error",
        diagnosticText(
          message,
          [...diagnosticSecrets, ...diagnosticHosts],
          1024,
        ),
      );
    });
    socket.on("close", () =>
      recordDiagnostic("websocket-close", url, "closed"),
    );
  });
  let socketDiagnostics;
  const socketURLs = new Map();
  try {
    socketDiagnostics = await context.newCDPSession(page);
    socketDiagnostics.on("Network.webSocketCreated", ({ requestId, url }) => {
      socketURLs.set(requestId, new URL(url));
      if (socketURLs.size > 20)
        socketURLs.delete(socketURLs.keys().next().value);
    });
    socketDiagnostics.on(
      "Network.webSocketHandshakeResponseReceived",
      ({ requestId, response }) => {
        const url = socketURLs.get(requestId);
        if (url) recordDiagnostic("websocket-handshake", url, response.status);
      },
    );
    await socketDiagnostics.send("Network.enable");
  } catch {
    recordDiagnostic(
      "diagnostic-error",
      new URL(page.url()),
      "cdp-unavailable",
    );
  }
  page.on("pageerror", (error) => {
    try {
      recordDiagnostic(
        "page-error",
        new URL(page.url()),
        error.name,
        diagnosticText(
          error.message,
          [...diagnosticSecrets, ...diagnosticHosts],
          1024,
        ),
      );
    } catch {
      // Diagnostic collection must not affect UI verification.
    }
  });
  try {
    try {
      await page.goto(uiURL.href, {
        waitUntil: "domcontentloaded",
        timeout: 30_000,
      });
    } catch {
      throw new Error(networkFailure ?? "controller UI navigation failed");
    }
    assertUIURL(page.url());
    const deviceIdentity = page.getByText(deviceIP, { exact: true }).first();
    await deviceIdentity.waitFor({ state: "visible", timeout: 90_000 });

    const reminder = page
      .getByRole("button", { name: "Remind Me Later", exact: true })
      .filter({ visible: true });
    if (await reminder.count()) await reminder.first().click();
    await deviceIdentity.click({ timeout: 90_000 });
    const portTable =
      controllerMode === "unifi-os" ? await openPortMetrics(page) : undefined;

    const deadline = Date.now() + 60_000;
    let metrics;
    while (Date.now() <= deadline) {
      assertUIURL(page.url());
      if (networkFailure) throw new Error(networkFailure);
      if (portTable) {
        metrics = await readPortMetrics(portTable, expectedSpeed);
      } else {
        const text = await page.locator("body").innerText();
        const speed = fieldValue(text, "Experience")?.match(expectedSpeed)?.[0];
        const down = positivePackets(fieldValue(text, "Down Pkts / Bytes"));
        const up = positivePackets(fieldValue(text, "Up Pkts / Bytes"));
        if (speed && down && up) metrics = { speed, down, up };
      }
      if (metrics) break;
      await page.waitForTimeout(2_000);
    }
    if (!metrics)
      throw new Error("device details did not show speed and traffic");
    const traffic = portTable
      ? `rx_packets=${metrics.rxPackets} rx_bytes=${metrics.rxBytes} tx_packets=${metrics.txPackets} tx_bytes=${metrics.txBytes}`
      : `down=${metrics.down} up=${metrics.up}`;
    process.stderr.write(
      `controller UI: device_ip=${deviceIP} speed=${metrics.speed} ${traffic}\n`,
    );
  } catch (error) {
    await writeUIDiagnostics(
      page,
      [...diagnosticSecrets, ...diagnosticHosts],
      diagnosticEvents,
    );
    throw error;
  } finally {
    try {
      await socketDiagnostics?.detach();
    } catch {
      // Observer cleanup must preserve the UI result.
    }
  }
} finally {
  await browser.close();
}
