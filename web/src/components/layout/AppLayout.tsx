import { Suspense, useEffect } from "react"
import { NavLink, Outlet, useLocation } from "react-router"
import { ChartColumn, CalendarDays, List, Plus, Settings, Wallet } from "lucide-react"
import { cn } from "@/lib/utils"
import { prefetchPages } from "@/pages/lazy"
import { PageErrorBoundary } from "./PageErrorBoundary"
import { PageSkeleton } from "./PageSkeleton"
import { ThemeToggle } from "./ThemeToggle"

const NAV = [
  { to: "/", label: "Add", icon: Plus, end: true },
  { to: "/month", label: "Month", icon: CalendarDays },
  { to: "/charts", label: "Charts", icon: ChartColumn },
  { to: "/entries", label: "Entries", icon: List },
  { to: "/settings", label: "Settings", icon: Settings },
]

/** Phone: bottom tab bar. Desktop (md+): left sidebar. */
export function AppLayout() {
  useEffect(prefetchPages, [])
  const { pathname } = useLocation()
  return (
    <div className="bg-background text-foreground min-h-svh md:flex">
      <aside className="bg-sidebar sticky top-0 hidden h-svh w-56 shrink-0 flex-col border-r p-3 md:flex">
        <div className="flex items-center gap-2.5 px-2 py-3">
          <span className="flex size-8 items-center justify-center rounded-lg bg-gradient-to-br from-emerald-400 to-teal-600 text-white shadow-sm">
            <Wallet className="size-4.5" />
          </span>
          <span className="text-lg font-semibold tracking-tight">Mealheiro</span>
        </div>
        <nav className="mt-2 flex flex-col gap-1">
          {NAV.map(({ to, label, icon: Icon, end }) => (
            <NavLink
              key={to}
              to={to}
              end={end}
              className={({ isActive }) =>
                cn(
                  "flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors",
                  isActive ? "bg-sidebar-accent text-sidebar-accent-foreground font-medium" : "text-muted-foreground hover:bg-sidebar-accent/60",
                )
              }
            >
              <Icon className="size-4" /> {label}
            </NavLink>
          ))}
        </nav>
        <div className="mt-auto">
          <ThemeToggle />
        </div>
      </aside>

      <main className="mx-auto w-full max-w-6xl px-4 pt-4 pb-24 md:px-8 md:pt-8 md:pb-10">
        <PageErrorBoundary key={pathname}>
          <Suspense fallback={<PageSkeleton />}>
            <Outlet />
          </Suspense>
        </PageErrorBoundary>
      </main>

      <nav className="bg-background/90 fixed inset-x-0 bottom-0 z-40 grid grid-cols-5 border-t pb-[env(safe-area-inset-bottom)] backdrop-blur md:hidden">
        {NAV.map(({ to, label, icon: Icon, end }) => (
          <NavLink
            key={to}
            to={to}
            end={end}
            className={({ isActive }) =>
              cn("flex flex-col items-center gap-0.5 py-2 text-[11px]", isActive ? "text-primary font-medium" : "text-muted-foreground")
            }
          >
            <Icon className="size-5" />
            {label}
          </NavLink>
        ))}
      </nav>
    </div>
  )
}
