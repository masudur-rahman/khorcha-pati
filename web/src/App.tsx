import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider } from './context/AuthContext'
import { SearchProvider } from './context/SearchContext'
import { useAuth } from './hooks/useAuth'
import AppLayout from './components/layout/AppLayout'
import Landing from './pages/Landing'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import Transactions from './pages/Transactions'
import Wallets from './pages/Wallets'
import Budgets from './pages/Budgets'
import Settings from './pages/Settings'
import Statement from './pages/Statement'
import Admin from './pages/Admin'

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: 1, staleTime: 30_000 } },
})

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <SearchProvider>
          <BrowserRouter>
            <AppRoutes />
          </BrowserRouter>
        </SearchProvider>
      </AuthProvider>
    </QueryClientProvider>
  )
}

function AppRoutes() {
  const { isAuthenticated, isLoading, isAdmin } = useAuth()

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <p className="text-muted-foreground">Loading...</p>
      </div>
    )
  }

  if (!isAuthenticated) {
    return (
      <Routes>
        <Route path="/" element={<Landing />} />
        <Route path="/login" element={<Login />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    )
  }

  return (
    <Routes>
      <Route path="/statement" element={<Statement />} />
      <Route element={<AppLayout />}>
        <Route path="/" element={<Dashboard />} />
        <Route path="/transactions" element={<Transactions />} />
        <Route path="/wallets" element={<Wallets />} />
        <Route path="/budgets" element={<Budgets />} />
        <Route path="/settings" element={<Settings />} />
        {isAdmin && <Route path="/admin" element={<Admin />} />}
      </Route>
      <Route path="/login" element={<Navigate to="/" replace />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
