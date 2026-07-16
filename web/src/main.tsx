import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App'
import { getApiBase } from './api/client'

/* Hydrate runtime config from the backend before first render so the bot
   identity (links, @handle) is always the deployed bot, not the compiled-in
   default. An explicit BOT_URL in config.js still wins. Failure or slow
   network (>1.5s) falls back to the static defaults. */
async function hydrateRuntimeConfig() {
  const cfg = ((window as any).__CONFIG__ ??= {})
  if (cfg.BOT_URL) return
  try {
    const ctrl = new AbortController()
    const timer = setTimeout(() => ctrl.abort(), 1500)
    const res = await fetch(`${getApiBase()}/api/v1/meta`, { signal: ctrl.signal })
    clearTimeout(timer)
    if (res.ok) {
      const meta = await res.json()
      if (meta.botUsername) cfg.BOT_URL = `https://t.me/${meta.botUsername}`
    }
  } catch {
    /* static defaults apply */
  }
}

hydrateRuntimeConfig().finally(() => {
  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <App />
    </StrictMode>,
  )
})
