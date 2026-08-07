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
  if ((await desktop.locator("#agent-architecture .agent-card").count()) !== 4) {
    throw new Error("Expected four bounded specialist agents in the architecture diagram.");
  }
  if (!((await desktop.locator("#agent-architecture").textContent()) ?? "").includes("Evidence-boundary reviewer")) {
    throw new Error("Agent architecture is missing its review gate.");
  }
  if ((await desktop.locator("#agent-architecture .arch-glyph, #agent-architecture .agent-glyph, #agent-architecture .google-glyph").count()) < 9) {
    throw new Error("Agent architecture is missing recognizable stage glyphs.");
  }
  if ((await desktop.locator("#agent-architecture .agent-type").count()) !== 4) {
    throw new Error("Every bounded specialist must be explicitly identified as an AI agent.");
  }
  if ((await desktop.locator("#agent-architecture .google-cloud-mark").count()) !== 1) {
    throw new Error("Cloud Run coordinator is missing its Google Cloud glyph.");
  }
  if ((await desktop.locator("#agent-architecture .gemini-mark").count()) !== 1) {
    throw new Error("Gemini synthesis is missing its Google technology glyph.");
  }
  const architectureText = (await desktop.locator("#agent-architecture").textContent()) ?? "";
  if (!architectureText.includes("GCP RUNTIME · NOT AN AGENT") || !architectureText.includes("GOOGLE AI MODEL · NOT AN AGENT")) {
    throw new Error("Technology and AI-agent roles are not explicitly distinguished.");
  }
  const flowAnimation = await desktop.locator(".arch-arrow").evaluate(
    (element) => getComputedStyle(element, "::before").animationName,
  );
  if (flowAnimation !== "signal-x") {
    throw new Error("Agent architecture directional flow animation is missing.");
  }

  const reducedMotion = await browser.newPage({viewport: {width: 1024, height: 900}});
  await reducedMotion.emulateMedia({reducedMotion: "reduce"});
  await reducedMotion.goto(`http://127.0.0.1:${port}/`, {waitUntil: "networkidle"});
  const reducedAnimation = await reducedMotion.locator(".arch-arrow").evaluate(
    (element) => getComputedStyle(element, "::before").animationName,
  );
  if (reducedAnimation !== "none") {
    throw new Error("Agent architecture does not respect reduced-motion preferences.");
  }

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

  const heroTitle = (await desktop.locator("h1").first().textContent()) ?? "";
  if (!heroTitle.includes("Apple is the proof")) {
    throw new Error("Apple is not the primary landing-page proof.");
  }
  if ((await desktop.getByRole("link", {name: /Run the Apple evidence workflow/}).count()) !== 1) {
    throw new Error("Primary Apple workflow action is missing.");
  }

  const theater = await browser.newPage({viewport: {width: 1920, height: 1080}});
  await theater.goto(`http://127.0.0.1:${port}/?tour=1`, {
    waitUntil: "networkidle",
  });
  if (!(await theater.locator(".competition-theater").isVisible())) {
    throw new Error("Product tour did not activate.");
  }
  if (!((await theater.locator(".competition-copy h1").textContent()) ?? "").includes("Start with Apple")) {
    throw new Error("Product tour does not lead with the Apple reference.");
  }
  if ((await theater.locator(".competition-stage").count()) !== 1) {
    throw new Error("Expected one product presentation stage.");
  }
  await theater.keyboard.press("ArrowRight");
  if (!(await theater.locator(".competition-copy h1").textContent())?.includes("fluent prose")) {
    throw new Error("Product-tour keyboard navigation did not reach slide 2.");
  }
  await theater.locator(".competition-controls button").nth(2).click();
  if (!(await theater.locator(".competition-copy h1").textContent())?.includes("Discover")) {
    throw new Error("Product-tour next control did not reach slide 3.");
  }
  const theaterMetrics = await theater.evaluate(() => ({
    overflowX: document.documentElement.scrollWidth > innerWidth,
    overflowY: document.documentElement.scrollHeight > innerHeight,
    viewport: [innerWidth, innerHeight],
  }));
  if (theaterMetrics.overflowX || theaterMetrics.overflowY) {
    throw new Error(`Product-tour overflow: ${JSON.stringify(theaterMetrics)}`);
  }
  await theater.keyboard.press("Escape");
  if (await theater.locator(".competition-theater").count()) {
    throw new Error("Escape did not exit the product tour.");
  }

  console.log("landing-page: PASS");
} finally {
  await browser?.close();
  server.kill("SIGTERM");
  if (process.exitCode) server.kill("SIGKILL");
}
