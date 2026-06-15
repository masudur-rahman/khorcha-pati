import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import Sidebar from './Sidebar'
import MobileNav from './MobileNav'
import { ICONS } from '../ui/Icons'

export default function AppLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const showFab = location.pathname !== '/transactions'

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
          onClick={() => navigate('/transactions?add=Expense')}
          className="md:hidden"
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
            zIndex: 150,
          }}
        >
          {ICONS.plus(24)}
        </button>
      )}
    </div>
  )
}
