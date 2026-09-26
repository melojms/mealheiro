import { expect, test } from "@playwright/test"

interface Person { id: number; name: string }
interface MonthReport { kpis: { expense_cents: number } }
interface Entry { id: number; amount_cents: number; category_name: string }

const CATEGORY = "Groceries"
const AMOUNT = "7,31"
const CENTS = 731

const digits = (s: string) => s.replace(/\D/g, "")

test("quick-add an expense on a phone and see it counted in Month", async ({ page, request }) => {
  const people: Person[] = await (await request.get("/api/people")).json()
  const before: MonthReport = await (await request.get("/api/reports/month")).json()
  let created: Entry | undefined

  try {
    await page.goto("/")

    // First visit on this device asks who is paying.
    const prompt = page.getByText("Who's using this device?")
    if (await prompt.isVisible({ timeout: 3_000 }).catch(() => false)) {
      await page.getByRole("button", { name: people[0].name, exact: true }).click()
    }

    const keypad = page.getByLabel("Amount keypad")
    for (const key of AMOUNT) {
      await keypad.getByRole("button", { name: key === "," ? "Decimal comma" : key, exact: true }).click()
    }
    await expect(page.getByRole("status", { name: "Amount" })).toContainText(AMOUNT)

    await page.getByRole("radiogroup", { name: "Category" }).getByRole("radio", { name: CATEGORY }).click()

    const saved = page.waitForResponse((r) => r.url().endsWith("/api/entries") && r.request().method() === "POST")
    await page.getByRole("button", { name: new RegExp(`^Save ${AMOUNT}`) }).click()
    const res = await saved
    expect(res.status()).toBe(201)
    created = await res.json()

    await expect(page.getByText(new RegExp(`Saved ${AMOUNT}.*${CATEGORY}`))).toBeVisible()
    const recent = page.getByRole("region", { name: "Recent entries" })
    await expect(recent.getByRole("button").filter({ hasText: CATEGORY }).filter({ hasText: AMOUNT }).first()).toBeVisible()

    // Month counts it in the Expenses KPI.
    await page.getByRole("link", { name: "Month" }).click()
    await expect(page).toHaveURL(/\/month/)
    const expenses = page.getByRole("group", { name: "Expenses" })
    const want = String(before.kpis.expense_cents + CENTS)
    await expect.poll(async () => digits((await expenses.textContent()) ?? "").startsWith(want)).toBe(true)
  } finally {
    if (created) await request.delete(`/api/entries/${created.id}`)
  }
})
