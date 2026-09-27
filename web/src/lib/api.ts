// Typed client for docs/API.md. All functions throw ApiError on non-2xx.
import type {
  BudgetStatus, Budgets, Category, Entry, EntryFilters, EntryInput, EntryList, EntryType, Insight,
  Meta, MonthReport, PendingEntry, Person, SharedReport, TagCount, Template, TemplateInput, TrendsReport, YearReport,
} from "./types"

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

type Query = Record<string, string | number | boolean | undefined | null>

function qs(params?: Query): string {
  if (!params) return ""
  const sp = new URLSearchParams()
  for (const [k, v] of Object.entries(params)) {
    if (v === undefined || v === null || v === "") continue
    sp.set(k, String(v))
  }
  const s = sp.toString()
  return s ? `?${s}` : ""
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch(path, {
    method,
    headers: body === undefined ? undefined : { "Content-Type": "application/json" },
    body: body === undefined ? undefined : JSON.stringify(body),
  })
  if (!res.ok) {
    let msg = res.statusText
    try {
      msg = ((await res.json()) as { error?: string }).error ?? msg
    } catch {
      /* non-JSON error */
    }
    throw new ApiError(res.status, msg)
  }
  if (res.status === 204) return undefined as T
  return (await res.json()) as T
}

export const api = {
  meta: () => request<Meta>("GET", "/api/meta"),

  people: () => request<Person[]>("GET", "/api/people"),
  renamePerson: (id: number, name: string) => request<Person>("PATCH", `/api/people/${id}`, { name }),

  categories: (p?: { type?: EntryType; include_archived?: boolean }) =>
    request<Category[]>("GET", `/api/categories${qs(p)}`),
  createCategory: (b: { type: EntryType; name: string; parent_id?: number | null; icon?: string; color?: string }) =>
    request<Category>("POST", "/api/categories", b),
  updateCategory: (id: number, b: { name?: string; icon?: string; color?: string; archived?: boolean }) =>
    request<Category>("PATCH", `/api/categories/${id}`, b),
  deleteCategory: (id: number) => request<void>("DELETE", `/api/categories/${id}`),

  entries: (f?: EntryFilters) => request<EntryList>("GET", `/api/entries${qs(f as Query)}`),
  entry: (id: number) => request<Entry>("GET", `/api/entries/${id}`),
  createEntry: (b: EntryInput) => request<Entry>("POST", "/api/entries", b),
  updateEntry: (id: number, b: EntryInput) => request<Entry>("PUT", `/api/entries/${id}`, b),
  confirmEntry: (id: number, amount_cents?: number) =>
    request<Entry>("POST", `/api/entries/${id}/confirm`, amount_cents === undefined ? {} : { amount_cents }),
  deleteEntry: (id: number) => request<void>("DELETE", `/api/entries/${id}`),

  tags: (q?: string) => request<TagCount[]>("GET", `/api/tags${qs({ q })}`),

  templates: () => request<Template[]>("GET", "/api/templates"),
  createTemplate: (b: TemplateInput) => request<Template>("POST", "/api/templates", b),
  updateTemplate: (id: number, b: TemplateInput) => request<Template>("PUT", `/api/templates/${id}`, b),
  deleteTemplate: (id: number) => request<void>("DELETE", `/api/templates/${id}`),
  runRecurring: () => request<{ created: number }>("POST", "/api/recurring/run"),
  pending: () => request<PendingEntry[]>("GET", "/api/pending"),

  budgets: (month?: string) => request<Budgets>("GET", `/api/budgets${qs({ month })}`),
  setBudget: (category_id: number | null, amount_cents: number | null) =>
    request<void>("PUT", "/api/budgets", { category_id, amount_cents }),
  budgetStatus: (category_id: number, month?: string) =>
    request<BudgetStatus>("GET", `/api/budgets/status${qs({ category_id, month })}`),

  monthReport: (month?: string, payer_id?: number) =>
    request<MonthReport>("GET", `/api/reports/month${qs({ month, payer_id })}`),
  trends: (p?: { end?: string; months?: number; category_id?: number; payer_id?: number }) =>
    request<TrendsReport>("GET", `/api/reports/trends${qs(p)}`),
  yearReport: (year?: number, payer_id?: number) =>
    request<YearReport>("GET", `/api/reports/year${qs({ year, payer_id })}`),
  years: () => request<number[]>("GET", "/api/reports/years"),
  sharedReport: (p: { month: string } | { year: number }) =>
    request<SharedReport>("GET", `/api/reports/shared${qs(p)}`),
  insights: (month?: string, payer_id?: number) =>
    request<Insight[]>("GET", `/api/insights${qs({ month, payer_id })}`),

  exportUrl: (p: { from?: string; to?: string; types?: EntryType[] }) =>
    `/api/export.csv${qs({ from: p.from, to: p.to, types: p.types?.join(",") })}`,
  backupUrl: "/api/backup",
}
