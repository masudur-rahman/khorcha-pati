const API_BASE = (window as any).__CONFIG__?.API_BASE || import.meta.env.VITE_API_BASE || ''

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
      const res = await fetch(`${API_BASE}/api/v1/auth/refresh`, {
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

  const url = path.startsWith('http') ? path : `${API_BASE}${path}`
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
