import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { Check, Loader2, User, Users } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import { Chip } from "@/components/entry/Chip"
import { usePeople } from "@/components/entry/hooks"
import { api } from "@/lib/api"
import { setDefaultPayer, useDefaultPayer } from "@/lib/device"
import type { Person } from "@/lib/types"

const MAX_NAME = 40

export function PeopleSection() {
  const { data: people, isLoading } = usePeople()
  const defaultPayer = useDefaultPayer()

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>Names</CardTitle>
          <CardDescription>Renaming updates every past entry too.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {isLoading
            ? Array.from({ length: 3 }, (_, i) => <Skeleton key={i} className="h-9" />)
            : people?.map((p) => <PersonRow key={`${p.id}-${p.name}`} person={p} />)}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>This device</CardTitle>
          <CardDescription>Preselected as payer on the Add screen. Stored only on this device.</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-wrap gap-2">
          {people?.map((p) => (
            <Chip key={p.id} selected={defaultPayer === p.id} onClick={() => setDefaultPayer(p.id)}>
              {p.kind === "joint" && <Users />}
              {p.name}
            </Chip>
          ))}
          <Chip selected={defaultPayer === null} onClick={() => setDefaultPayer(null)}>
            Ask me
          </Chip>
        </CardContent>
      </Card>
    </div>
  )
}

function PersonRow({ person }: { person: Person }) {
  const queryClient = useQueryClient()
  const [name, setName] = useState(person.name)
  const trimmed = name.trim()
  const dirty = trimmed !== person.name
  const valid = trimmed.length > 0 && trimmed.length <= MAX_NAME

  const rename = useMutation({
    mutationFn: () => api.renamePerson(person.id, trimmed),
    onSuccess: (p) => {
      queryClient.invalidateQueries()
      toast.success(`Renamed to ${p.name}`)
    },
    onError: (e) => toast.error(e.message),
  })

  return (
    <form
      className="flex items-center gap-2"
      onSubmit={(e) => {
        e.preventDefault()
        if (dirty && valid) rename.mutate()
      }}
    >
      <span className="bg-muted text-muted-foreground flex size-9 shrink-0 items-center justify-center rounded-full">
        {person.kind === "joint" ? <Users className="size-4" /> : <User className="size-4" />}
      </span>
      <Input
        aria-label={`Name of ${person.name}`}
        value={name}
        maxLength={MAX_NAME}
        aria-invalid={!valid}
        onChange={(e) => setName(e.target.value)}
        onBlur={() => !valid && setName(person.name)}
        className="h-9"
      />
      <Button type="submit" size="icon-lg" variant={dirty ? "default" : "ghost"} disabled={!dirty || !valid || rename.isPending} aria-label="Save name">
        {rename.isPending ? <Loader2 className="animate-spin" /> : <Check />}
      </Button>
    </form>
  )
}
