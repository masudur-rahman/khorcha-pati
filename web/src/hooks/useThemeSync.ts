import { useEffect, useRef } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'

import { getProfile, updateProfile } from '../api/endpoints'
import { useTheme } from '../context/ThemeContext'
import type { Profile } from '../types'

/* Theme precedence: profile (if set) → localStorage → device scheme.
   The profile value is applied once per session so later refetches never
   fight a toggle the user just made locally. */
export function useServerTheme() {
  const { setTheme } = useTheme()
  const { data: profile } = useQuery({ queryKey: ['profile'], queryFn: getProfile })
  const applied = useRef(false)

  useEffect(() => {
    if (applied.current) return
    if (profile?.theme === 'light' || profile?.theme === 'dark') {
      setTheme(profile.theme)
      applied.current = true
    }
  }, [profile?.theme, setTheme])
}

/* Setter that also persists the choice to the profile (fire-and-forget)
   and keeps the cached profile in sync so refetches agree. */
export function useThemeSetter() {
  const { setTheme } = useTheme()
  const queryClient = useQueryClient()

  return (next: 'light' | 'dark') => {
    setTheme(next)
    queryClient.setQueryData<Profile | undefined>(['profile'], p => (p ? { ...p, theme: next } : p))
    updateProfile({ theme: next }).catch(() => {
      /* theme still applies locally; profile save retries on next change */
    })
  }
}

export function useThemeToggle() {
  const { theme } = useTheme()
  const set = useThemeSetter()
  return () => set(theme === 'dark' ? 'light' : 'dark')
}
