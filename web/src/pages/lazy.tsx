import { lazy, type ComponentType } from "react"

type Loader = () => Promise<{ default: ComponentType }>

const loaders: Loader[] = []

const RELOAD_KEY = "mm-budget.chunk-reload"

/**
 * React.lazy that also registers the chunk for idle prefetching (see prefetchPages).
 * A tab left open across an upgrade asks for chunks that no longer exist: reload once to pick up the new build.
 */
export function lazyPage(load: Loader) {
  loaders.push(load)
  return lazy(() =>
    load().then(
      (mod) => {
        sessionStorage.removeItem(RELOAD_KEY)
        return mod
      },
      (err) => {
        if (!sessionStorage.getItem(RELOAD_KEY)) {
          sessionStorage.setItem(RELOAD_KEY, "1")
          window.location.reload()
          return new Promise<never>(() => {})
        }
        throw err
      },
    ),
  )
}

/** Warms the other screens' chunks once the landing page is idle, so tab switches don't wait on the network. */
export function prefetchPages() {
  const run = () => loaders.forEach((load) => void load().catch(() => {}))
  if ("requestIdleCallback" in window) window.requestIdleCallback(run, { timeout: 3000 })
  else setTimeout(run, 1500)
}
