import { describe, expect, it } from "vitest"
import { canBePersonal } from "./personal"

describe("canBePersonal", () => {
  it.each([
    ["expense", "person", true],
    ["expense", "joint", false],
    ["income", "person", false],
    ["investment", "person", false],
  ] as const)("%s paid by %s → %s", (type, kind, want) => {
    expect(canBePersonal(type, { kind })).toBe(want)
  })

  it("is false while the payer is unknown", () => {
    expect(canBePersonal("expense", undefined)).toBe(false)
  })
})
