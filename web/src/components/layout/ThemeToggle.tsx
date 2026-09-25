import { Monitor, Moon, Sun } from "lucide-react"
import { useTheme } from "next-themes"
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group"

export function ThemeToggle() {
  const { theme, setTheme } = useTheme()
  return (
    <ToggleGroup type="single" variant="outline" size="sm" value={theme} onValueChange={(v) => v && setTheme(v)}>
      <ToggleGroupItem value="light" aria-label="Light theme"><Sun /></ToggleGroupItem>
      <ToggleGroupItem value="dark" aria-label="Dark theme"><Moon /></ToggleGroupItem>
      <ToggleGroupItem value="system" aria-label="System theme"><Monitor /></ToggleGroupItem>
    </ToggleGroup>
  )
}
