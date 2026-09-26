// Captures the README screenshots into docs/screenshots/ from a running demo instance.
// Usage: seed an empty instance with `node e2e/seed-demo.mjs`, then `make screenshots`
// (or `node e2e/screenshots.mjs [baseURL]`, default http://localhost:7447).
// PNGs are palette-quantized with ImageMagick when `convert` is available.
import { chromium } from "@playwright/test"
import { execFileSync } from "node:child_process"
import { mkdirSync, statSync } from "node:fs"
import { dirname, join } from "node:path"
import { fileURLToPath } from "node:url"

const base = process.argv[2] ?? process.env.BASE_URL ?? "http://localhost:7447"
const out = join(dirname(fileURLToPath(import.meta.url)), "..", "docs", "screenshots")
mkdirSync(out, { recursive: true })

const MOBILE = { viewport: { width: 390, height: 844 }, deviceScaleFactor: 3, isMobile: true, hasTouch: true }
const DESKTOP = { viewport: { width: 1440, height: 900 }, deviceScaleFactor: 2 }

// Screenshot hygiene: no scrollbars, toasts, carets or CSS motion.
const CSS = `
  ::-webkit-scrollbar { display: none !important; }
  html { scrollbar-width: none !important; }
  [data-sonner-toaster] { display: none !important; }
  *, *::before, *::after {
    animation-duration: 0s !important; animation-delay: 0s !important;
    transition-duration: 0s !important; transition-delay: 0s !important;
    caret-color: transparent !important;
  }`

const people = await (await fetch(`${base}/api/people`)).json()
if (!people.length) throw new Error(`no people at ${base}: seed it first with e2e/seed-demo.mjs`)
const payer = people[0].id

const browser = await chromium.launch()

async function shoot(name, device, theme, act) {
  const ctx = await browser.newContext({ ...device, colorScheme: theme, reducedMotion: "reduce", serviceWorkers: "block" })
  await ctx.addInitScript(
    ({ theme, payer, css }) => {
      localStorage.setItem("mealheiro.theme", theme)
      localStorage.setItem("mealheiro.default-payer", String(payer))
      document.addEventListener("DOMContentLoaded", () => {
        const s = document.createElement("style")
        s.textContent = css
        document.head.append(s)
      })
    },
    { theme, payer, css: CSS },
  )
  const page = await ctx.newPage()
  await act(page)
  await settle(page)
  const file = join(out, `${name}.png`)
  await page.screenshot({ path: file })
  await ctx.close()
  optimize(file)
}

async function settle(page) {
  await page.waitForLoadState("networkidle")
  await page.evaluate(() => document.fonts.ready.then(() => undefined))
  await page.waitForFunction(() => !document.querySelector('[data-slot="skeleton"], .animate-pulse'))
  await page.waitForTimeout(1500) // recharts animates in JS, not CSS
}

function optimize(file) {
  const before = statSync(file).size
  try {
    execFileSync("convert", [file, "-strip", "-dither", "None", "-colors", "256", `PNG8:${file}`])
  } catch {
    console.warn("convert not found: keeping unoptimized PNG")
  }
  console.log(`${file.slice(out.length + 1)}: ${(before / 1024).toFixed(0)} KB -> ${(statSync(file).size / 1024).toFixed(0)} KB`)
}

// Mobile (dark)
await shoot("mobile-add", MOBILE, "dark", async (page) => {
  await page.goto(`${base}/`)
  const keypad = page.getByLabel("Amount keypad")
  for (const key of "42,50") {
    await keypad.getByRole("button", { name: key === "," ? "Decimal comma" : key, exact: true }).click()
  }
  await page.getByRole("radiogroup", { name: "Category" }).getByRole("radio", { name: "Groceries" }).click()
})
await shoot("mobile-month", MOBILE, "dark", (page) => page.goto(`${base}/month`))
await shoot("mobile-charts", MOBILE, "dark", (page) => page.goto(`${base}/charts`))
// Further down Month: the "To confirm" inbox of pending estimates, with Budgets below.
await shoot("mobile-budgets", MOBILE, "dark", async (page) => {
  await page.goto(`${base}/month`)
  await settle(page)
  await page.getByText(/To confirm/).first().evaluate((el) => {
    const card = el.closest('[data-slot="card"]') ?? el
    window.scrollTo(0, card.getBoundingClientRect().top + window.scrollY - 16)
  })
})

// Desktop (hero in both themes)
for (const theme of ["dark", "light"]) {
  await shoot(`desktop-month-${theme}`, DESKTOP, theme, (page) => page.goto(`${base}/month`))
}
await shoot("desktop-charts", DESKTOP, "dark", (page) => page.goto(`${base}/charts`))
await shoot("desktop-entries", DESKTOP, "dark", (page) => page.goto(`${base}/entries?type=expense&payer=${payer}`))
await shoot("desktop-recurring", DESKTOP, "dark", (page) => page.goto(`${base}/settings/recurring`))

await browser.close()
