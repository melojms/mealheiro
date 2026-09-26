import type { ReactNode } from "react"
import { Link, NavLink, Route, Routes, useMatch } from "react-router"
import { ChevronLeft, ChevronRight, Database, Palette, PiggyBank, Repeat, Shapes, Users, type LucideIcon } from "lucide-react"
import { PageHeader } from "@/components/PageHeader"
import { AppearanceSection } from "@/components/settings/AppearanceSection"
import { BudgetsSection } from "@/components/settings/BudgetsSection"
import { CategoriesSection } from "@/components/settings/CategoriesSection"
import { DataSection } from "@/components/settings/DataSection"
import { PeopleSection } from "@/components/settings/PeopleSection"
import { TemplatesSection } from "@/components/settings/TemplatesSection"
import { cn } from "@/lib/utils"

interface Section {
  path: string
  label: string
  description: string
  icon: LucideIcon
  element: ReactNode
}

const SECTIONS: Section[] = [
  { path: "people", label: "People", description: "Names and this device's payer", icon: Users, element: <PeopleSection /> },
  { path: "categories", label: "Categories", description: "Icons, colors, subcategories", icon: Shapes, element: <CategoriesSection /> },
  { path: "recurring", label: "Recurring", description: "Monthly templates", icon: Repeat, element: <TemplatesSection /> },
  { path: "budgets", label: "Budgets", description: "Monthly limits per category", icon: PiggyBank, element: <BudgetsSection /> },
  { path: "data", label: "Data", description: "CSV export and backups", icon: Database, element: <DataSection /> },
  { path: "appearance", label: "Appearance", description: "Theme and app version", icon: Palette, element: <AppearanceSection /> },
]

/**
 * Phone: /settings is a list of sections, each opening as its own page.
 * Desktop: section nav on the left, content on the right (/settings shows People).
 */
export default function SettingsPage() {
  const isIndex = useMatch("/settings") !== null

  return (
    <div className="mx-auto max-w-5xl">
      <div className={cn(!isIndex && "hidden md:block")}>
        <PageHeader title="Settings" />
      </div>
      <div className="md:grid md:grid-cols-[13rem_minmax(0,1fr)] md:gap-8">
        <nav aria-label="Settings sections" className={cn("md:block", !isIndex && "hidden")}>
          <ul className="bg-card divide-y overflow-hidden rounded-xl border md:divide-y-0 md:border-0 md:bg-transparent">
            {SECTIONS.map((s) => (
              <li key={s.path}>
                <NavLink
                  to={`/settings/${s.path}`}
                  className={({ isActive }) =>
                    cn(
                      "flex items-center gap-3 px-4 py-3 transition-colors md:rounded-lg md:px-3 md:py-2",
                      "hover:bg-muted/60",
                      (isActive || (isIndex && s.path === "people")) && "md:bg-muted md:font-medium",
                    )
                  }
                >
                  <span className="bg-muted text-muted-foreground flex size-8 items-center justify-center rounded-lg md:size-auto md:bg-transparent">
                    <s.icon className="size-4" />
                  </span>
                  <span className="min-w-0 flex-1">
                    <span className="block text-sm">{s.label}</span>
                    <span className="text-muted-foreground block truncate text-xs md:hidden">{s.description}</span>
                  </span>
                  <ChevronRight className="text-muted-foreground size-4 md:hidden" />
                </NavLink>
              </li>
            ))}
          </ul>
        </nav>

        <div className={cn(isIndex && "hidden md:block")}>
          <Routes>
            <Route index element={<SectionPage section={SECTIONS[0]} />} />
            {SECTIONS.map((s) => (
              <Route key={s.path} path={s.path} element={<SectionPage section={s} />} />
            ))}
            <Route path="*" element={<p className="text-muted-foreground text-sm">Unknown section.</p>} />
          </Routes>
        </div>
      </div>
    </div>
  )
}

function SectionPage({ section }: { section: Section }) {
  return (
    <div>
      <Link to="/settings" className="text-muted-foreground hover:text-foreground mb-2 inline-flex items-center gap-1 text-sm md:hidden">
        <ChevronLeft className="size-4" /> Settings
      </Link>
      <div className="mb-5">
        <h2 className="text-2xl font-semibold tracking-tight md:text-xl">{section.label}</h2>
        <p className="text-muted-foreground mt-1 text-sm">{section.description}</p>
      </div>
      {section.element}
    </div>
  )
}
