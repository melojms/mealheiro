import { Smartphone, Users } from "lucide-react"
import { Button } from "@/components/ui/button"
import { setDefaultPayer } from "@/lib/device"
import type { Person } from "@/lib/types"

/** Asked once per device when no default payer is stored yet. */
export function DevicePayerPrompt({ people }: { people: Person[] }) {
  return (
    <div className="bg-muted/40 animate-in fade-in rounded-xl border p-3">
      <p className="mb-2.5 flex items-center gap-2 text-sm font-medium">
        <Smartphone className="text-muted-foreground size-4" /> Who's using this device?
      </p>
      <div className="flex flex-wrap gap-2">
        {people.map((p) => (
          <Button key={p.id} variant="outline" size="sm" className="rounded-full" onClick={() => setDefaultPayer(p.id)}>
            {p.kind === "joint" && <Users />}
            {p.name}
          </Button>
        ))}
      </div>
      <p className="text-muted-foreground mt-2 text-xs">It will be preselected as payer here. Change it anytime in Settings.</p>
    </div>
  )
}
