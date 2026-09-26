import { useMutation, useQueryClient } from "@tanstack/react-query"
import { Check, CircleCheckBig, Inbox, Loader2, Pencil, TriangleAlert } from "lucide-react"
import { useState } from "react"
import { toast } from "sonner"
import { CategoryIcon } from "@/components/CategoryIcon"
import { mediumMonthLabel } from "@/components/filters/dates"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { InputGroup, InputGroupAddon, InputGroupInput } from "@/components/ui/input-group"
import { Skeleton } from "@/components/ui/skeleton"
import { api } from "@/lib/api"
import { formatCents, parseAmount } from "@/lib/money"
import type { Category, Entry, PendingEntry } from "@/lib/types"
import { centsToInput } from "./metrics"

function PendingRow({
  entry, category, onOpen,
}: { entry: PendingEntry; category?: Category; onOpen: () => void }) {
  const qc = useQueryClient()
  const [amount, setAmount] = useState(() => centsToInput(entry.amount_cents))
  const cents = parseAmount(amount)
  const invalid = cents === null || cents <= 0

  const confirm = useMutation({
    mutationFn: () => api.confirmEntry(entry.id, cents !== entry.amount_cents ? cents! : undefined),
    onSuccess: (e) => {
      toast.success(`${e.category_name} confirmed`, { description: formatCents(e.amount_cents) })
      qc.invalidateQueries()
    },
    onError: (err) => toast.error("Couldn't confirm entry", { description: err.message }),
  })

  const inputId = `pending-${entry.id}`
  return (
    <li className="py-3 first:pt-0 last:pb-0">
      <div className="flex items-center gap-3">
        <CategoryIcon icon={category?.icon ?? "circle"} color={category?.color ?? "#64748b"} size="sm" />
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-1.5">
            <span className="truncate text-sm font-medium">{entry.category_name}</span>
            {entry.stale && (
              <Badge variant="outline" className="border-amber-500/40 text-amber-700 dark:text-amber-400">
                <TriangleAlert /> Past month
              </Badge>
            )}
          </div>
          <p className="text-muted-foreground truncate text-xs">
            {mediumMonthLabel(entry.date.slice(0, 7))} · {entry.payer_name}
            {entry.note && ` · ${entry.note}`}
          </p>
        </div>
        <Button variant="ghost" size="icon-sm" onClick={onOpen} aria-label={`Edit ${entry.category_name} entry`}>
          <Pencil />
        </Button>
      </div>
      <form
        className="mt-2 flex gap-2 pl-10"
        onSubmit={(ev) => {
          ev.preventDefault()
          if (!invalid) confirm.mutate()
        }}
      >
        <label htmlFor={inputId} className="sr-only">
          Amount for {entry.category_name}
        </label>
        <InputGroup className="h-9 flex-1">
          <InputGroupInput
            id={inputId}
            inputMode="decimal"
            autoComplete="off"
            value={amount}
            onChange={(ev) => setAmount(ev.target.value)}
            onFocus={(ev) => ev.target.select()}
            aria-invalid={invalid}
            className="text-right font-medium tabular-nums"
          />
          <InputGroupAddon align="inline-end">€</InputGroupAddon>
        </InputGroup>
        <Button type="submit" size="lg" disabled={invalid || confirm.isPending}>
          {confirm.isPending ? <Loader2 className="animate-spin" /> : <Check />}
          Confirm
        </Button>
      </form>
    </li>
  )
}

/** "To confirm" inbox: pending estimates from variable recurring templates. */
export function PendingInbox({
  entries, isPending, byId, onOpen,
}: {
  entries: PendingEntry[] | undefined
  isPending: boolean
  byId: Map<number, Category>
  onOpen: (e: Entry) => void
}) {
  if (!isPending && !entries?.length) {
    return (
      <Card size="sm" className="flex-row items-center gap-3 px-4">
        <span className="flex size-9 shrink-0 items-center justify-center rounded-full bg-emerald-500/15 text-emerald-700 dark:text-emerald-400">
          <CircleCheckBig className="size-4" />
        </span>
        <p className="text-sm">
          All caught up <span className="text-muted-foreground">— no estimates to confirm.</span>
        </p>
      </Card>
    )
  }
  const stale = entries?.filter((e) => e.stale).length ?? 0
  return (
    <Card className="ring-amber-500/30">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Inbox className="size-4 text-amber-600 dark:text-amber-400" /> To confirm
          {entries && <Badge variant="secondary">{entries.length}</Badge>}
        </CardTitle>
        <CardDescription>
          Estimated amounts — adjust and confirm once the bill arrives.
          {stale > 0 && ` ${stale} from past months.`}
        </CardDescription>
      </CardHeader>
      <CardContent>
        {isPending ? (
          <div className="space-y-3">
            <Skeleton className="h-14 w-full" />
            <Skeleton className="h-14 w-full" />
          </div>
        ) : (
          <ul className="divide-y">
            {entries!.map((e) => (
              <PendingRow key={`${e.id}-${e.amount_cents}`} entry={e} category={byId.get(e.category_id)} onOpen={() => onOpen(e)} />
            ))}
          </ul>
        )}
      </CardContent>
    </Card>
  )
}
