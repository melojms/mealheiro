import { useState } from "react"
import { CalendarDays } from "lucide-react"
import { Calendar } from "@/components/ui/calendar"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { addDays, dateChip, formatDay, fromISODate, toISODate } from "@/components/add/dates"
import { Chip } from "./Chip"

/** Today · Yesterday · calendar pick. `value` and `today` are YYYY-MM-DD. */
export function DateChips({ value, today, onChange }: { value: string; today: string; onChange: (iso: string) => void }) {
  const [open, setOpen] = useState(false)
  const chip = dateChip(value, today)
  return (
    <div role="group" aria-label="Date" className="flex flex-wrap gap-2">
      <Chip selected={chip === "today"} onClick={() => onChange(today)}>
        Today
      </Chip>
      <Chip selected={chip === "yesterday"} onClick={() => onChange(addDays(today, -1))}>
        Yesterday
      </Chip>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Chip selected={chip === "other"} aria-label="Pick a date">
            <CalendarDays />
            {chip === "other" ? formatDay(value, today) : "Pick"}
          </Chip>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0" align="start">
          <Calendar
            mode="single"
            required
            selected={fromISODate(value)}
            defaultMonth={fromISODate(value)}
            weekStartsOn={1}
            onSelect={(d) => {
              if (!d) return
              onChange(toISODate(d))
              setOpen(false)
            }}
          />
        </PopoverContent>
      </Popover>
    </div>
  )
}
