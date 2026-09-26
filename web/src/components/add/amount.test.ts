import { amountCents, centsToInput, displayParts, keyFromKeyboard, pressKey, type KeypadKey } from "./amount"

function type(keys: KeypadKey[], start = ""): string {
  return keys.reduce(pressKey, start)
}

describe("pressKey", () => {
  it.each<[string, KeypadKey[], string]>([
    ["digits append", ["1", "2"], "12"],
    ["decimal comma", ["1", "2", ",", "5"], "12,5"],
    ["max two decimals", ["1", ",", "2", "3", "4"], "1,23"],
    ["single comma only", ["1", ",", ",", "5"], "1,5"],
    ["leading comma becomes 0,", [",", "5"], "0,5"],
    ["leading zero replaced", ["0", "7"], "7"],
    ["zero stays zero", ["0", "0"], "0"],
    ["zero then comma", ["0", ",", "0", "5"], "0,05"],
    ["backspace", ["1", "2", "backspace"], "1"],
    ["backspace on empty", ["backspace"], ""],
    ["backspace removes comma", ["1", ",", "backspace", "5"], "15"],
    ["clear", ["1", "2", "clear"], ""],
    ["integer digits capped", ["1", "2", "3", "4", "5", "6", "7", "8"], "1234567"],
    ["decimals allowed after cap", ["1", "2", "3", "4", "5", "6", "7", ",", "9", "9"], "1234567,99"],
  ])("%s", (_name, keys, want) => expect(type(keys)).toBe(want))
})

describe("keyFromKeyboard", () => {
  it.each([
    ["5", "5"], [",", ","], [".", ","], ["Backspace", "backspace"], ["Escape", "clear"], ["Delete", "clear"],
    ["a", null], ["Enter", null], ["F1", null],
  ])("%s -> %s", (key, want) => expect(keyFromKeyboard(key)).toBe(want))
})

describe("amountCents", () => {
  it.each<[string, number | null]>([
    ["", null], ["0", null], ["0,", null], ["0,00", null], ["12", 1200], ["12,", 1200], ["12,5", 1250], ["0,05", 5],
    ["1234567,99", 123456799],
  ])("%s -> %s", (input, want) => expect(amountCents(input)).toBe(want))
})

describe("displayParts", () => {
  it.each([
    ["", "0", ",00"], ["12", "12", ",00"], ["12,", "12,", "00"], ["12,5", "12,5", "0"], ["12,50", "12,50", ""],
    ["0,", "0,", "00"], ["1234", "1234", ",00"], ["12345", "12 345", ",00"],
  ])("%s -> %s|%s", (input, main, ghost) => expect(displayParts(input)).toEqual({ main, ghost }))
})

describe("centsToInput", () => {
  it.each([[1250, "12,50"], [5, "0,05"], [100, "1,00"], [123456, "1234,56"]])("%i -> %s", (c, want) =>
    expect(centsToInput(c)).toBe(want),
  )
})
