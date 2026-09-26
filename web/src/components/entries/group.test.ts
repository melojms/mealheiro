import { groupByDate } from "./group"
import type { Entry } from "@/lib/types"

const e = (id: number, date: string, type: Entry["type"] = "expense", amount_cents = 100) =>
  ({ id, date, type, amount_cents }) as Entry

describe("groupByDate", () => {
  it("returns [] for no entries", () => expect(groupByDate([])).toEqual([]))

  it("buckets consecutive dates and sums expenses only", () => {
    const got = groupByDate([
      e(5, "2026-03-15", "expense", 1000),
      e(4, "2026-03-15", "income", 99999),
      e(3, "2026-03-15", "expense", 250),
      e(2, "2026-03-10"),
      e(1, "2026-02-28", "investment", 5000),
    ])
    expect(got.map((g) => [g.date, g.entries.map((x) => x.id), g.expense_cents])).toEqual([
      ["2026-03-15", [5, 4, 3], 1250],
      ["2026-03-10", [2], 100],
      ["2026-02-28", [1], 0],
    ])
  })

  it("keeps pages appended in order without duplicating headers", () => {
    const page1 = [e(3, "2026-03-15"), e(2, "2026-03-14")]
    const page2 = [e(1, "2026-03-14")]
    expect(groupByDate([...page1, ...page2]).map((g) => g.entries.length)).toEqual([1, 2])
  })
})
