import { expect, test, type Page } from "@playwright/test"

// iPhone 12-15 width: the narrowest common phone the category tiles must fit.
test.use({ viewport: { width: 390, height: 844 } })

/** Category tile labels where a single word was split across lines (e.g. "Subscription / s"). */
async function splitWords(page: Page): Promise<string[]> {
  const grid = page.getByRole("radiogroup", { name: "Category" })
  await expect(grid.getByRole("radio").first()).toBeVisible()
  return grid.evaluate((el) => {
    const bad: string[] = []
    for (const label of el.querySelectorAll('[role="radio"] span')) {
      const text = label.firstChild
      if (!text || text.nodeType !== Node.TEXT_NODE) continue
      const value = text.textContent ?? ""
      for (const m of value.matchAll(/\S+/g)) {
        const range = document.createRange()
        range.setStart(text, m.index)
        range.setEnd(text, m.index + m[0].length)
        const lines = new Set([...range.getClientRects()].map((r) => Math.round(r.top)))
        if (lines.size > 1) bad.push(`${value.trim()} (${m[0]})`)
      }
    }
    return bad
  })
}

test("category labels never split a word on the Add screen", async ({ page }) => {
  await page.goto("/")
  expect(await splitWords(page)).toEqual([])

  // Selected tiles use a heavier weight: check the label with the longest word while active too.
  const radios = page.getByRole("radiogroup", { name: "Category" }).getByRole("radio")
  const names = await radios.allTextContents()
  const longestWord = (s: string) => Math.max(...s.trim().split(/\s+/).map((w) => w.length))
  const widest = names.reduce((a, b) => (longestWord(b) > longestWord(a) ? b : a)).trim()
  await radios.filter({ hasText: widest }).click()
  await expect(radios.filter({ hasText: widest })).toHaveAttribute("aria-checked", "true")
  expect(await splitWords(page)).toEqual([])
})

test("category labels never split a word in the edit dialog", async ({ page, request }) => {
  const people: { id: number }[] = await (await request.get("/api/people")).json()
  const cats: { id: number; parent_id: number | null; archived?: boolean }[] = await (await request.get("/api/categories")).json()
  const meta: { today: string } = await (await request.get("/api/meta")).json()
  const top = cats.find((c) => c.parent_id === null && !c.archived)!
  const res = await request.post("/api/entries", {
    data: { type: "expense", date: meta.today, amount_cents: 123, category_id: top.id, payer_id: people[0].id, note: "layout-check" },
  })
  expect(res.status()).toBe(201)
  const entry: { id: number } = await res.json()

  try {
    await page.goto("/entries?q=layout-check")
    await page.locator("section[aria-label] button").first().click()
    await expect(page.getByRole("dialog")).toBeVisible()
    expect(await splitWords(page)).toEqual([])
  } finally {
    await request.delete(`/api/entries/${entry.id}`)
  }
})
