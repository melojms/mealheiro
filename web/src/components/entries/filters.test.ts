import { activeFilterCount, datePreset, readUrlFilters, toApiFilters, writeUrlFilters } from "./filters"

const today = "2026-03-15"

describe("readUrlFilters / writeUrlFilters", () => {
  it("reads known keys, trims and drops blanks", () => {
    const sp = new URLSearchParams("q=%20coffee%20&type=expense&tag=&foo=bar")
    expect(readUrlFilters(sp)).toEqual({ q: "coffee", type: "expense" })
  })

  it("writes filters and keeps unrelated params", () => {
    const sp = new URLSearchParams("foo=bar&q=old&tag=x")
    const next = writeUrlFilters(sp, { q: "new", min: "5" })
    expect(next.get("foo")).toBe("bar")
    expect(next.get("q")).toBe("new")
    expect(next.get("min")).toBe("5")
    expect(next.has("tag")).toBe(false)
  })
})

describe("toApiFilters", () => {
  it("returns {} with no filters (all time)", () => expect(toApiFilters({}, today)).toEqual({}))

  it("maps everything", () => {
    expect(
      toApiFilters(
        { q: "pingo", type: "expense", cat: "12", payer: "2", tag: "Holidays", min: "10,50", max: "1.234,56", range: "this-month" },
        today,
      ),
    ).toEqual({
      q: "pingo", type: "expense", category_id: 12, payer_id: 2, tag: "holidays",
      min_cents: 1050, max_cents: 123456, from: "2026-03-01", to: "2026-03-31",
    })
  })

  it("drops invalid values", () => {
    expect(toApiFilters({ type: "gift", cat: "abc", payer: "-1", min: "x", max: "" }, today)).toEqual({})
  })

  it("maps sharing and drops unknown values", () => {
    expect(toApiFilters({ sharing: "shared" }, today)).toEqual({ sharing: "shared" })
    expect(toApiFilters({ sharing: "personal" }, today)).toEqual({ sharing: "personal" })
    expect(toApiFilters({ sharing: "both" }, today)).toEqual({})
  })

  it("counts sharing as an active filter", () => expect(activeFilterCount({ sharing: "shared" })).toBe(1))

  it("keeps a zero min amount", () => expect(toApiFilters({ min: "0" }, today)).toEqual({ min_cents: 0 }))

  it("uses custom from/to and ignores malformed dates", () => {
    expect(toApiFilters({ range: "custom", from: "2026-01-10", to: "bad" }, today)).toEqual({ from: "2026-01-10" })
  })

  it("treats bare from/to as custom", () => {
    expect(datePreset({ to: "2026-02-01" })).toBe("custom")
    expect(toApiFilters({ to: "2026-02-01" }, today)).toEqual({ to: "2026-02-01" })
  })

  it("ignores unknown presets", () => {
    expect(datePreset({ range: "forever" })).toBeUndefined()
    expect(toApiFilters({ range: "forever" }, today)).toEqual({})
  })
})

describe("activeFilterCount", () => {
  it("counts groups, not keys, and ignores q", () => {
    expect(activeFilterCount({})).toBe(0)
    expect(activeFilterCount({ q: "x" })).toBe(0)
    expect(activeFilterCount({ min: "1", max: "2" })).toBe(1)
    expect(activeFilterCount({ type: "expense", tag: "a", range: "this-year", payer: "1", cat: "3" })).toBe(5)
  })
})
