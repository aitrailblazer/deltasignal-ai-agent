import {chromium} from "playwright";
import {mkdir, rename} from "node:fs/promises";
import {resolve} from "node:path";

const baseURL =
  process.env.DELTASIGNAL_SITE_URL ??
  "https://aitrailblazer.github.io/deltasignal-ai-agent/";
const outputDirectory = resolve(
  process.env.DELTASIGNAL_VIDEO_OUTPUT ?? "artifacts/competition-theater",
);
await mkdir(outputDirectory, {recursive: true});

const browser = await chromium.launch({channel: "chrome", headless: true});
const context = await browser.newContext({
  viewport: {width: 1920, height: 1080},
  recordVideo: {
    dir: outputDirectory,
    size: {width: 1920, height: 1080},
  },
});

const page = await context.newPage();
try {
  const target = new URL(baseURL);
  target.searchParams.set("competition", "1");
  await page.goto(target.toString(), {waitUntil: "networkidle"});
  await page.locator(".competition-theater").waitFor();
  await page.waitForTimeout(1800);

  for (let slide = 1; slide < 6; slide += 1) {
    await page.keyboard.press("ArrowRight");
    await page.waitForTimeout(6500);
  }
  await page.waitForTimeout(1800);

  const video = page.video();
  await page.close();
  const recordedPath = await video.path();
  const finalPath = resolve(outputDirectory, "DeltaSignal_Competition_Theater.webm");
  await rename(recordedPath, finalPath);
  console.log(finalPath);
} finally {
  await context.close();
  await browser.close();
}
