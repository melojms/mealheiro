import { addTags, normalizeTag } from "./tags"

describe("normalizeTag", () => {
  it.each([["  Trip ", "trip"], ["Summer  Holiday", "summer holiday"], ["   ", ""]])("%s -> %s", (raw, want) =>
    expect(normalizeTag(raw)).toBe(want),
  )
})

describe("addTags", () => {
  it.each<[string[], string, string[]]>([
    [[], "Trip", ["trip"]],
    [["trip"], "TRIP", ["trip"]],
    [["trip"], "a, b ,,c", ["trip", "a", "b", "c"]],
    [[], "  ", []],
    [[], "x,x", ["x"]],
  ])("%j + %s", (tags, raw, want) => expect(addTags(tags, raw)).toEqual(want))
})
