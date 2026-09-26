import { Component, type ReactNode } from "react"
import { ErrorState } from "@/components/filters/states"

/** Keeps the navigation usable when a screen throws; resets when the route changes (via `key`). */
export class PageErrorBoundary extends Component<{ children: ReactNode }, { failed: boolean }> {
  state = { failed: false }

  static getDerivedStateFromError() {
    return { failed: true }
  }

  componentDidCatch(error: unknown) {
    console.error(error)
  }

  render() {
    if (this.state.failed) return <ErrorState className="py-20" onRetry={() => window.location.reload()} />
    return this.props.children
  }
}
