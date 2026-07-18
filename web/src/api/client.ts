function getApiBaseUrl(): string {
  return (window as any).__CONFIG__?.API_BASE || import.meta.env.VITE_API_BASE || ''
}

export function getApiBase(): string {
  return getApiBaseUrl()
}

const DEFAULT_BOT_URL = 'https://t.me/KhorchaPatiBot'
const DEFAULT_REPO_URL = 'https://github.com/masudur-rahman/khorcha-pati'

// absoluteUrl ensures a link has a scheme so the browser never resolves it relative to the origin.
function absoluteUrl(url: string): string {
  return /^https?:\/\//i.test(url) ? url : `https://${url.replace(/^\/+/, '')}`
}

// getBotUrl returns the Telegram bot link from runtime config, falling back to the default handle.
export function getBotUrl(): string {
  return absoluteUrl((window as any).__CONFIG__?.BOT_URL || import.meta.env.VITE_BOT_URL || DEFAULT_BOT_URL)
}

// getRepoUrl returns the source repository link from runtime config, falling back to the default repo.
export function getRepoUrl(): string {
  return absoluteUrl((window as any).__CONFIG__?.REPO_URL || import.meta.env.VITE_REPO_URL || DEFAULT_REPO_URL)
}

// getBotHandle derives the bot's @username from the configured bot URL.
export function getBotHandle(): string {
  const match = getBotUrl().match(/t\.me\/([A-Za-z0-9_]+)/i)
  return match ? `@${match[1]}` : ''
}

// getAdminHandle returns the instance admin's Telegram @username (bot owner).
export function getAdminHandle(): string {
  const owner = (window as any).__CONFIG__?.BOT_OWNER
  return owner ? `@${owner}` : ''
}

const RT_KEY = 'expense_rt'

let accessToken = ''
let refreshPromise: Promise<boolean> | null = null

export function setAccessToken(token: string) {
  accessToken = token
}

export function getAccessToken(): string {
  return accessToken
}

export function setRefreshToken(token: string) {
  localStorage.setItem(RT_KEY, token)
}

export function getRefreshToken(): string {
  return localStorage.getItem(RT_KEY) || ''
}

export function clearTokens() {
  accessToken = ''
  localStorage.removeItem(RT_KEY)
}

async function refreshAccessToken(): Promise<boolean> {
  if (refreshPromise) return refreshPromise

  const refreshToken = getRefreshToken()
  refreshPromise = (async () => {
    try {
      const res = await fetch(`${getApiBaseUrl()}/api/v1/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refreshToken }),
        credentials: 'include',
      })
      if (!res.ok) return false
      const data = await res.json()
      setAccessToken(data.accessToken)
      if (data.refreshToken) setRefreshToken(data.refreshToken)
      return true
    } catch {
      return false
    } finally {
      refreshPromise = null
    }
  })()

  return refreshPromise
}

export async function apiFetch<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  }

  if (accessToken) {
    headers['Authorization'] = `Bearer ${accessToken}`
  }

  const url = path.startsWith('http') ? path : `${getApiBaseUrl()}${path}`
  let res = await fetch(url, { ...options, headers, credentials: 'include' })

  if (res.status === 401 && accessToken) {
    const refreshed = await refreshAccessToken()
    if (refreshed) {
      headers['Authorization'] = `Bearer ${accessToken}`
      res = await fetch(url, { ...options, headers, credentials: 'include' })
    } else {
      clearTokens()
      throw new Error('Session expired')
    }
  }

  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: res.statusText }))
    let msg = err.message || res.statusText
    
    // Some backend errors return a JSON string in the message field
    try {
      const parsed = JSON.parse(msg)
      if (parsed.message) {
        msg = parsed.message
      }
    } catch {
      // Not a JSON string, keep original message
    }
    
    throw new Error(msg)
  }

  const contentType = res.headers.get('Content-Type')
  if (contentType?.includes('application/pdf')) {
    return res.blob() as any
  }

  return res.json()
}
