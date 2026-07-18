import { apiFetch, getRefreshToken } from './client'
import type {
  Transaction, Wallet, Contact, BudgetStatus, BudgetAlert,
  ChartData, TxnCategory, TxnSubcategory, Profile, StatementReport,
} from '../types'

const API = '/api/v1'

// Auth
export const requestOTP = (identifier: string) =>
  apiFetch(`${API}/auth/request-otp`, { method: 'POST', body: JSON.stringify({ identifier }) })

export const verifyOTP = (identifier: string, code: string) =>
  apiFetch<{ accessToken: string; refreshToken: string }>(`${API}/auth/verify-otp`, {
    method: 'POST', body: JSON.stringify({ identifier, code }),
  })

export const initQR = () =>
  apiFetch<{ sessionID: string; deepLink: string }>(`${API}/auth/qr/init`, { method: 'POST' })

export const pollQR = (session: string) =>
  apiFetch<{ status: string; accessToken?: string; refreshToken?: string }>(`${API}/auth/qr/status?session=${session}`)

export const verifyMagicLink = (token: string) =>
  apiFetch<{ accessToken: string; refreshToken: string }>(`${API}/auth/magic-link`, {
    method: 'POST', body: JSON.stringify({ token }),
  })

export const logout = () =>
  apiFetch(`${API}/auth/logout`, {
    method: 'POST',
    body: JSON.stringify({ refreshToken: getRefreshToken() }),
  })

// Transactions
export const listTransactions = (params?: Record<string, string>) => {
  const qs = params ? '?' + new URLSearchParams(params).toString() : ''
  return apiFetch<{ data: Transaction[]; pagination: any }>(`${API}/transactions${qs}`)
}

export const createTransaction = (txn: Partial<Transaction>) =>
  apiFetch(`${API}/transactions`, { method: 'POST', body: JSON.stringify(txn) })

export const updateTransaction = (id: number, txn: Partial<Transaction>) =>
  apiFetch(`${API}/transactions/${id}`, { method: 'PUT', body: JSON.stringify(txn) })

export const deleteTransaction = (id: number) =>
  apiFetch(`${API}/transactions/${id}`, { method: 'DELETE' })

// Wallets
export const listWallets = () => apiFetch<Wallet[]>(`${API}/wallets`)

export const createWallet = (wallet: Partial<Wallet>) =>
  apiFetch<Wallet>(`${API}/wallets`, { method: 'POST', body: JSON.stringify(wallet) })

export const updateWallet = (id: number, wallet: Partial<Wallet>) =>
  apiFetch<{ message: string }>(`${API}/wallets/${id}`, { method: 'PUT', body: JSON.stringify(wallet) })

export const deleteWallet = (shortName: string) =>
  apiFetch<{ message: string }>(`${API}/wallets/${shortName}`, { method: 'DELETE' })

// Contacts
export const listContacts = () => apiFetch<Contact[]>(`${API}/contacts`)

export const createContact = (contact: Partial<Contact>) =>
  apiFetch<Contact>(`${API}/contacts`, { method: 'POST', body: JSON.stringify(contact) })

export const updateContact = (id: number, contact: Partial<Contact>) =>
  apiFetch<{ message: string }>(`${API}/contacts/${id}`, { method: 'PUT', body: JSON.stringify(contact) })

export const deleteContact = (id: number) =>
  apiFetch<{ message: string }>(`${API}/contacts/${id}`, { method: 'DELETE' })

// Budgets
export const listBudgets = () => apiFetch<BudgetStatus[]>(`${API}/budgets`)

export const setBudget = (categoryId: string, amount: number, alertAt: number) =>
  apiFetch(`${API}/budgets`, { method: 'POST', body: JSON.stringify({ categoryId, amount, alertAt }) })

export const deleteBudget = (categoryId: string) =>
  apiFetch(`${API}/budgets/${categoryId}`, { method: 'DELETE' })

export const getBudgetAlerts = () => apiFetch<BudgetAlert[]>(`${API}/budgets/alerts`)

// Summary
export const getChartData = (year?: number, month?: number, months?: number) => {
  const params = new URLSearchParams()
  if (year) params.set('year', String(year))
  if (month) params.set('month', String(month))
  if (months) params.set('months', String(months))
  return apiFetch<ChartData>(`${API}/summary/charts?${params}`)
}

export const downloadReport = (duration: string) => {
  const url = `/api/v1/summary/report?duration=${duration}`
  return apiFetch<Blob>(url, { headers: { 'Accept': 'application/pdf' } })
}

export const fetchReportData = (duration?: string, start?: string, end?: string) => {
  const qs = new URLSearchParams()
  if (start && end) {
    qs.set('start', start)
    qs.set('end', end)
  } else if (duration) {
    qs.set('duration', duration)
  }
  return apiFetch<StatementReport>(`${API}/summary/report-data?${qs.toString()}`)
}

// Categories
export const listCategories = (type?: string) => {
  const qs = type ? `?type=${encodeURIComponent(type)}` : ''
  return apiFetch<TxnCategory[]>(`${API}/categories${qs}`)
}

export const listSubcategories = (catId?: string, type?: string) => {
  const params = new URLSearchParams()
  if (catId) params.set('catId', catId)
  if (type) params.set('type', type)
  const qs = params.toString() ? `?${params.toString()}` : ''
  return apiFetch<TxnSubcategory[]>(`${API}/subcategories${qs}`)
}

