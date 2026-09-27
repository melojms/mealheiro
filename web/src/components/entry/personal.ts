import type { EntryType, Person } from "@/lib/types"

/** Mirrors the API rule: only expenses paid by a person (never Joint) can be personal. */
export function canBePersonal(type: EntryType, payer: Pick<Person, "kind"> | undefined): boolean {
  return type === "expense" && payer?.kind === "person"
}
