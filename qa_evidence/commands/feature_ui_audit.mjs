import { chromium } from "playwright";
import fs from "node:fs/promises";
import path from "node:path";

const root = process.cwd();
const evidenceRoot = path.join(root, "qa_evidence");
const screenshotDir = path.join(evidenceRoot, "screenshots");
const logDir = path.join(evidenceRoot, "logs");
await fs.mkdir(screenshotDir, { recursive: true });
await fs.mkdir(logDir, { recursive: true });

const browser = await chromium.launch({
  headless: true,
  executablePath: "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
});
const results = [];

const pages = [
  { id: "landing", url: "http://127.0.0.1:8881/" },
  { id: "tour", url: "http://127.0.0.1:8881/?tour=1" },
  {
    id: "aapl-demo",
    url: "http://127.0.0.1:8881/client-demo/google-cloud-aapl/Google_Cloud_AAPL_DeltaSignal_MCP_Client_Demo_2026_08_06.html",
  },
  {
    id: "aapl-presentation",
    url: "http://127.0.0.1:8881/client-demo/google-cloud-aapl/Google_Cloud_AAPL_DeltaSignal_MCP_Client_Demo_2026_08_06.html?presentation=1",
  },
  {
    id: "architecture",
    url: "http://127.0.0.1:8881/docs/DeltaSignal_Daily_Backend_to_Google_Agents_Architecture_2026_08_06.html",
  },
  {
    id: "comprehension-audit",
    url: "http://127.0.0.1:8881/docs/DeltaSignal_First_10_Seconds_Comprehension_Audit_2026_08_07.html",
  },
  { id: "cloud-root", url: "http://127.0.0.1:8090/" },
  { id: "cloud-demo", url: "http://127.0.0.1:8090/demo" },
  { id: "cloud-run", url: "http://127.0.0.1:8090/demo/run" },
];

async function inspect(spec, viewport) {
  const context = await browser.newContext({ viewport });
  const page = await context.newPage();
  const response = await page.goto(spec.url, { waitUntil: "networkidle", timeout: 30_000 });
  await page.waitForTimeout(700);
  const metrics = await page.evaluate(() => {
    const app = document.querySelector("deltasignal-site");
    const scope = app?.shadowRoot || document;
    const visible = (node) => {
      const style = getComputedStyle(node);
      const rect = node.getBoundingClientRect();
      return style.visibility !== "hidden" && style.display !== "none" && rect.width > 0 && rect.height > 0;
    };
    const links = [...scope.querySelectorAll("a")].filter(visible);
    const buttons = [...scope.querySelectorAll("button")].filter(visible);
    const headings = [...scope.querySelectorAll("h1,h2,h3")].filter(visible).map((n) => n.textContent.trim());
    const tinyTargets = [...links, ...buttons].filter((n) => {
      const r = n.getBoundingClientRect();
      return r.width < 36 || r.height < 36;
    }).map((n) => ({ text: n.textContent.trim().slice(0, 80), width: Math.round(n.getBoundingClientRect().width), height: Math.round(n.getBoundingClientRect().height) }));
    const text = (scope.textContent || document.body.innerText).replace(/\s+/g, " ").trim();
    return {
      title: document.title,
      h1: scope.querySelector("h1")?.textContent?.trim() || "",
      headings,
      bodyTextSample: text.slice(0, 1000),
      bodyTextLength: text.length,
      scrollWidth: document.documentElement.scrollWidth,
      clientWidth: document.documentElement.clientWidth,
      overflowX: document.documentElement.scrollWidth > document.documentElement.clientWidth + 1,
      links: links.length,
      buttons: buttons.length,
      images: [...scope.querySelectorAll("img")].filter(visible).length,
      svgs: [...scope.querySelectorAll("svg")].filter(visible).length,
      canvases: [...scope.querySelectorAll("canvas")].filter(visible).length,
      tinyTargets,
      hasMain: Boolean(scope.querySelector("main")),
      hasNav: Boolean(scope.querySelector("nav")),
      mentionsCompetition: /challenge|competition|judge/i.test(text),
      mentionsApple: /apple|aapl/i.test(text),
      mentionsHUT: /\bhut\b|hut 8/i.test(text),
    };
  });
  const suffix = `${viewport.width}x${viewport.height}`;
  const screenshot = path.join(screenshotDir, `${spec.id}-${suffix}.png`);
  await page.screenshot({ path: screenshot, fullPage: spec.id !== "tour" && spec.id !== "aapl-presentation" });
  results.push({
    id: spec.id,
    url: spec.url,
    viewport,
    httpStatus: response?.status() ?? null,
    screenshot: path.relative(root, screenshot),
    ...metrics,
  });
  await context.close();
}

for (const spec of pages) {
  await inspect(spec, { width: 1440, height: 1000 });
  await inspect(spec, { width: 390, height: 844 });
}

// Exercise the complete product-tour navigation separately.
{
  const context = await browser.newContext({ viewport: { width: 1920, height: 1080 } });
  const page = await context.newPage();
  await page.goto("http://127.0.0.1:8881/?tour=1", { waitUntil: "networkidle" });
  await page.waitForTimeout(700);
  const slides = [];
  for (let index = 0; index < 6; index += 1) {
    const snapshot = await page.evaluate(() => ({
      heading: document.querySelector("deltasignal-site")?.shadowRoot?.querySelector(".competition-copy h1")?.textContent?.trim() || "",
      body: document.querySelector("deltasignal-site")?.shadowRoot?.querySelector(".competition-copy p")?.textContent?.trim() || "",
      index: document.querySelector("deltasignal-site")?.shadowRoot?.querySelector(".slide-index")?.textContent?.trim() || "",
      overflowX: document.documentElement.scrollWidth > document.documentElement.clientWidth + 1,
      overflowY: document.documentElement.scrollHeight > document.documentElement.clientHeight + 1,
    }));
    slides.push(snapshot);
    if (index < 5) {
      await page.getByRole("button", { name: "Next slide" }).click();
      await page.waitForTimeout(120);
    }
  }
  await page.screenshot({ path: path.join(screenshotDir, "tour-final-1920x1080.png") });
  results.push({ id: "tour-navigation", slides });
  await context.close();
}

await browser.close();

const output = path.join(logDir, "feature_ui_audit.json");
await fs.writeFile(output, JSON.stringify({ generatedAt: new Date().toISOString(), results }, null, 2));
console.log(`Wrote ${output}`);
console.log(JSON.stringify(results.map(({ id, viewport, httpStatus, overflowX, mentionsCompetition, mentionsApple, mentionsHUT }) => ({
  id,
  viewport,
  httpStatus,
  overflowX,
  mentionsCompetition,
  mentionsApple,
  mentionsHUT,
})), null, 2));