// Admin
export interface AdminStats {
  userCount: number
  txnCount: number
  walletCount: number
  databaseType: string
}

export interface AdminUser {
  id: number
  telegramId: number
  username: string
  firstName: string
  lastName: string
  isAdmin: boolean
  isActive: boolean
  walletCount: number
  txnCount: number
  contactCount: number
  createdAt: number
  lastTxnAt: number
}

export const getAdminStats = () => apiFetch<AdminStats>(`${API}/admin/stats`)
export interface PaginatedUsers {
  users: AdminUser[]
  total: number
}

export type AdminUserSort = 'registered' | 'last_txn'
export type SortOrder = 'asc' | 'desc'

export const getAdminUsers = (page?: number, limit?: number, sort?: AdminUserSort, order?: SortOrder) => {
  const params = new URLSearchParams()
  if (page !== undefined) params.set('page', String(page))
  if (limit !== undefined) params.set('limit', String(limit))
  if (sort) params.set('sort', sort)
  if (order) params.set('order', order)
  const qs = params.toString() ? `?${params.toString()}` : ''
  return apiFetch<PaginatedUsers>(`${API}/admin/users${qs}`)
}
export const getAdminUserDetail = (id: number) => apiFetch<AdminUser>(`${API}/admin/users/${id}`)
export const setAdminUserActive = (id: number, isActive: boolean) =>
  apiFetch<{ id: number; isActive: boolean }>(`${API}/admin/users/${id}/activate`, {
    method: 'PATCH',
    body: JSON.stringify({ isActive }),
  })

export const sendAdminBroadcast = (message: string, includeUserIds?: number[], excludeUserIds?: number[]) =>
  apiFetch<{ success: boolean; sent: number; failed: number }>(`${API}/admin/broadcast`, {
    method: 'POST',
    body: JSON.stringify({ message, includeUserIds, excludeUserIds }),
  })

// Admin — access control (restricted stage/dev instances)
export interface AllowedUser {
  id: number
  telegramId: number
  username: string
  revoked: boolean
  revokedAt: number
  createdAt: number
}

export interface AccessSettings {
  restricted: boolean
  replyText: string
  allowedUsers: AllowedUser[]
}

export const getAccessSettings = (all?: boolean) =>
  apiFetch<AccessSettings>(`${API}/admin/access${all ? '?all=true' : ''}`)

export const updateAccessSettings = (body: { restricted?: boolean; replyText?: string }) =>
  apiFetch<AccessSettings>(`${API}/admin/access`, { method: 'PUT', body: JSON.stringify(body) })

export const addAllowedUser = (body: { username?: string; telegramId?: number }) =>
  apiFetch<AllowedUser>(`${API}/admin/access/allowed`, { method: 'POST', body: JSON.stringify(body) })

// Soft revoke — the entry is tombstoned, restorable later.
export const removeAllowedUser = (id: number) =>
  apiFetch<{ id: number; revoked: boolean }>(`${API}/admin/access/allowed/${id}`, { method: 'DELETE' })

export const restoreAllowedUser = (id: number) =>
  apiFetch<{ id: number; revoked: boolean }>(`${API}/admin/access/allowed/${id}/restore`, { method: 'POST' })

// Admin — AI classification cache
export interface AICacheEntry {
  id: number
  inputText: string
  subcategoryId: string
  intent: string
  confidence: number
  createdAt: number
}

export interface AICacheInput {
  inputText?: string
  subcategoryId: string
  intent: string
  confidence?: number
}

export const listAICache = (q?: string) => {
  const qs = q ? `?q=${encodeURIComponent(q)}` : ''
  return apiFetch<AICacheEntry[]>(`${API}/admin/ai-cache${qs}`)
}

export const createAICache = (body: AICacheInput) =>
  apiFetch<AICacheEntry>(`${API}/admin/ai-cache`, { method: 'POST', body: JSON.stringify(body) })

export const updateAICache = (id: number, body: AICacheInput) =>
  apiFetch<AICacheEntry>(`${API}/admin/ai-cache/${id}`, { method: 'PUT', body: JSON.stringify(body) })

export const deleteAICache = (id: number) =>
  apiFetch<{ id: number }>(`${API}/admin/ai-cache/${id}`, { method: 'DELETE' })

export interface AICacheExportEntry {
  inputText: string
  subcategoryId: string
  intent: string
  confidence: number
}

export type AICacheImportMode = 'skip' | 'overwrite' | 'confidence'

export interface AICacheImportSummary {
  imported: number
  overwritten: number
  skipped: number
  invalid: number
}

export const exportAICache = () =>
  apiFetch<AICacheExportEntry[]>(`${API}/admin/ai-cache/export`)

export const importAICache = (mode: AICacheImportMode, entries: AICacheExportEntry[]) =>
  apiFetch<AICacheImportSummary>(`${API}/admin/ai-cache/import`, {
    method: 'POST',
    body: JSON.stringify({ mode, entries }),
  })

// Profile
export const getProfile = () => apiFetch<Profile>(`${API}/profile`)

export const updateProfile = (data: Partial<Profile>) =>
  apiFetch<Profile>(`${API}/profile`, { method: 'PUT', body: JSON.stringify(data) })
