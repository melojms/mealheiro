import { Users } from "lucide-react"
import type { Person } from "@/lib/types"
import { Chip } from "./Chip"

export function PayerChips({
  people, value, onChange,
}: { people: Person[]; value: number | null; onChange: (id: number) => void }) {
  return (
    <div role="group" aria-label="Payer" className="flex flex-wrap gap-2">
      {people.map((p) => (
        <Chip key={p.id} selected={value === p.id} onClick={() => onChange(p.id)}>
          {p.kind === "joint" && <Users />}
          {p.name}
        </Chip>
      ))}
    </div>
  )
}
