import React, { useState } from 'react'
import { NavLink } from 'react-router-dom'
import { ICONS } from '../ui/Icons'
import Logo from '../ui/Logo'
import { useAuth } from '../../hooks/useAuth'

const navItems = [
  { id: 'dashboard', to: '/', label: 'Dashboard', icon: ICONS.dashboard },
  { id: 'transactions', to: '/transactions', label: 'Transactions', icon: ICONS.transactions },
  { id: 'wallets', to: '/wallets', label: 'Wallets', icon: ICONS.wallet },
  { id: 'budgets', to: '/budgets', label: 'Budgets', icon: ICONS.budget },
  { id: 'settings', to: '/settings', label: 'Settings', icon: ICONS.settings },
]

export default function Sidebar() {
  const { logout } = useAuth()
  const [collapsed, setCollapsed] = useState(false)

  const sidebarStyles: Record<string, React.CSSProperties> = {
    sidebar: {
      width: collapsed ? 72 : 260,
      height: '100vh',
      background: 'var(--color-sidebar-bg)',
      borderRight: '1px solid var(--color-border)',
      display: 'flex',
      flexDirection: 'column',
      transition: 'width 0.25s cubic-bezier(.4,0,.2,1)',
      position: 'sticky',
      top: 0,
      zIndex: 100,
      flexShrink: 0,
    },
    header: {
      padding: collapsed ? '24px 0' : '24px 20px',
      display: 'flex',
      alignItems: 'center',
      justifyContent: collapsed ? 'center' : 'flex-start',
      borderBottom: '1px solid var(--color-border)',
      height: 80,
      flexShrink: 0,
    },
    nav: {
      flex: 1,
      padding: collapsed ? '16px 10px' : '16px 12px',
      display: 'flex',
      flexDirection: 'column',
      gap: 4,
      overflowY: 'auto',
      overflowX: 'hidden',
    },
    toggle: {
      position: 'absolute',
      right: -14,
      top: 32,
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
      boxShadow: '0 2px 8px rgba(0,0,0,0.06)',
      color: 'var(--color-text-secondary)',
      transition: 'all 0.15s',
    },
    footer: {
      padding: collapsed ? '16px 10px' : '16px 12px',
      borderTop: '1px solid var(--color-border)',
      flexShrink: 0,
    },
  }

  const getItemStyle = (active: boolean): React.CSSProperties => ({
    display: 'flex',
    alignItems: 'center',
    gap: 12,
    padding: collapsed ? '12px 0' : '12px 16px',
    justifyContent: collapsed ? 'center' : 'flex-start',
    borderRadius: 12,
    fontSize: 14,
    fontWeight: active ? 600 : 500,
    color: active ? 'var(--color-primary)' : 'var(--color-text-secondary)',
    background: active ? 'var(--color-primary-subtle)' : 'transparent',
    cursor: 'pointer',
    border: 'none',
    width: '100%',
    textAlign: 'left',
    transition: 'all 0.15s ease',
    whiteSpace: 'nowrap',
    position: 'relative',
    textDecoration: 'none',
  })

  return (
    <div style={sidebarStyles.sidebar} className="hidden md:flex">
      <div 
        style={sidebarStyles.toggle} 
        onClick={() => setCollapsed(!collapsed)}
        onMouseEnter={e => e.currentTarget.style.background = 'var(--color-primary-subtle)'}
        onMouseLeave={e => e.currentTarget.style.background = 'var(--color-surface)'}
      >
        {collapsed ? ICONS.chevronRight(14) : ICONS.chevronLeft(14)}
      </div>

      <div style={sidebarStyles.header}>
        <Logo size={collapsed ? 28 : 32} collapsed={collapsed} />
      </div>

      <nav style={sidebarStyles.nav}>
        {navItems.map(item => (
          <NavLink
            key={item.id}
            to={item.to}
            style={({ isActive }) => getItemStyle(isActive)}
            title={collapsed ? item.label : undefined}
          >
            {({ isActive }) => (
              <>
                {isActive && (
                  <div style={{
                    position: 'absolute',
                    left: collapsed ? '50%' : 0,
                    top: collapsed ? 'auto' : '50%',
                    bottom: collapsed ? 0 : 'auto',
                    transform: collapsed ? 'translateX(-50%)' : 'translateY(-50%)',
                    width: collapsed ? 20 : 3,
                    height: collapsed ? 3 : 20,
                    borderRadius: 2,
                    background: 'var(--color-primary)',
                  }} />
                )}
                <span style={{ flexShrink: 0, opacity: isActive ? 1 : 0.6 }}>{item.icon(20)}</span>
                {!collapsed && <span>{item.label}</span>}
              </>
            )}
          </NavLink>
        ))}
      </nav>

      <div style={sidebarStyles.footer}>
        <button
          style={{ ...getItemStyle(false), color: 'var(--color-danger)' }}
          onClick={logout}
          onMouseEnter={e => e.currentTarget.style.background = 'var(--color-danger-subtle)'}
          onMouseLeave={e => e.currentTarget.style.background = 'transparent'}
          title={collapsed ? 'Logout' : undefined}
        >
          <span style={{ flexShrink: 0 }}>{ICONS.logout(20)}</span>
          {!collapsed && <span>Logout</span>}
        </button>
      </div>
    </div>
  )
}
