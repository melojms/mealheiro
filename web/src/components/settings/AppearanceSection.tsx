import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { ThemeToggle } from "@/components/layout/ThemeToggle"
import { useMeta } from "@/components/entry/hooks"

export function AppearanceSection() {
  const { data: meta, isLoading } = useMeta()
  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>Theme</CardTitle>
          <CardDescription>Light, dark, or follow the system setting.</CardDescription>
        </CardHeader>
        <CardContent>
          <ThemeToggle />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>About</CardTitle>
        </CardHeader>
        <CardContent>
          <dl className="grid grid-cols-[auto_1fr] gap-x-6 gap-y-2 text-sm">
            <dt className="text-muted-foreground">Version</dt>
            <dd className="font-mono">{isLoading ? <Skeleton className="h-4 w-20" /> : (meta?.version ?? "—")}</dd>
            <dt className="text-muted-foreground">Time zone</dt>
            <dd>{isLoading ? <Skeleton className="h-4 w-28" /> : (meta?.timezone ?? "—")}</dd>
          </dl>
        </CardContent>
      </Card>
    </div>
  )
}
