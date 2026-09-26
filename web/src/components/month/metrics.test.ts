import {
  budgetState, centsToInput, delta, deltaTone, formatPoints, formatSignedPct, ratePoints, share, sumAmount, sumPending,
} from "./metrics"
import type { CategoryAmount } from "@/lib/types"

describe("deltaTone", () => {
  it.each([
    [10, true, "good"],
    [10, false, "bad"],
    [-10, true, "bad"],
    [-10, false, "good"],
    [0, true, "neutral"],
    [0, false, "neutral"],
  ] as const)("diff %d upIsGood=%s -> %s", (d, up, want) => expect(deltaTone(d, up)).toBe(want))
})

describe("delta", () => {
  it("computes diff, pct and tone", () => {
    expect(delta(12000, 10000, false)).toEqual({ diff: 2000, pct: 0.2, tone: "bad" })
    expect(delta(8000, 10000, true)).toEqual({ diff: -2000, pct: -0.2, tone: "bad" })
  })
  it("has no pct without a baseline", () => {
    expect(delta(500, 0, false)).toEqual({ diff: 500, pct: null, tone: "bad" })
    expect(delta(0, 0, true).pct).toBeNull()
  })
})

describe("formatting", () => {
  it.each([
    [0.2, "+20%"],
    [-0.084, "−8%"],
    [0.004, "0%"],
    [null, null],
    [Infinity, null],
  ])("formatSignedPct(%s)", (p, want) => expect(formatSignedPct(p)).toBe(want))

  it.each([
    [3.24, "+3.2 pts"],
    [-5, "−5.0 pts"],
    [0.01, "0 pts"],
    [null, null],
  ])("formatPoints(%s)", (p, want) => expect(formatPoints(p)).toBe(want))

  it("ratePoints", () => {
    expect(ratePoints(0.3, 0.25)).toBeCloseTo(5)
    expect(ratePoints(null, 0.2)).toBeNull()
    expect(ratePoints(0.2, null)).toBeNull()
  })

  it.each([
    [0, "0,00"],
    [5, "0,05"],
    [8420, "84,20"],
    [123456, "1234,56"],
  ])("centsToInput(%i) -> %s", (c, want) => expect(centsToInput(c)).toBe(want))
})

describe("budgetState", () => {
  it.each([
    [0, "ok"],
    [0.79, "ok"],
    [0.8, "warn"],
    [0.99, "warn"],
    [1, "over"],
    [1.7, "over"],
  ] as const)("%f -> %s", (r, want) => expect(budgetState(r)).toBe(want))
})

describe("sums", () => {
  const ca = (amount_cents: number, pending_cents: number) => ({ amount_cents, pending_cents }) as CategoryAmount
  it("sums amounts and pending", () => {
    expect(sumAmount([ca(100, 0), ca(250, 50)])).toBe(350)
    expect(sumPending([ca(100, 0), ca(250, 50)])).toBe(50)
    expect(sumAmount([])).toBe(0)
  })
  it("share guards zero totals", () => {
    expect(share(25, 100)).toBe(0.25)
    expect(share(25, 0)).toBe(0)
  })
})
