import { useState, useRef, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { ICONS } from '../ui/Icons'
import { useSearch } from '../../context/SearchContext'
import { useAuth } from '../../hooks/useAuth'
import { useTheme } from '../../context/ThemeContext'
import { useBudgetAlerts } from '../../hooks/useBudgets'
import { getProfile } from '../../api/endpoints'
import { fmt } from '../../lib/formatter'
import SearchResults from './SearchResults'

interface TopBarProps {
  title: string
  subtitle?: string
}

export default function TopBar({ title, subtitle }: TopBarProps) {
  const { searchTerm, setSearchTerm } = useSearch()
  const { logout, isAuthenticated } = useAuth()
  const { data: alerts } = useBudgetAlerts()
  const { data: profile } = useQuery({ queryKey: ['profile'], queryFn: getProfile, enabled: isAuthenticated })
  const fullName = profile ? `${profile.firstName} ${profile.lastName}`.trim() : ''
  const username = profile?.username ? `@${profile.username}` : ''
  const [isSearchExpanded, setIsSearchExpanded] = useState(false)
  const [showProfile, setShowProfile] = useState(false)
  const [showNotifications, setShowNotifications] = useState(false)
  const profileRef = useRef<HTMLDivElement>(null)
  const notifRef = useRef<HTMLDivElement>(null)
  const notifBtnRef = useRef<HTMLButtonElement>(null)
  const profileBtnRef = useRef<HTMLButtonElement>(null)
  const searchRef = useRef<HTMLDivElement>(null)
  const searchInputRef = useRef<HTMLInputElement>(null)
  const [dropdownTop, setDropdownTop] = useState(0)
  const [searchResultsTop, setSearchResultsTop] = useState(0)

  const { theme, toggle: toggleDarkMode } = useTheme()
  const darkMode = theme === 'dark'

  // Close dropdowns on outside click
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (profileRef.current && !profileRef.current.contains(e.target as Node)) setShowProfile(false)
      if (notifRef.current && !notifRef.current.contains(e.target as Node)) setShowNotifications(false)
      if (searchRef.current && !searchRef.current.contains(e.target as Node)) {
        setIsSearchExpanded(false)
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  // Esc closes search panel
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && searchTerm) {
        setSearchTerm('')
        searchInputRef.current?.blur()
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [searchTerm, setSearchTerm])

  // Track search input bottom for results panel anchor
  useEffect(() => {
    if (!isSearchExpanded || !searchRef.current) return
    setSearchResultsTop(searchRef.current.getBoundingClientRect().bottom + 8)
  }, [isSearchExpanded, searchTerm])

  const iconBtnStyle: React.CSSProperties = {
    width: 42,
    height: 42,
    borderRadius: 'var(--radius-md)',
    border: '1px solid var(--color-border)',
    background: 'var(--color-surface)',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    cursor: 'pointer',
    color: 'var(--color-text-secondary)',
    position: 'relative',
    transition: 'all var(--transition-fast)',
  }

  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '20px 0', marginBottom: 4, gap: 12, flexWrap: 'wrap' }}>
      <div style={{ minWidth: 0 }}>
        <h1 style={{ fontSize: 'clamp(20px, 4vw, 28px)', fontWeight: 700, color: 'var(--color-text-primary)', letterSpacing: '-0.025em', margin: 0 }}>
          {title}
        </h1>
        {subtitle && (
          <p style={{ fontSize: 14, color: 'var(--color-text-tertiary)', marginTop: 4, fontWeight: 500 }}>{subtitle}</p>
        )}
      </div>

      <div style={{ display: 'flex', alignItems: 'center', gap: 10, position: 'relative' }}>
        {/* Search */}
        <div ref={searchRef}>
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              background: 'var(--color-surface)',
              border: '1px solid var(--color-border)',
              borderRadius: 'var(--radius-md)',
              padding: '0 12px',
              height: 42,
              width: isSearchExpanded ? 260 : 42,
              transition: 'width 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
              overflow: 'hidden',
              cursor: isSearchExpanded ? 'default' : 'pointer',
            }}
            onClick={() => !isSearchExpanded && setIsSearchExpanded(true)}
          >
            <span style={{ color: 'var(--color-text-tertiary)', display: 'flex', flexShrink: 0 }}>
              {ICONS.search(18)}
            </span>
            {isSearchExpanded && (
              <>
                <input
                  ref={searchInputRef}
                  autoFocus
                  value={searchTerm}
                  onChange={e => setSearchTerm(e.target.value)}
                  placeholder="Search anything..."
                  style={{
                    background: 'transparent',
                    border: 'none',
                    outline: 'none',
                    fontSize: 13,
                    color: 'var(--color-text-primary)',
                    marginLeft: 10,
                    width: '100%',
                    fontFamily: 'inherit',
                  }}
                />
                {searchTerm && (
                  <button
                    onClick={() => { setSearchTerm(''); searchInputRef.current?.focus() }}
                    aria-label="Clear search"
                    style={{
                      width: 22, height: 22, borderRadius: '50%', border: 'none',
                      background: 'var(--color-bg)', color: 'var(--color-text-tertiary)',
                      cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center',
                      flexShrink: 0,
                    }}
                  >
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                  </button>
                )}
              </>
            )}
          </div>
          {isSearchExpanded && searchTerm && (
            <SearchResults anchorTop={searchResultsTop} onClose={() => setIsSearchExpanded(false)} />
          )}
        </div>

        {/* Notifications */}
        <div ref={notifRef}>
          <button
            ref={notifBtnRef}
            style={iconBtnStyle}
            onClick={() => {
              if (notifBtnRef.current) setDropdownTop(notifBtnRef.current.getBoundingClientRect().bottom + 8)
              setShowNotifications(!showNotifications); setShowProfile(false)
            }}
          >
            {ICONS.bell(18)}
            {alerts && alerts.length > 0 && (
              <span style={{
                position: 'absolute', top: 8, right: 8, width: 8, height: 8, borderRadius: '50%',
                background: 'var(--color-danger)', border: '2px solid var(--color-surface)',
              }} />
            )}
          </button>
          {showNotifications && (
            <div style={{
              position: 'fixed', top: dropdownTop, right: 16,
              width: 'min(340px, calc(100vw - 32px))',
              background: 'var(--color-surface)', borderRadius: 'var(--radius-lg)',
              border: '1px solid var(--color-border)',
              boxShadow: '0 16px 48px rgba(0,0,0,0.12)',
              zIndex: 200, overflow: 'hidden',
            }}>
              <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--color-border)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <h3 style={{ fontSize: 14, fontWeight: 700, color: 'var(--color-text-primary)', margin: 0 }}>Notifications</h3>
                {alerts && alerts.length > 0 && (
                  <span style={{
                    fontSize: 11, fontWeight: 700, color: 'var(--color-danger)',
                    background: 'var(--color-danger-subtle)', borderRadius: 10,
                    padding: '2px 8px',
                  }}>{alerts.length}</span>
                )}
              </div>
              <div style={{ maxHeight: 300, overflowY: 'auto' }}>
                {alerts && alerts.length > 0 ? alerts.map(alert => (
                  <div key={alert.categoryId} style={{
                    padding: '14px 20px',
                    borderBottom: '1px solid var(--color-border)',
                    display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                    gap: 12,
                  }}>
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <p style={{ fontSize: 13, fontWeight: 600, color: 'var(--color-text-primary)', margin: 0 }}>
                        {alert.categoryName}
                      </p>
                      <p style={{ fontSize: 12, color: 'var(--color-text-tertiary)', margin: '2px 0 0' }}>
                        {alert.percent >= 100 ? 'Budget exceeded' : 'Approaching budget limit'} — {Math.round(alert.percent)}% used
                      </p>
                    </div>
                    <span style={{
                      fontSize: 12, fontWeight: 700, whiteSpace: 'nowrap',
                      color: alert.percent >= 100 ? 'var(--color-danger)' : 'var(--color-warning)',
                    }}>
                      {fmt(alert.spent, 0)} / {fmt(alert.budgetAmount, 0)}
                    </span>
                  </div>
                )) : (
                  <div style={{ padding: 24, textAlign: 'center' }}>
                    <p style={{ fontSize: 13, color: 'var(--color-text-tertiary)', margin: 0 }}>No budget warnings</p>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>

        {/* Profile */}
        <div ref={profileRef}>
          <button
            style={{
              width: 42,
              height: 42,
              borderRadius: 'var(--radius-md)',
              background: 'var(--color-primary)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: 'white',
              fontWeight: 700,
              fontSize: 14,
              cursor: 'pointer',
              border: 'none',
              transition: 'all var(--transition-fast)',
            }}
            ref={profileBtnRef}
            onClick={() => {
              if (profileBtnRef.current) setDropdownTop(profileBtnRef.current.getBoundingClientRect().bottom + 8)
              setShowProfile(!showProfile); setShowNotifications(false)
            }}
            aria-label="Profile menu"
          >
            <UserAvatarIcon size={22} />
          </button>
          {showProfile && (
            <div style={{
              position: 'fixed', top: dropdownTop, right: 16,
              width: 'min(260px, calc(100vw - 32px))',
              background: 'var(--color-surface)', borderRadius: 'var(--radius-lg)',
              border: '1px solid var(--color-border)',
              boxShadow: '0 16px 48px rgba(0,0,0,0.12)',
              zIndex: 200, overflow: 'hidden',
            }}>
              {/* Profile info */}
              <div style={{ padding: '20px', borderBottom: '1px solid var(--color-border)' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                  <div style={{
                    width: 40, height: 40, borderRadius: 'var(--radius-md)',
                    background: 'var(--color-primary)', color: 'white',
                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                  }}>
                    <UserAvatarIcon size={22} />
                  </div>
                  <div style={{ minWidth: 0 }}>
                    <p style={{ fontSize: 14, fontWeight: 700, color: 'var(--color-text-primary)', margin: 0, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                      {fullName || 'Profile'}
                    </p>
                    {username && (
                      <p style={{ fontSize: 12, color: 'var(--color-text-tertiary)', margin: 0 }}>{username}</p>
                    )}
                  </div>
                </div>
              </div>

              {/* Menu items */}
              <div style={{ padding: '8px' }}>
                <Link to="/settings" onClick={() => setShowProfile(false)} style={menuItemStyle}>
                  <span style={{ display: 'flex', color: 'var(--color-text-tertiary)' }}>{ICONS.user(16)}</span>
                  Profile
                </Link>

                {/* Dark mode toggle */}
                <div
                  style={{ ...menuItemStyle, cursor: 'pointer', justifyContent: 'space-between' } as any}
                  onClick={toggleDarkMode}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                    <span style={{ display: 'flex', color: 'var(--color-text-tertiary)' }}>
                      <svg width={16} height={16} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                        {darkMode
                          ? <><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></>
                          : <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
                        }
                      </svg>
                    </span>
                    Dark Mode
                  </div>
                  {/* Toggle switch */}
                  <div style={{
                    width: 36, height: 20, borderRadius: 10,
                    background: darkMode ? 'var(--color-primary)' : 'var(--color-border)',
                    padding: 2, transition: 'all var(--transition-fast)', cursor: 'pointer',
                  }}>
                    <div style={{
                      width: 16, height: 16, borderRadius: '50%', background: 'white',
                      transition: 'transform var(--transition-fast)',
                      transform: darkMode ? 'translateX(16px)' : 'translateX(0)',
                      boxShadow: '0 1px 3px rgba(0,0,0,0.15)',
                    }} />
                  </div>
                </div>
              </div>

              {/* Logout */}
              <div style={{ padding: '8px', borderTop: '1px solid var(--color-border)' }}>
                <div
                  style={{ ...menuItemStyle, color: 'var(--color-danger)', cursor: 'pointer' } as any}
                  onClick={() => { setShowProfile(false); logout() }}
                >
                  <span style={{ display: 'flex' }}>{ICONS.logout(16)}</span>
                  Sign Out
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

const menuItemStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: 10,
  padding: '10px 12px',
  borderRadius: 'var(--radius-sm)',
  fontSize: 13,
  fontWeight: 500,
  color: 'var(--color-text-primary)',
  textDecoration: 'none',
  transition: 'background var(--transition-fast)',
}

function UserAvatarIcon({ size = 22 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden>
      <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
      <circle cx="12" cy="7" r="4" />
    </svg>
  )
}
