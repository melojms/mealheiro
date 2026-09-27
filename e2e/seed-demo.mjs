// Fills a running mealheiro with ~12 months of realistic demo data through the public API.
// Usage: node e2e/seed-demo.mjs [baseURL]   (default http://localhost:7447). Use on an empty DB only.
const base = process.argv[2] ?? "http://localhost:7447"

async function call(method, path, body) {
  const res = await fetch(base + path, {
    method,
    headers: body ? { "Content-Type": "application/json" } : undefined,
    body: body ? JSON.stringify(body) : undefined,
  })
  if (!res.ok) throw new Error(`${method} ${path}: ${res.status} ${await res.text()}`)
  return res.status === 204 ? null : res.json()
}

let seed = 42
const rand = () => ((seed = (seed * 16807) % 2147483647) / 2147483647)
const between = (a, b) => Math.round(a + rand() * (b - a))
const pad = (n) => String(n).padStart(2, "0")

const meta = await call("GET", "/api/meta")
const [curY, curM] = meta.month.split("-").map(Number)
const today = Number(meta.today.slice(8))
const cats = await call("GET", "/api/categories?include_archived=true")
const id = (name) => {
  const c = cats.find((c) => c.name === name)
  if (!c) throw new Error(`category ${name} missing`)
  return c.id
}

await call("PATCH", "/api/people/1", { name: "João" })
await call("PATCH", "/api/people/2", { name: "Maria" })

const startMonth = (() => {
  const d = new Date(Date.UTC(curY, curM - 1 - 11, 1))
  return `${d.getUTCFullYear()}-${pad(d.getUTCMonth() + 1)}`
})()

// Recurring templates (generation creates the monthly entries back to startMonth).
const templates = [
  { type: "expense", category_id: id("Mortgage"), payer_id: 3, amount_cents: 65000, variable: false, note: "Crédito habitação" },
  { type: "expense", category_id: id("Management"), payer_id: 3, amount_cents: 4500, variable: false },
  { type: "expense", category_id: id("Internet & Phone"), payer_id: 3, amount_cents: 4999, variable: false, note: "Fibra + 2 telemóveis" },
  { type: "expense", category_id: id("Subscriptions"), payer_id: 1, amount_cents: 1799, variable: false, note: "Streaming" },
  { type: "expense", category_id: id("Electricity"), payer_id: 3, amount_cents: 6500, variable: true },
  { type: "income", category_id: id("Salary"), payer_id: 1, amount_cents: 210000, variable: true },
  { type: "income", category_id: id("Salary"), payer_id: 2, amount_cents: 185000, variable: true },
  { type: "investment", category_id: id("ETFs/Stocks"), payer_id: 3, amount_cents: 40000, variable: false },
  { type: "investment", category_id: id("BTC"), payer_id: 1, amount_cents: 10000, variable: false },
]
for (const t of templates) await call("POST", "/api/templates", { ...t, start_month: startMonth })

// Confirm past pending entries with slightly varying amounts; leave the current month pending.
for (const p of await call("GET", "/api/pending")) {
  if (p.date.slice(0, 7) === meta.month) continue
  const jitter = p.type === "income" ? between(-2000, 6000) : between(-1500, 2500)
  await call("POST", `/api/entries/${p.id}/confirm`, { amount_cents: Math.max(100, p.amount_cents + jitter) })
}

// Variable day-to-day spending.
const spend = [
  ["Groceries", 6, 2500, 11000, ["", "Continente", "Pingo Doce", "Lidl", "Mercado"]],
  ["Eating out", 3, 1200, 6500, ["Jantar", "Almoço", "Pizza", ""]],
  ["Car", 2, 4000, 7500, ["Gasolina", "Via Verde"]],
  ["Transport", 2, 300, 2500, ["Metro", "Uber"]],
  ["Water", 1, 2200, 3200, [""]],
  ["Gas", 1, 1800, 4500, [""]],
  ["Health", 1, 800, 6000, ["Farmácia", "Consulta"]],
  ["Leisure", 1, 1000, 5000, ["Cinema", "Concerto", ""]],
  ["Clothes", 0.6, 2000, 12000, ["Zara", "Sapatos", ""]],
  ["Technology", 0.3, 3000, 45000, ["Auscultadores", "Portátil", "Cabo USB-C"]],
  ["Gifts", 0.4, 1500, 8000, ["Aniversário"]],
  ["Personal care", 0.8, 1000, 3500, ["Cabeleireiro"]],
]
for (let back = 11; back >= 0; back--) {
  const d = new Date(Date.UTC(curY, curM - 1 - back, 1))
  const y = d.getUTCFullYear()
  const m = d.getUTCMonth() + 1
  const days = back === 0 ? today : new Date(Date.UTC(y, m, 0)).getUTCDate()
  const seasonal = m === 12 ? 1.4 : m === 8 ? 1.2 : 1
  for (const [name, perMonth, min, max, notes] of spend) {
    const n = Math.max(0, Math.round(perMonth * seasonal * (back === 0 ? today / 30 : 1) + (rand() - 0.5)))
    for (let i = 0; i < n; i++) {
      const note = notes[between(0, notes.length - 1)]
      const tags = name === "Eating out" && m === 8 ? ["holiday-" + y] : []
      await call("POST", "/api/entries", {
        type: "expense",
        date: `${y}-${pad(m)}-${pad(between(1, days))}`,
        amount_cents: between(min, max),
        category_id: id(name),
        payer_id: between(1, 3),
        note,
        tags,
      })
    }
  }
  if (m === 12 || m === 6) {
    await call("POST", "/api/entries", { type: "income", date: `${y}-${pad(m)}-15`, amount_cents: 150000, category_id: id("Bonus"), payer_id: 1 })
  }
}

await call("PUT", "/api/budgets", { category_id: id("Groceries"), amount_cents: 45000 })
await call("PUT", "/api/budgets", { category_id: id("Eating out"), amount_cents: 12000 })
await call("PUT", "/api/budgets", { category_id: id("Leisure"), amount_cents: 6000 })
await call("PUT", "/api/budgets", { category_id: null, amount_cents: 220000 })
console.log("demo data seeded")
