import { addDays, addMonths, dateChip, formatDay, formatMonth, fromISODate, monthEnd, toISODate } from "./dates"

describe("addDays", () => {
  it.each([
    ["2026-03-15", -1, "2026-03-14"],
    ["2026-03-01", -1, "2026-02-28"],
    ["2024-03-01", -1, "2024-02-29"],
    ["2026-01-01", -1, "2025-12-31"],
    ["2026-03-29", 1, "2026-03-30"], // DST change in Europe
    ["2026-12-31", 1, "2027-01-01"],
  ])("%s %+i -> %s", (iso, n, want) => expect(addDays(iso, n)).toBe(want))
})

describe("dateChip", () => {
  it.each([
    ["2026-03-15", "today"],
    ["2026-03-14", "yesterday"],
    ["2026-03-13", "other"],
    ["2026-03-16", "other"],
  ])("%s", (iso, want) => expect(dateChip(iso, "2026-03-15")).toBe(want))
})

describe("formatDay", () => {
  const today = "2026-03-15"
  it.each([
    ["2026-03-15", "Today"],
    ["2026-03-14", "Yesterday"],
    ["2026-03-02", "2 Mar"],
    ["2025-12-24", "24 Dec 2025"],
    ["2026-09-12", "12 Sep"],
  ])("%s -> %s", (iso, want) => expect(formatDay(iso, today)).toBe(want))
})

describe("ISO <-> Date", () => {
  it("round-trips local dates", () => {
    expect(toISODate(fromISODate("2026-03-29"))).toBe("2026-03-29")
    expect(toISODate(new Date(2026, 0, 5))).toBe("2026-01-05")
  })
})

describe("months", () => {
  it.each([
    ["2026-03", -1, "2026-02"],
    ["2026-01", -1, "2025-12"],
    ["2026-12", 1, "2027-01"],
    ["2026-03", -14, "2025-01"],
  ])("addMonths(%s, %i) -> %s", (m, n, want) => expect(addMonths(m, n)).toBe(want))

  it.each([
    ["2026-02", "2026-02-28"],
    ["2024-02", "2024-02-29"],
    ["2026-12", "2026-12-31"],
    ["2026-04", "2026-04-30"],
  ])("monthEnd(%s) -> %s", (m, want) => expect(monthEnd(m)).toBe(want))

  it("formats", () => expect(formatMonth("2026-03")).toBe("Mar 2026"))
})
