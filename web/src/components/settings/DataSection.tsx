import { useState } from "react"
import { DatabaseBackup, Download, FileSpreadsheet } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Checkbox } from "@/components/ui/checkbox"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Chip } from "@/components/entry/Chip"
import { ENTRY_TYPES, useToday } from "@/components/entry/hooks"
import { ResponsiveDialog } from "@/components/entry/ResponsiveDialog"
import { api } from "@/lib/api"
import type { EntryType } from "@/lib/types"
import { EXPORT_PRESETS, matchPreset, presetRange } from "./presets"

export function DataSection() {
  const [exportOpen, setExportOpen] = useState(false)
  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileSpreadsheet className="size-4" /> Export CSV
          </CardTitle>
          <CardDescription>Entries for a date range, ready for a spreadsheet (semicolon-separated, decimal comma).</CardDescription>
        </CardHeader>
        <CardContent>
          <Button onClick={() => setExportOpen(true)}>
            <Download /> Export…
          </Button>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <DatabaseBackup className="size-4" /> Backup
          </CardTitle>
          <CardDescription>A consistent snapshot of the whole database. Nightly backups also run on the server.</CardDescription>
        </CardHeader>
        <CardContent>
          <Button variant="outline" asChild>
            <a href={api.backupUrl} download>
              <Download /> Download backup
            </a>
          </Button>
        </CardContent>
      </Card>

      <ExportDialog open={exportOpen} onOpenChange={setExportOpen} />
    </div>
  )
}

function ExportDialog({ open, onOpenChange }: { open: boolean; onOpenChange: (o: boolean) => void }) {
  const today = useToday()
  const [range, setRange] = useState(() => presetRange("this_month", today))
  const [types, setTypes] = useState<EntryType[]>(ENTRY_TYPES.map((t) => t.value))
  const preset = matchPreset(range, today)
  const invalid = types.length === 0 || (range.from !== "" && range.to !== "" && range.from > range.to)
  const url = api.exportUrl({ from: range.from || undefined, to: range.to || undefined, types })

  return (
    <ResponsiveDialog
      open={open}
      onOpenChange={onOpenChange}
      title="Export CSV"
      description="Choose a date range and entry types."
      footer={
        <>
          <Button variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
          {invalid ? (
            <Button disabled>
              <Download /> Download
            </Button>
          ) : (
            <Button asChild>
              <a href={url} download onClick={() => onOpenChange(false)}>
                <Download /> Download
              </a>
            </Button>
          )}
        </>
      }
    >
      <div className="space-y-5">
        <div role="group" aria-label="Quick ranges" className="flex flex-wrap gap-2">
          {EXPORT_PRESETS.map((p) => (
            <Chip key={p.value} selected={preset === p.value} onClick={() => setRange(presetRange(p.value, today))}>
              {p.label}
            </Chip>
          ))}
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-1.5">
            <Label htmlFor="export-from">From</Label>
            <Input id="export-from" type="date" value={range.from} onChange={(e) => setRange((r) => ({ ...r, from: e.target.value }))} className="h-9" />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="export-to">To</Label>
            <Input
              id="export-to"
              type="date"
              value={range.to}
              min={range.from || undefined}
              aria-invalid={range.from !== "" && range.to !== "" && range.from > range.to}
              onChange={(e) => setRange((r) => ({ ...r, to: e.target.value }))}
              className="h-9"
            />
          </div>
          <p className="text-muted-foreground col-span-2 -mt-1 text-xs">Leave a date empty for no limit.</p>
        </div>

        <fieldset className="space-y-2.5">
          <legend className="mb-2 text-sm font-medium">Types</legend>
          {ENTRY_TYPES.map((t) => (
            <div key={t.value} className="flex items-center gap-2">
              <Checkbox
                id={`export-${t.value}`}
                checked={types.includes(t.value)}
                onCheckedChange={(c) => setTypes((ts) => (c ? [...ts, t.value] : ts.filter((x) => x !== t.value)))}
              />
              <Label htmlFor={`export-${t.value}`} className="font-normal">{t.label}</Label>
            </div>
          ))}
          {types.length === 0 && <p className="text-destructive text-xs">Select at least one type.</p>}
        </fieldset>
      </div>
    </ResponsiveDialog>
  )
}
