import { NavLink } from 'react-router-dom'
import { LayoutDashboard, ArrowLeftRight, Wallet, Target, Settings, LogOut } from 'lucide-react'
import { useAuth } from '../../hooks/useAuth'

const links = [
  { to: '/', icon: LayoutDashboard, label: 'Home' },
  { to: '/transactions', icon: ArrowLeftRight, label: 'Txns' },
  { to: '/wallets', icon: Wallet, label: 'Wallets' },
  { to: '/budgets', icon: Target, label: 'Budgets' },
  { to: '/settings', icon: Settings, label: 'Settings' },
]

export default function MobileNav() {
  const { logout } = useAuth()
  return (
    <nav className="md:hidden fixed bottom-0 left-0 right-0 bg-white border-t flex justify-around py-2 z-50">
      {links.map(({ to, icon: Icon, label }) => (
        <NavLink
          key={to}
          to={to}
          className={({ isActive }) =>
            `flex flex-col items-center text-xs ${isActive ? 'text-blue-600' : 'text-gray-500'}`
          }
        >
          <Icon size={20} />
          {label}
        </NavLink>
      ))}
      <button
        onClick={logout}
        className="flex flex-col items-center text-xs text-gray-500 hover:text-red-500 transition-colors cursor-pointer"
      >
        <LogOut size={20} />
        Logout
      </button>
    </nav>
  )
}
