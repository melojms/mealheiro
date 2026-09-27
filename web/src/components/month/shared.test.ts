import { describe, expect, it } from "vitest"
import { splitPercents } from "./shared"

describe("splitPercents", () => {
  it("returns null when nothing was shared", () => {
    expect(splitPercents([0, 0])).toBeNull()
    expect(splitPercents([])).toBeNull()
  })

  it("splits proportionally", () => {
    expect(splitPercents([105000, 105000])).toEqual([50, 50])
    expect(splitPercents([123000, 117000])).toEqual([51, 49])
  })

  it("gives 100/0 when one person paid everything", () => {
    expect(splitPercents([0, 7000])).toEqual([0, 100])
  })

  it("always sums to 100 (largest remainder)", () => {
    // 1/3 each would round to 33+33+33 = 99.
    expect(splitPercents([1, 1, 1])).toEqual([34, 33, 33])
    // 50.5 / 49.5 would round to 51 + 50 = 101.
    expect(splitPercents([505, 495])).toEqual([51, 49])
  })
})
