import { formatCents, parseAmount, pctChange, formatPct } from "./money"

describe("parseAmount", () => {
  it.each([
    ["12,50", 1250], ["12.50", 1250], ["12.5", 1250], ["7", 700], ["0,05", 5], [",5", 50],
    ["1 234,56", 123456], ["1.234,56", 123456], ["1,234.56", 123456], ["12 €", 1200], ["1.234", 123400],
  ])("%s -> %i", (input, want) => expect(parseAmount(input)).toBe(want))

  it.each(["", "abc", "-5", "1e3", "12,5a"])("rejects %s", (input) => expect(parseAmount(input)).toBeNull())
})

describe("formatCents", () => {
  it("formats pt-PT euros", () => {
    expect(formatCents(1250)).toBe("12,50 €")
    expect(formatCents(5)).toBe("0,05 €")
  })
})

describe("pct", () => {
  it("computes change", () => {
    expect(pctChange(100, 150)).toBe(0.5)
    expect(pctChange(0, 10)).toBeNull()
    expect(formatPct(0.256)).toBe("26%")
    expect(formatPct(null)).toBe("—")
  })
})
