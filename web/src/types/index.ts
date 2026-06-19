export interface Transaction {
  id: number
  userId: number
  amount: number
  subcategoryId: string
  type: 'Expense' | 'Income' | 'Transfer'
  srcId: string
  dstId: string
  contactName: string
  timestamp: number
  remarks: string
  deletedAt: number
  createdAt: number
}

export interface Wallet {
  id: number
  userId: number
  type: 'Cash' | 'Bank'
  shortName: string
  name: string
  balance: number
  version: number
}

export interface Contact {
  id: number
  userId: number
  nickName: string
  fullName: string
  email: string
  netBalance: number
  lastTxnTimestamp: number
}

export interface BudgetStatus {
  id: number
  categoryId: string
  categoryName: string
  amount: number
  spent: number
  remaining: number
  percent: number
  alertAt: number
}

export interface BudgetAlert {
  categoryId: string
  categoryName: string
  budgetAmount: number
  spent: number
  percent: number
}

export interface CategorySpend {
  categoryId: string
  categoryName: string
  amount: number
  percent: number
}

export interface MonthlyComparison {
  month: string
  income: number
  expense: number
}

export interface MonthlyOverview {
  totalBalance: number
  monthIncome: number
  monthExpense: number
  budgetUsage: number
}

export type TxnType = 'Expense' | 'Income' | 'Transfer'

export interface TxnCategory {
  id: string
  name: string
  types: TxnType[]
}

export interface TxnSubcategory {
  id: string
  catId: string
  name: string
  types: TxnType[]
}

export interface Profile {
  id: number
  telegramId: number
  username: string
  firstName: string
  lastName: string
  timezone: string
  mobileNumber: string
}

export interface ChartData {
  overview: MonthlyOverview
  categories: CategorySpend[]
  comparison: MonthlyComparison[]
}

export interface FieldCost {
  name: string
  amount: number
  type?: string
}

export interface SummaryGroups {
  type: Record<string, FieldCost>
  category: Record<string, FieldCost>
  subcategory: Record<string, FieldCost>
}

export interface StatementTransaction {
  date: string
  type: string
  amount: number
  source: string
  destination: string
  person: string
  category: string
  subcategory: string
  remarks: string
  runningBalance?: number
}

export interface StatementWallet {
  id: number
  type: string
  shortName: string
  name: string
  balance: number
}

export interface StatementContact {
  id: number
  nickName: string
  fullName: string
  netBalance: number
}

export interface StatementReport {
  name: string
  transactions: StatementTransaction[]
  summary: SummaryGroups
  wallets?: StatementWallet[]
  contacts?: StatementContact[]
  startDate: string
  endDate: string
  typeSummary?: FieldCost[]
  categorySummary?: FieldCost[]
  subcategorySummary?: FieldCost[]
  totalAmount: number
  netBalance: number
  generatedAt?: string
}
