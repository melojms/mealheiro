import { StrictMode } from "react"
import { createRoot } from "react-dom/client"
import "./index.css"
import App from "./App"

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)

// Production only: the service worker makes the app installable and adds an offline page.
if (import.meta.env.PROD && "serviceWorker" in navigator) {
  window.addEventListener("load", () => {
    // A new worker takes over via skipWaiting + clients.claim. Reload once so the page
    // runs under it, but never on the first install and never more than once.
    const hadController = navigator.serviceWorker.controller !== null
    let reloaded = false
    navigator.serviceWorker.addEventListener("controllerchange", () => {
      if (!hadController || reloaded) return
      reloaded = true
      window.location.reload()
    })
    navigator.serviceWorker.register("/sw.js").catch((err) => console.warn("service worker registration failed", err))
  })
}
