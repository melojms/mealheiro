import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { ThemeProvider } from "next-themes"
import { BrowserRouter, Route, Routes } from "react-router"
import { Toaster } from "@/components/ui/sonner"
import { TooltipProvider } from "@/components/ui/tooltip"
import { AppLayout } from "@/components/layout/AppLayout"
import AddPage from "@/pages/AddPage"
import { lazyPage } from "@/pages/lazy"

// Add is the landing screen and ships in the main bundle; the rest (charts, dialogs) load on demand.
const MonthPage = lazyPage(() => import("@/pages/MonthPage"))
const ChartsPage = lazyPage(() => import("@/pages/ChartsPage"))
const EntriesPage = lazyPage(() => import("@/pages/EntriesPage"))
const SettingsPage = lazyPage(() => import("@/pages/SettingsPage"))

// Data volume is tiny: after any mutation, call queryClient.invalidateQueries() (no args)
// so every screen refetches.
export const queryClient = new QueryClient({
  defaultOptions: { queries: { staleTime: 30_000, refetchOnWindowFocus: true, retry: 1 } },
})

export default function App() {
  return (
    <ThemeProvider attribute="class" defaultTheme="system" enableSystem storageKey="mealheiro.theme">
      <QueryClientProvider client={queryClient}>
        <TooltipProvider>
          <BrowserRouter>
            <Routes>
              <Route element={<AppLayout />}>
                <Route index element={<AddPage />} />
                <Route path="month" element={<MonthPage />} />
                <Route path="charts" element={<ChartsPage />} />
                <Route path="entries" element={<EntriesPage />} />
                <Route path="settings/*" element={<SettingsPage />} />
              </Route>
            </Routes>
          </BrowserRouter>
          <Toaster position="top-center" richColors />
        </TooltipProvider>
      </QueryClientProvider>
    </ThemeProvider>
  )
}
