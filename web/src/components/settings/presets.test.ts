import { matchPreset, presetRange, type ExportPreset } from "./presets"

describe("presetRange", () => {
  it.each<[ExportPreset, string, string, string]>([
    ["this_month", "2026-03-15", "2026-03-01", "2026-03-31"],
    ["this_month", "2026-02-10", "2026-02-01", "2026-02-28"],
    ["last_month", "2026-03-15", "2026-02-01", "2026-02-28"],
    ["last_month", "2026-01-05", "2025-12-01", "2025-12-31"],
    ["this_year", "2026-03-15", "2026-01-01", "2026-12-31"],
    ["last_year", "2026-03-15", "2025-01-01", "2025-12-31"],
    ["all", "2026-03-15", "", ""],
  ])("%s on %s", (preset, today, from, to) => expect(presetRange(preset, today)).toEqual({ from, to }))
})

describe("matchPreset", () => {
  const today = "2026-03-15"
  it("finds a preset", () => expect(matchPreset({ from: "2026-02-01", to: "2026-02-28" }, today)).toBe("last_month"))
  it("returns null for a custom range", () => expect(matchPreset({ from: "2026-02-02", to: "2026-02-28" }, today)).toBeNull())
})
