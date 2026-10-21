const test = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const http = require("node:http");
const path = require("node:path");

const dataDir = path.resolve(__dirname, "..", "..");
const width = 608;
const height = 512;

function startServer(posts) {
  const server = http.createServer((req, res) => {
    const url = new URL(req.url, "http://127.0.0.1");
    if (req.method === "POST" && url.pathname.startsWith("/game/")) {
      const chunks = [];
      req.on("data", (chunk) => chunks.push(chunk));
      req.on("end", () => {
        posts.push(JSON.parse(Buffer.concat(chunks).toString("utf8")));
        res.setHeader("Content-Type", "application/json");
        res.end(
          JSON.stringify({
            current: 1,
            grid: "#####\n#.O.#\n#####\n",
          }),
        );
      });
      return;
    }

    const files = {
      "/app/dist/bundle.js": {
        type: "application/javascript",
        file: path.join(dataDir, "app/dist/bundle.js"),
      },
      "/assets/tileset.png": {
        type: "image/png",
        file: path.join(dataDir, "assets/tileset.png"),
      },
    };
    if (req.method === "GET" && url.pathname.startsWith("/game/")) {
      res.setHeader("Content-Type", "text/html");
      res.end(fs.readFileSync(path.join(dataDir, "index.html")));
      return;
    }
    const asset = files[url.pathname];
    if (req.method === "GET" && asset) {
      res.setHeader("Content-Type", asset.type);
      res.end(fs.readFileSync(asset.file));
      return;
    }
    res.statusCode = 404;
    res.end("not found");
  });

  return new Promise((resolve) => {
    server.listen(0, "127.0.0.1", () => resolve(server));
  });
}

function waitFor(predicate, label) {
  const start = Date.now();
  return new Promise((resolve, reject) => {
    const timer = setInterval(() => {
      if (predicate()) {
        clearInterval(timer);
        resolve();
        return;
      }
      if (Date.now() - start > 10000) {
        clearInterval(timer);
        reject(new Error(`timed out waiting for ${label}`));
      }
    }, 50);
  });
}

test("mobile quadrant taps post the matching action", { timeout: 60000 }, async () => {
  const { chromium } = require("playwright");
  const posts = [];
  const server = await startServer(posts);
  const port = server.address().port;
  const browser = await chromium.launch({ headless: true });
  const pageErrors = [];
  try {
    const context = await browser.newContext({
      viewport: { width: 390, height: 844 },
      isMobile: true,
      hasTouch: true,
      deviceScaleFactor: 2,
    });
    const page = await context.newPage();
    page.on("pageerror", (error) => pageErrors.push(String(error)));
    await page.goto(`http://127.0.0.1:${port}/game/touch-test`, {
      waitUntil: "domcontentloaded",
    });
    await page.waitForFunction(() => {
      const canvas = document.querySelector("canvas");
      if (!canvas) {
        return false;
      }
      const rect = canvas.getBoundingClientRect();
      return rect.width > 0 && rect.height > 0;
    });
    await waitFor(() => posts.some((post) => post.action === "."), "initial post");

    const canvas = page.locator("canvas");
    const taps = [
      [500, 256, "right"],
      [80, 256, "left"],
      [304, 40, "up"],
      [304, 400, "down"],
    ];
    const actions = [];
    for (const [gameX, gameY, expected] of taps) {
      const box = await canvas.boundingBox();
      const before = posts.length;
      await canvas.tap({
        position: {
          x: box.width * (gameX / width),
          y: box.height * (gameY / height),
        },
      });
      await waitFor(() => posts.length > before, `${expected} post`);
      const added = posts.slice(before).map((post) => post.action);
      assert.deepEqual(added, [expected], `page errors: ${pageErrors.join("\n")}`);
      actions.push(expected);
    }
    assert.deepEqual(actions, ["right", "left", "up", "down"]);
  } finally {
    await browser.close();
    await new Promise((resolve, reject) => {
      server.close((error) => (error ? reject(error) : resolve()));
    });
  }
});
