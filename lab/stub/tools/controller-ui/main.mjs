import { chromium } from "playwright-core";

const baseURL = "https://unifi-controller:8443";
const deviceIP = process.env.UNIFI_STUB_UI_DEVICE_IP;
const speedMbps = Number(process.env.UNIFI_STUB_UI_SPEED_MBPS);
if (!deviceIP || !Number.isInteger(speedMbps) || speedMbps <= 0) {
  throw new Error("expected a device IP and positive observed speed");
}

function fieldValue(text, label) {
  const lines = text.split(/\r?\n/).map((line) => line.trim());
  const index = lines.indexOf(label);
  return index < 0 ? undefined : lines.slice(index + 1).find(Boolean);
}

function positivePackets(value) {
  return Number(value?.match(/^([0-9]+)/)?.[1]) > 0 ? value : undefined;
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
  await context.route("**/*", (route) => {
    const { hostname, protocol } = new URL(route.request().url());
    return hostname === "unifi-controller" ||
      ["data:", "blob:"].includes(protocol)
      ? route.continue()
      : route.abort();
  });

  const login = await context.request.post("/api/login", {
    data: { username: "admin", password: "admin", remember: false },
  });
  if (!login.ok())
    throw new Error(`controller login returned ${login.status()}`);

  const page = await context.newPage();
  await page.goto("/manage/default/devices", {
    waitUntil: "domcontentloaded",
    timeout: 30_000,
  });
  const deviceRow = page
    .getByRole("row")
    .filter({ has: page.getByText(deviceIP, { exact: true }) });
  await deviceRow.waitFor({ state: "visible", timeout: 90_000 });

  const reminder = page
    .getByRole("button", { name: "Remind Me Later", exact: true })
    .filter({ visible: true });
  if (await reminder.count()) await reminder.first().click();
  await deviceRow.locator("p").first().click({ timeout: 90_000 });

  const deadline = Date.now() + 60_000;
  let metrics;
  while (Date.now() <= deadline) {
    const text = await page.locator("body").innerText();
    const speed = fieldValue(text, "Experience")?.match(expectedSpeed)?.[0];
    const down = positivePackets(fieldValue(text, "Down Pkts / Bytes"));
    const up = positivePackets(fieldValue(text, "Up Pkts / Bytes"));
    if (speed && down && up) {
      metrics = { speed, down, up };
      break;
    }
    await page.waitForTimeout(2_000);
  }
  if (!metrics)
    throw new Error("device details did not show speed and traffic");
  process.stderr.write(
    `controller UI: device_ip=${deviceIP} speed=${metrics.speed} down=${metrics.down} up=${metrics.up}\n`,
  );
} finally {
  await browser.close();
}
