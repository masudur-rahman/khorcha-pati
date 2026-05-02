import { createContext, useState, useCallback, useEffect, type ReactNode } from 'react'
import {
  setAccessToken, clearTokens, getAccessToken,
  setRefreshToken, getRefreshToken, getApiBase,
} from '../api/client'
import { logout as apiLogout } from '../api/endpoints'

interface AuthContextValue {
  isAuthenticated: boolean
  isLoading: boolean
  login: (accessToken: string, refreshToken?: string) => void
  logout: () => void
}

export const AuthContext = createContext<AuthContextValue>({
  isAuthenticated: false,
  isLoading: true,
  login: () => {},
  logout: () => {},
})

export function AuthProvider({ children }: { children: ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(() => !!getAccessToken())
  const [isLoading, setIsLoading] = useState(() => !getAccessToken() && !!getRefreshToken())

  const login = useCallback((accessToken: string, refreshToken?: string) => {
    setAccessToken(accessToken)
    if (refreshToken) setRefreshToken(refreshToken)
    setIsAuthenticated(true)
  }, [])

  const logout = useCallback(async () => {
    try { await apiLogout() } catch { /* ignore */ }
    clearTokens()
    setIsAuthenticated(false)
  }, [])

  useEffect(() => {
    if (isAuthenticated) {
      setIsLoading(false)
      return
    }
    const rt = getRefreshToken()
    if (!rt) {
      setIsLoading(false)
      return
    }

    fetch(`${getApiBase()}/api/v1/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refreshToken: rt }),
      credentials: 'include',
    })
      .then(res => res.ok ? res.json() : null)
      .then(data => {
        if (data?.accessToken) {
          setAccessToken(data.accessToken)
          if (data.refreshToken) setRefreshToken(data.refreshToken)
          setIsAuthenticated(true)
        } else {
          clearTokens()
        }
      })
      .catch(() => { clearTokens() })
      .finally(() => { setIsLoading(false) })
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <AuthContext.Provider value={{ isAuthenticated, isLoading, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}
