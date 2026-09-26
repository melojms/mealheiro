import {
  addMonths, dayLabel, daysInMonth, isDate, isMonth, monthLabel, monthRange, presetRange, shortMonthLabel,
} from "./dates"

describe("addMonths", () => {
  it.each([
    ["2026-03", 1, "2026-04"],
    ["2026-03", -1, "2026-02"],
    ["2026-01", -1, "2025-12"],
    ["2025-12", 1, "2026-01"],
    ["2026-03", -12, "2025-03"],
    ["2026-03", 0, "2026-03"],
    ["2026-03", -27, "2023-12"],
  ])("%s %+d -> %s", (m, n, want) => expect(addMonths(m, n)).toBe(want))
})

describe("month helpers", () => {
  it("validates months and dates", () => {
    expect(isMonth("2026-03")).toBe(true)
    expect(isMonth("2026-13")).toBe(false)
    expect(isMonth("2026-3")).toBe(false)
    expect(isMonth(null)).toBe(false)
    expect(isDate("2026-03-15")).toBe(true)
    expect(isDate("2026-3-15")).toBe(false)
    expect(isDate("")).toBe(false)
  })

  it("knows month lengths incl. leap years", () => {
    expect(daysInMonth("2026-02")).toBe(28)
    expect(daysInMonth("2028-02")).toBe(29)
    expect(daysInMonth("2026-04")).toBe(30)
    expect(monthRange("2026-03")).toEqual({ from: "2026-03-01", to: "2026-03-31" })
  })

  it("formats labels", () => {
    expect(monthLabel("2026-03")).toBe("March 2026")
    expect(shortMonthLabel("2026-03")).toBe("Mar")
    expect(shortMonthLabel("2026-01")).toBe("Jan '26")
  })
})

describe("dayLabel", () => {
  const today = "2026-03-15"
  it.each([
    ["2026-03-15", "Today"],
    ["2026-03-14", "Yesterday"],
    ["2026-03-01", "Sun, 1 Mar"],
    ["2025-12-31", "Wed, 31 Dec 2025"],
  ])("%s -> %s", (d, want) => expect(dayLabel(d, today)).toBe(want))

  it("handles yesterday across a month boundary", () => {
    expect(dayLabel("2026-02-28", "2026-03-01")).toBe("Yesterday")
  })
})

describe("presetRange", () => {
  const today = "2026-03-15"
  it.each([
    ["this-month", { from: "2026-03-01", to: "2026-03-31" }],
    ["last-month", { from: "2026-02-01", to: "2026-02-28" }],
    ["last-3-months", { from: "2026-01-01", to: "2026-03-31" }],
    ["this-year", { from: "2026-01-01", to: "2026-12-31" }],
  ] as const)("%s", (p, want) => expect(presetRange(p, today)).toEqual(want))

  it("last month wraps the year", () => {
    expect(presetRange("last-month", "2026-01-10")).toEqual({ from: "2025-12-01", to: "2025-12-31" })
  })

  it("custom passes through", () => {
    expect(presetRange("custom", today, { from: "2026-01-05" })).toEqual({ from: "2026-01-05", to: undefined })
  })
})
