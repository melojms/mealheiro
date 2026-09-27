import { UserRound } from "lucide-react"
import { Chip } from "./Chip"

/**
 * Toggle that keeps an expense out of the shared split. Render only when canBePersonal.
 * Sits after the payer chips, with a divider so it doesn't read as another payer.
 */
export function PersonalChip({ value, onChange }: { value: boolean; onChange: (v: boolean) => void }) {
  return (
    <>
      <span aria-hidden className="bg-border h-5 w-px" />
      <Chip selected={value} onClick={() => onChange(!value)} title="Personal expenses are not split between you">
        <UserRound />
        Personal
      </Chip>
    </>
  )
}
