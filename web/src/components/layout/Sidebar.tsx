import { useMemo, useState } from 'react'
import { NavLink } from 'react-router-dom'
import { useAuth } from '../../hooks/useAuth'
import { ICONS } from '../ui/Icons'
import Logo from '../ui/Logo'

const baseNavItems = [
  { id: 'dashboard', to: '/', label: 'Dashboard', icon: ICONS.dashboard },
  { id: 'transactions', to: '/transactions', label: 'Transactions', icon: ICONS.transactions },
  { id: 'wallets', to: '/wallets', label: 'Wallets', icon: ICONS.wallet },
  { id: 'budgets', to: '/budgets', label: 'Budgets', icon: ICONS.budget },
  { id: 'settings', to: '/settings', label: 'Settings', icon: ICONS.settings },
]

export default function Sidebar() {
  const [collapsed, setCollapsed] = useState(false)
  const { isAdmin } = useAuth()
  const navItems = useMemo(() => {
    if (!isAdmin) return baseNavItems
    return [...baseNavItems, { id: 'admin', to: '/admin', label: 'Admin', icon: ICONS.admin }]
  }, [isAdmin])

  return (
    <div
      style={{
        width: collapsed ? 76 : 264,
        height: '100vh',
        background: 'var(--color-sidebar-bg)',
        borderRight: '1px solid var(--color-border)',
        flexDirection: 'column',
        transition: 'width var(--transition-normal)',
        position: 'sticky',
        top: 0,
        zIndex: 100,
        flexShrink: 0,
      }}
      className="hidden md:flex"
    >
      {/* Collapse toggle */}
      <button
        style={{
          position: 'absolute',
          right: -14,
          top: 36,
          width: 28,
          height: 28,
          borderRadius: '50%',
          background: 'var(--color-surface)',
          border: '1px solid var(--color-border)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          cursor: 'pointer',
          zIndex: 110,
          boxShadow: '0 2px 8px rgba(0,0,0,0.08)',
          color: 'var(--color-text-secondary)',
          transition: 'all var(--transition-fast)',
        }}
        onClick={() => setCollapsed(!collapsed)}
        onMouseEnter={e => {
          e.currentTarget.style.background = 'var(--color-primary)'
          e.currentTarget.style.color = 'white'
          e.currentTarget.style.borderColor = 'var(--color-primary)'
        }}
        onMouseLeave={e => {
          e.currentTarget.style.background = 'var(--color-surface)'
          e.currentTarget.style.color = 'var(--color-text-secondary)'
          e.currentTarget.style.borderColor = 'var(--color-border)'
        }}
      >
        {collapsed ? ICONS.chevronRight(14) : ICONS.chevronLeft(14)}
      </button>

      {/* Logo header */}
      <NavLink
        to="/"
        style={{
          padding: collapsed ? '24px 0' : '24px 24px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: collapsed ? 'center' : 'flex-start',
          borderBottom: '1px solid var(--color-border)',
          height: 80,
          flexShrink: 0,
          textDecoration: 'none',
          color: 'inherit',
        }}
      >
        <Logo size={collapsed ? 28 : 32} collapsed={collapsed} />
      </NavLink>

      {/* Navigation */}
      <nav
        style={{
          flex: 1,
          padding: collapsed ? '20px 12px' : '20px 16px',
          display: 'flex',
          flexDirection: 'column',
          gap: 4,
          overflowY: 'auto',
          overflowX: 'hidden',
        }}
      >
        {navItems.map(item => (
          <NavLink
            key={item.id}
            to={item.to}
            style={({ isActive }) => ({
              display: 'flex',
              alignItems: 'center',
              gap: 14,
              padding: collapsed ? '12px 0' : '11px 16px',
              justifyContent: collapsed ? 'center' : 'flex-start',
              borderRadius: 'var(--radius-md)',
              fontSize: 14,
              fontWeight: isActive ? 600 : 500,
              color: isActive ? 'var(--color-primary)' : 'var(--color-text-secondary)',
              background: isActive ? 'var(--color-primary-subtle)' : 'transparent',
              cursor: 'pointer',
              textDecoration: 'none',
              transition: 'all var(--transition-fast)',
              position: 'relative',
              whiteSpace: 'nowrap',
            })}
            title={collapsed ? item.label : undefined}
          >
            {({ isActive }) => (
              <>
                {isActive && (
                  <div
                    style={{
                      position: 'absolute',
                      left: collapsed ? '50%' : 0,
                      top: collapsed ? 'auto' : '50%',
                      bottom: collapsed ? 0 : 'auto',
                      transform: collapsed ? 'translateX(-50%)' : 'translateY(-50%)',
                      width: collapsed ? 20 : 3,
                      height: collapsed ? 3 : 24,
                      borderRadius: 2,
                      background: 'var(--color-primary)',
                    }}
                  />
                )}
                <span style={{ flexShrink: 0, opacity: isActive ? 1 : 0.65 }}>
                  {item.icon(20)}
                </span>
                {!collapsed && <span>{item.label}</span>}
              </>
            )}
          </NavLink>
        ))}
      </nav>

      {/* Version info footer */}
      <div
        style={{
          padding: collapsed ? '16px 12px' : '16px 24px',
          borderTop: '1px solid var(--color-border)',
          flexShrink: 0,
        }}
      >
        {!collapsed && (
          <p
            style={{
              fontSize: 10,
              fontWeight: 600,
              color: 'var(--color-text-tertiary)',
              textTransform: 'uppercase',
              letterSpacing: '0.08em',
              textAlign: 'center',
            }}
          >
            Expense Tracker v2.0
          </p>
        )}
      </div>
    </div>
  )
}
