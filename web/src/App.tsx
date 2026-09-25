import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { ThemeProvider } from "next-themes"
import { BrowserRouter, Route, Routes } from "react-router"
import { Toaster } from "@/components/ui/sonner"
import { TooltipProvider } from "@/components/ui/tooltip"
import { AppLayout } from "@/components/layout/AppLayout"
import AddPage from "@/pages/AddPage"
import MonthPage from "@/pages/MonthPage"
import ChartsPage from "@/pages/ChartsPage"
import EntriesPage from "@/pages/EntriesPage"
import SettingsPage from "@/pages/SettingsPage"

// Data volume is tiny: after any mutation, call queryClient.invalidateQueries() (no args)
// so every screen refetches.
export const queryClient = new QueryClient({
  defaultOptions: { queries: { staleTime: 30_000, refetchOnWindowFocus: true, retry: 1 } },
})

export default function App() {
  return (
    <ThemeProvider attribute="class" defaultTheme="system" enableSystem storageKey="mm-budget.theme">
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
