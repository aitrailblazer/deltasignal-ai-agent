import {chromium} from "playwright";
import {spawn} from "node:child_process";
import process from "node:process";

const root = new URL("../", import.meta.url).pathname;
const port = 8880;
const server = spawn("python3", ["-m", "http.server", String(port)], {
  cwd: root,
  stdio: "ignore",
});
const wait = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

let browser;
try {
  await wait(500);
  browser = await chromium.launch({channel: "chrome", headless: true});

  const desktop = await browser.newPage({viewport: {width: 1440, height: 1000}});
  await desktop.goto(`http://127.0.0.1:${port}/`, {waitUntil: "networkidle"});
  await desktop.locator("deltasignal-site").waitFor();

  if ((await desktop.locator("[role=tab]").count()) !== 4) {
    throw new Error("Expected four platform capability tabs.");
  }
  await desktop.locator("[role=tab]").nth(2).click();
  if (!(await desktop.locator(".cap-panel h3").textContent())?.includes("Bounded tools")) {
    throw new Error("Lit capability state did not update.");
  }
  const desktopOverflow = await desktop.evaluate(
    () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
  );
  if (desktopOverflow) throw new Error("Desktop landing page has horizontal overflow.");

  const essentialLinks = [
    "./client-demo/google-cloud-aapl/",
    "./client-demo/google-cloud-aapl/Google_Cloud_AAPL_DeltaSignal_MCP_Client_Demo_2026_08_06.html?presentation=1",
    "./docs/DeltaSignal_Daily_Backend_to_Google_Agents_Architecture_2026_08_06.html",
    "https://github.com/aitrailblazer/deltasignal-ai-agent",
  ];
  const links = await desktop.locator("a").evaluateAll((items) =>
    items.map((item) => item.getAttribute("href")),
  );
  for (const href of essentialLinks) {
    if (!links.includes(href)) throw new Error(`Missing essential link: ${href}`);
  }

  const mobile = await browser.newPage({viewport: {width: 390, height: 844}});
  await mobile.goto(`http://127.0.0.1:${port}/`, {waitUntil: "networkidle"});
  await mobile.locator(".menu").click();
  if (!(await mobile.locator(".navlinks").getAttribute("class"))?.includes("open")) {
    throw new Error("Mobile navigation did not open.");
  }
  const mobileOverflow = await mobile.evaluate(
    () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
  );
  if (mobileOverflow) throw new Error("Mobile landing page has horizontal overflow.");

  console.log("landing-page: PASS");
} finally {
  await browser?.close();
  server.kill("SIGTERM");
  if (process.exitCode) server.kill("SIGKILL");
}
