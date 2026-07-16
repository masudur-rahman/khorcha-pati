import { useEffect, useState } from 'react'
import { Outlet, useLocation } from 'react-router-dom'
import Sidebar from './Sidebar'
import MobileNav from './MobileNav'
import { ICONS } from '../ui/Icons'
import TxnDialog from '../ui/TxnDialog'
import { useServerTheme } from '../../hooks/useThemeSync'

export default function AppLayout() {
  const location = useLocation()
  // Apply the profile's saved theme once the profile loads.
  useServerTheme()
  const showFab = true
  const [showAddTxn, setShowAddTxn] = useState(false)
  const [fabVisible, setFabVisible] = useState(true)

  // Reset scroll to the top whenever the route changes so a page never opens mid-scroll.
  useEffect(() => {
    window.scrollTo(0, 0)
    setFabVisible(true)
  }, [location.pathname])

  // Track scroll direction to hide/show FAB
  useEffect(() => {
    let lastScrollY = window.scrollY
    const handleScroll = () => {
      const currentScrollY = window.scrollY
      if (currentScrollY > lastScrollY && currentScrollY > 60) {
        setFabVisible(false)
      } else {
        setFabVisible(true)
      }
      lastScrollY = currentScrollY
    }
    window.addEventListener('scroll', handleScroll, { passive: true })
    return () => window.removeEventListener('scroll', handleScroll)
  }, [])

  return (
    <div style={{ display: 'flex', minHeight: '100vh', background: 'var(--color-bg)' }}>
      <Sidebar />
      <main className="px-4 pb-24 md:px-8 md:pb-10" style={{ flex: 1, minWidth: 0 }}>
        <Outlet />
      </main>
      <MobileNav />
      {showFab && (
        <button
          aria-label="Add transaction"
          onClick={() => setShowAddTxn(true)}
          className="khp-fab md:hidden"
          data-tooltip="Add transaction"
          style={{
            position: 'fixed',
            right: 20,
            bottom: 'calc(72px + env(safe-area-inset-bottom, 8px))',
            width: 56,
            height: 56,
            borderRadius: '50%',
            background: 'var(--color-primary)',
            color: 'white',
            border: 'none',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            boxShadow: '0 8px 24px rgba(0, 82, 204, 0.35)',
            cursor: 'pointer',
            zIndex: 90,
            transform: fabVisible ? 'scale(1)' : 'scale(0) translateY(40px)',
            opacity: fabVisible ? 1 : 0,
            transition: 'transform 0.25s cubic-bezier(0.4, 0, 0.2, 1), opacity 0.25s',
            pointerEvents: fabVisible ? 'auto' : 'none',
          }}
        >
          {ICONS.plus(24)}
        </button>
      )}
      {showAddTxn && <TxnDialog initialType="Expense" onClose={() => setShowAddTxn(false)} />}
    </div>
  )
}
