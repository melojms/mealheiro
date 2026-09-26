/** Normalizes a tag like the server does (trim + lowercase + inner whitespace collapsed). */
export function normalizeTag(raw: string): string {
  return raw.trim().toLowerCase().replace(/\s+/g, " ")
}

/** Adds one or more comma-separated tags, skipping empties and duplicates. */
export function addTags(tags: string[], raw: string): string[] {
  const next = [...tags]
  for (const t of raw.split(",").map(normalizeTag)) {
    if (t && !next.includes(t)) next.push(t)
  }
  return next
}
