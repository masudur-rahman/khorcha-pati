import { NavLink } from 'react-router-dom'
import { LayoutDashboard, ArrowLeftRight, Wallet, Target, Settings, LogOut } from 'lucide-react'
import { useAuth } from '../../hooks/useAuth'

const links = [
  { to: '/', icon: LayoutDashboard, label: 'Dashboard' },
  { to: '/transactions', icon: ArrowLeftRight, label: 'Transactions' },
  { to: '/wallets', icon: Wallet, label: 'Wallets' },
  { to: '/budgets', icon: Target, label: 'Budgets' },
  { to: '/settings', icon: Settings, label: 'Settings' },
]

export default function Sidebar() {
  const { logout } = useAuth()
  return (
    <aside className="hidden md:flex flex-col w-64 bg-gray-900 text-gray-300 h-screen sticky top-0 p-4 border-r border-gray-800">
      <div className="flex items-center gap-3 px-3 mb-10 mt-2">
        <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center text-white font-bold">E</div>
        <h2 className="text-xl font-bold text-white tracking-tight">Expense<span className="text-blue-500"> Tracker</span></h2>
      </div>
      
      <nav className="flex-1 space-y-2">
        {links.map(({ to, icon: Icon, label }) => (
          <NavLink
            key={to}
            to={to}
            className={({ isActive }) =>
              `flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-bold transition-all cursor-pointer ${
                isActive 
                  ? 'bg-blue-600/10 text-blue-500 shadow-sm shadow-blue-900/20' 
                  : 'hover:bg-gray-800 hover:text-white'
              }`
            }
          >
            <Icon size={20} />
            {label}
          </NavLink>
        ))}
      </nav>

      <div className="pt-4 border-t border-gray-800 mt-4">
        <button
          className="w-full flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-bold hover:bg-red-500/10 text-gray-400 hover:text-red-500 transition-all cursor-pointer"
          onClick={logout}
        >
          <LogOut size={20} /> Logout
        </button>
      </div>
    </aside>
  )
}
