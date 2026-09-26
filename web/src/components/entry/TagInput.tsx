import { useDeferredValue, useId, useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { Hash, X } from "lucide-react"
import { api } from "@/lib/api"
import { cn } from "@/lib/utils"
import { addTags, normalizeTag } from "./tags"

/** Free-form tag chips with autocomplete from existing tags. Enter or comma commits a tag. */
export function TagInput({ value, onChange, id }: { value: string[]; onChange: (tags: string[]) => void; id?: string }) {
  const [text, setText] = useState("")
  const [focused, setFocused] = useState(false)
  const listId = useId()
  const q = useDeferredValue(normalizeTag(text))
  const { data } = useQuery({ queryKey: ["tags", q], queryFn: () => api.tags(q), enabled: focused, staleTime: 60_000 })
  const suggestions = (data ?? []).filter((t) => !value.includes(t.name)).slice(0, 8)

  const commit = (raw: string) => {
    onChange(addTags(value, raw))
    setText("")
  }

  return (
    <div className="space-y-2">
      <div
        className={cn(
          "border-input dark:bg-input/30 flex min-h-9 flex-wrap items-center gap-1.5 rounded-lg border px-2 py-1.5 transition-colors",
          "focus-within:border-ring focus-within:ring-ring/50 focus-within:ring-3",
        )}
      >
        {value.map((t) => (
          <span key={t} className="bg-secondary text-secondary-foreground inline-flex h-6 items-center gap-1 rounded-full pr-1 pl-2 text-xs">
            #{t}
            <button
              type="button"
              aria-label={`Remove tag ${t}`}
              className="hover:bg-foreground/10 rounded-full p-0.5"
              onClick={() => onChange(value.filter((x) => x !== t))}
            >
              <X className="size-3" />
            </button>
          </span>
        ))}
        <input
          id={id}
          value={text}
          role="combobox"
          aria-expanded={focused && suggestions.length > 0}
          aria-controls={listId}
          autoComplete="off"
          autoCapitalize="none"
          enterKeyHint="done"
          placeholder={value.length ? "" : "Add tags…"}
          className="placeholder:text-muted-foreground min-w-24 flex-1 bg-transparent text-sm outline-none"
          onChange={(e) => {
            const v = e.target.value
            if (v.includes(",")) commit(v)
            else setText(v)
          }}
          onFocus={() => setFocused(true)}
          onBlur={() => {
            setFocused(false)
            if (text.trim()) commit(text)
          }}
          onKeyDown={(e) => {
            if (e.key === "Enter") {
              e.preventDefault()
              if (text.trim()) commit(text)
            } else if (e.key === "Backspace" && text === "" && value.length) {
              onChange(value.slice(0, -1))
            }
          }}
        />
      </div>
      {focused && suggestions.length > 0 && (
        <div id={listId} role="listbox" aria-label="Tag suggestions" className="flex flex-wrap gap-1.5">
          {suggestions.map((t) => (
            <button
              key={t.name}
              type="button"
              role="option"
              aria-selected={false}
              // mousedown keeps focus on the input so the list doesn't close before the click lands
              onMouseDown={(e) => e.preventDefault()}
              onClick={() => commit(t.name)}
              className="text-muted-foreground hover:bg-muted hover:text-foreground inline-flex h-7 items-center gap-1 rounded-full border px-2.5 text-xs"
            >
              <Hash className="size-3" />
              {t.name}
              <span className="text-muted-foreground/70 tabular-nums">{t.count}</span>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
