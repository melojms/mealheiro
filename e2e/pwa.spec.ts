import { expect, test, type Response } from "@playwright/test"

// Chromium only (CDP). Needs the production build: the service worker is not registered in dev.

test("installable: manifest parses and Chrome reports no installability errors", async ({ page }) => {
  await page.goto("/")
  await page.evaluate(() => navigator.serviceWorker.ready.then(() => undefined))

  const cdp = await page.context().newCDPSession(page)
  const manifest = await cdp.send("Page.getAppManifest")
  expect(manifest.url).toMatch(/\/manifest\.webmanifest$/)
  expect(manifest.errors).toEqual([])

  const { installabilityErrors } = await cdp.send("Page.getInstallabilityErrors")
  expect(installabilityErrors).toEqual([])
})

test("offline navigation shows the offline page; /api is never served from the cache", async ({ page, context }) => {
  const api: Response[] = []
  page.on("response", (r) => {
    if (new URL(r.url()).pathname.startsWith("/api/")) api.push(r)
  })

  await page.goto("/")
  await page.waitForFunction(() => navigator.serviceWorker.controller !== null)

  await context.setOffline(true)
  await page.goto("/month")
  await expect(page.getByText("You're offline — reconnect to add expenses.")).toBeVisible()
  // No stale data: API calls fail offline instead of being answered from a cache.
  expect(await page.evaluate(() => fetch("/api/meta").then(() => "ok", () => "failed"))).toBe("failed")

  await context.setOffline(false)
  const back = page.waitForResponse((r) => new URL(r.url()).pathname.startsWith("/api/"))
  await page.getByRole("button", { name: "Try again" }).click()
  expect((await back).ok()).toBe(true)
  await expect(page).toHaveTitle("Mealheiro")

  expect(api.length).toBeGreaterThan(0)
  for (const r of api) expect(r.fromServiceWorker(), r.url()).toBe(false)

  // The cache holds only the precached offline shell, never app data.
  const cached = await page.evaluate(async () => {
    const paths: string[] = []
    for (const name of await caches.keys()) {
      for (const req of await (await caches.open(name)).keys()) paths.push(new URL(req.url).pathname)
    }
    return paths.sort()
  })
  expect(cached).toEqual(["/favicon.svg", "/icon-192.png", "/icon-512.png", "/offline.html"])
})
