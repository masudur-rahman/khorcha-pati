import { useEffect, useMemo, useState } from 'react'
import { NavLink, useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '../../hooks/useAuth'
import { useTheme } from '../../context/ThemeContext'
import { ICONS } from '../ui/Icons'

const primary = [
  { to: '/', icon: ICONS.dashboard, label: 'Home', match: (p: string) => p === '/' },
  { to: '/transactions', icon: ICONS.transactions, label: 'Txns', match: (p: string) => p.startsWith('/transactions') },
  { to: '/wallets', icon: ICONS.wallet, label: 'Wallets', match: (p: string) => p.startsWith('/wallets') },
  { to: '/contacts', icon: ICONS.users, label: 'Contacts', match: (p: string) => p.startsWith('/contacts') },
  { to: '/budgets', icon: ICONS.budget, label: 'Budgets', match: (p: string) => p.startsWith('/budgets') },
]

const moreRoutes = ['/settings', '/admin']

export default function MobileNav() {
  const { isAdmin, logout } = useAuth()
  const { theme, toggle } = useTheme()
  const navigate = useNavigate()
  const location = useLocation()
  const [showMore, setShowMore] = useState(false)
  const [startY, setStartY] = useState(0)
  const [currentY, setCurrentY] = useState(0)

  const moreActive = useMemo(() => moreRoutes.some(r => location.pathname.startsWith(r)), [location.pathname])

  useEffect(() => {
    setShowMore(false)
  }, [location.pathname])

  useEffect(() => {
    if (!showMore) return
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') setShowMore(false) }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [showMore])

  const tabStyle = (active: boolean): React.CSSProperties => ({
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    gap: 2,
    padding: '8px 12px',
    borderRadius: 'var(--radius-sm)',
    textDecoration: 'none',
    color: active ? 'var(--color-primary)' : 'var(--color-text-tertiary)',
    fontSize: 10,
    fontWeight: active ? 700 : 500,
    background: 'none',
    border: 'none',
    cursor: 'pointer',
    fontFamily: 'inherit',
    transition: 'color var(--transition-fast)',
    minWidth: 56,
  })

  return (
    <>
      <nav
        style={{
          position: 'fixed',
          bottom: 0,
          left: 0,
          right: 0,
          background: 'var(--color-surface)',
          borderTop: '1px solid var(--color-border)',
          zIndex: 95,
          padding: '6px 8px env(safe-area-inset-bottom, 8px)',
        }}
        className="flex! md:hidden!"
      >
        <div style={{ display: 'flex', justifyContent: 'space-around', width: '100%' }}>
          {primary.map(item => (
            <NavLink key={item.to} to={item.to} style={({ isActive }) => tabStyle(isActive)}>
              {({ isActive }) => (
                <>
                  <span style={{ opacity: isActive ? 1 : 0.6 }}>{item.icon(22)}</span>
                  <span>{item.label}</span>
                </>
              )}
            </NavLink>
          ))}
          <button
            onClick={() => setShowMore(s => !s)}
            style={tabStyle(moreActive || showMore)}
            aria-label="More menu"
          >
            <span style={{ opacity: moreActive || showMore ? 1 : 0.6 }}>{ICONS.moreHorizontal(22)}</span>
            <span>More</span>
          </button>
        </div>
      </nav>

      {showMore && (
        <div
          className="md:hidden"
          style={{
            position: 'fixed', inset: 0, background: 'rgba(15, 23, 42, 0.45)',
            backdropFilter: 'blur(2px)', zIndex: 220, display: 'flex', alignItems: 'flex-end',
          }}
          onClick={() => setShowMore(false)}
        >
          <div
            onClick={e => e.stopPropagation()}
            onTouchStart={e => setStartY(e.touches[0].clientY)}
            onTouchMove={e => {
              if (startY === 0) return
              const y = e.touches[0].clientY
              if (y > startY) {
                setCurrentY(y - startY)
              }
            }}
            onTouchEnd={() => {
              if (currentY > 50) {
                setShowMore(false)
              }
              setStartY(0)
              setCurrentY(0)
            }}
            style={{
              width: '100%', background: 'var(--color-surface)',
              borderTopLeftRadius: 20, borderTopRightRadius: 20,
              padding: '12px 16px calc(20px + env(safe-area-inset-bottom, 8px))',
              boxShadow: '0 -12px 32px rgba(0,0,0,0.18)',
              animation: 'sheetUp 0.22s cubic-bezier(0.4, 0, 0.2, 1)',
              transform: `translateY(${currentY}px)`,
              transition: currentY === 0 ? 'transform 0.2s' : 'none',
            }}
          >
            <div style={{
              width: 36, height: 4, borderRadius: 2, background: 'var(--color-border)',
              margin: '4px auto 14px',
            }} />
            <h3 style={{
              fontSize: 11, fontWeight: 700, color: 'var(--color-text-tertiary)',
              textTransform: 'uppercase', letterSpacing: '0.08em', margin: '0 12px 8px',
            }}>More</h3>

            <SheetRow icon={ICONS.settings(20)} label="Settings" onClick={() => navigate('/settings')} />
            {isAdmin && <SheetRow icon={ICONS.admin(20)} label="Admin" onClick={() => navigate('/admin')} />}

            <div style={{ height: 1, background: 'var(--color-border)', margin: '8px 0' }} />

            <SheetRow
              icon={theme === 'dark' ? ICONS.sun(20) : ICONS.moon(20)}
              label={theme === 'dark' ? 'Light mode' : 'Dark mode'}
              onClick={toggle}
            />
            <SheetRow
              icon={ICONS.logout(20)}
              label="Sign Out"
              danger
              onClick={() => { setShowMore(false); logout() }}
            />
            <div style={{ height: 1, background: 'var(--color-border)', margin: '8px 0' }} />
            <div
              style={{
                fontSize: 11,
                textAlign: 'center',
                color: 'var(--color-text-tertiary)',
                lineHeight: 1.4,
                paddingTop: 4,
              }}
            >
              <div style={{ fontWeight: 600 }}>Khorcha-Pati</div>
              <div style={{ fontSize: 9, opacity: 0.8 }}>
                © {new Date().getFullYear()} by <span style={{ fontWeight: 600, color: 'var(--color-text-secondary)' }}>Masudur Rahman</span>
              </div>
            </div>
          </div>
          <style>{`@keyframes sheetUp { from { transform: translateY(100%); } to { transform: translateY(0); } }`}</style>
        </div>
      )}
    </>
  )
}

function SheetRow({ icon, label, onClick, danger }: { icon: React.ReactNode; label: string; onClick: () => void; danger?: boolean }) {
  return (
    <button
      onClick={onClick}
      style={{
        display: 'flex', alignItems: 'center', gap: 14, width: '100%',
        padding: '14px 12px', borderRadius: 'var(--radius-md)',
        background: 'transparent', border: 'none', cursor: 'pointer',
        color: danger ? 'var(--color-danger)' : 'var(--color-text-primary)',
        fontSize: 14, fontWeight: 600, fontFamily: 'inherit', textAlign: 'left',
      }}
    >
      <span style={{ display: 'flex', color: danger ? 'var(--color-danger)' : 'var(--color-text-tertiary)' }}>{icon}</span>
      {label}
    </button>
  )
}
