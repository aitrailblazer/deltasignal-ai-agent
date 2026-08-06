import { chromium } from "playwright";
import { spawn } from "node:child_process";
import process from "node:process";

const root = new URL("../../", import.meta.url).pathname;
const port = 8878;
const server = spawn("python3", ["-m", "http.server", String(port)], {
  cwd: root,
  stdio: "ignore",
});

const delay = (milliseconds) =>
  new Promise((resolve) => setTimeout(resolve, milliseconds));

let browser;
try {
  await delay(500);
  browser = await chromium.launch({ channel: "chrome", headless: true });
  const page = await browser.newPage({ viewport: { width: 1440, height: 1000 } });
  const url = `http://127.0.0.1:${port}/client-demo/google-cloud-aapl/Google_Cloud_AAPL_DeltaSignal_MCP_Client_Demo_2026_08_06.html`;

  await page.goto(url, { waitUntil: "networkidle" });
  await page.locator("#interactive-walkthrough").scrollIntoViewIfNeeded();

  const tabs = page.locator("[data-step-target]");
  if ((await tabs.count()) !== 6) {
    throw new Error("Expected six walkthrough stages.");
  }

  await page.locator("#walk-next").click();
  await page.locator("#walk-next").click();
  if ((await page.locator("#stage-count").textContent())?.trim() !== "3 / 6") {
    throw new Error("Next-stage navigation did not reach stage 3.");
  }
  if (!(await page.locator('[data-step-target="3"]').getAttribute("class"))?.includes("active")) {
    throw new Error("Stage 3 tab was not activated.");
  }

  await page.locator('[data-step-target="6"]').click();
  if (!(await page.locator('[data-activate="6"]').getAttribute("class"))?.includes("is-live")) {
    throw new Error("Client evidence packet did not activate at stage 6.");
  }
  if (!(await page.locator("#walk-next").isDisabled())) {
    throw new Error("Next control must be disabled at the final stage.");
  }

  await page.locator("#walk-reset").click();
  if ((await page.locator("#stage-count").textContent())?.trim() !== "1 / 6") {
    throw new Error("Reset did not return to stage 1.");
  }

  await page.locator("#walk-play").click();
  await page.waitForFunction(
    () => document.querySelector("#stage-count")?.textContent?.trim() === "2 / 6",
    undefined,
    { timeout: 3000 },
  );
  await page.locator("#walk-play").click();
  if (!(await page.locator("#walk-play").textContent())?.includes("Auto play")) {
    throw new Error("Autoplay did not return to its paused control state.");
  }

  await page.locator("#walk-present").click();
  if (!(await page.locator("body").getAttribute("class"))?.includes("presentation-mode")) {
    throw new Error("Presentation view did not activate.");
  }
  await page.keyboard.press("Escape");
  if ((await page.locator("body").getAttribute("class"))?.includes("presentation-mode")) {
    throw new Error("Escape did not exit presentation view.");
  }

  const overflow = await page.evaluate(
    () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
  );
  if (overflow) {
    throw new Error("Desktop walkthrough introduces horizontal overflow.");
  }

  const mobile = await browser.newPage({ viewport: { width: 390, height: 844 } });
  await mobile.goto(`${url}?presentation=1`, { waitUntil: "networkidle" });
  if (!(await mobile.locator("body").getAttribute("class"))?.includes("presentation-mode")) {
    throw new Error("Presentation query option did not activate.");
  }
  const mobileOverflow = await mobile.evaluate(
    () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
  );
  if (mobileOverflow) {
    throw new Error("Mobile walkthrough introduces horizontal overflow.");
  }

  console.log("interactive-walkthrough: PASS");
} finally {
  await browser?.close();
  server.kill("SIGTERM");
}
