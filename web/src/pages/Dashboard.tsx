import { useMemo, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getChartData, listCategories, listSubcategories } from '../api/endpoints'
import { useTransactions } from '../hooks/useTransactions'
import { fmt } from '../lib/formatter'

import TopBar from '../components/layout/TopBar'
import Card from '../components/ui/Card'
import Badge from '../components/ui/Badge'
import Modal from '../components/ui/Modal'
import MiniDonut from '../components/charts/MiniDonut'
import MiniBarChart from '../components/charts/MiniBarChart'
import BudgetGauge from '../components/charts/BudgetGauge'
import { ICONS } from '../components/ui/Icons'
import WalletFlow from '../components/ui/WalletFlow'
import { useWallets } from '../hooks/useWallets'
import { useContacts } from '../hooks/useContacts'
import TransactionDetails from '../components/ui/TransactionDetails'

export default function Dashboard() {
  const navigate = useNavigate()
  const { data: charts, isLoading: isChartsLoading } = useQuery({
    queryKey: ['chartData'],
    queryFn: () => getChartData(),
  })
  const { data: resp } = useTransactions()
  const txns = resp?.data ?? []
  const { data: wallets } = useWallets()
  const { data: contacts } = useContacts()
  const { data: allCategories, isLoading: isCatsLoading } = useQuery({ queryKey: ['categories'], queryFn: () => listCategories() })
  const { data: subcategories, isLoading: isSubsLoading } = useQuery({ queryKey: ['subcategories'], queryFn: () => listSubcategories() })
  const [showStatementModal, setShowStatementModal] = useState(false)
  const [selectedTxn, setSelectedTxn] = useState<any>(null)

  const isLoading = isChartsLoading || isCatsLoading || isSubsLoading

  const catMap = useMemo(() => {
    const m = new Map<string, string>()
    allCategories?.forEach(c => m.set(c.id, c.name))
    return m
  }, [allCategories])

  const subcatMap = useMemo(() => {
    const m = new Map<string, string>()
    subcategories?.forEach(s => m.set(s.id, s.name))
    return m
  }, [subcategories])

  const chartCategories = useMemo(() => {
    const categorySpends = [...(charts?.categories || [])].sort((a, b) => b.amount - a.amount)
    const colors = ['var(--color-primary)', '#FF991F', 'var(--color-danger)', '#6554C0', 'var(--color-text-tertiary)']
    if (categorySpends.length <= 5) {
      return categorySpends.map((c, i) => ({
        value: c.amount,
        color: colors[i] || colors[4],
        name: catMap.get(c.categoryId) || c.categoryName || c.categoryId
      }))
    }
    const top4 = categorySpends.slice(0, 4).map((c, i) => ({
      value: c.amount, color: colors[i], name: catMap.get(c.categoryId) || c.categoryName || c.categoryId
    }))
    const othersAmount = categorySpends.slice(4).reduce((sum, c) => sum + c.amount, 0)
    top4.push({ value: othersAmount, color: colors[4], name: 'Others' })
    return top4
  }, [charts?.categories, catMap])

  const comparisonData = useMemo(() => {
    return (charts?.comparison || []).map(d => ({
      label: d.month, income: d.income, expense: d.expense
    }))
  }, [charts?.comparison])

  if (isLoading) return <p style={{ color: 'var(--color-text-tertiary)', padding: 40 }}>Loading...</p>
  if (!charts) return null

  const overview = charts.overview
  const recentTxns = [...txns].sort((a, b) => b.timestamp - a.timestamp).slice(0, 5)
  const hour = new Date().getHours()
  const greeting = hour < 12 ? 'Good morning' : hour < 17 ? 'Good afternoon' : 'Good evening'

  const quickActions = [
    { label: 'Add Expense', icon: ICONS.arrowDown, onClick: () => navigate('/transactions?add=Expense'), mobileOrder: 1 },
    { label: 'Add Income', icon: ICONS.arrowUp, onClick: () => navigate('/transactions?add=Income'), mobileOrder: 3 },
    { label: 'Transfer', icon: ICONS.transfer, onClick: () => navigate('/transactions?add=Transfer'), mobileOrder: 2 },
    { label: 'Statement', icon: ICONS.file, onClick: () => setShowStatementModal(true), mobileOrder: 4 },
  ]

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 28 }}>
      <TopBar title="Dashboard" subtitle="Summary of your financial activity" />

      {/* Hero Balance Card */}
      <div style={{
        background: 'var(--hero-gradient)',
        borderRadius: 'var(--radius-xl)',
        padding: '32px 36px',
        color: 'white',
        position: 'relative',
        overflow: 'hidden',
      }}>
        {/* Decorative circles */}
        <div style={{ position: 'absolute', top: -50, right: -30, width: 220, height: 220, borderRadius: '50%', background: 'rgba(255,255,255,0.06)' }} />
        <div style={{ position: 'absolute', bottom: -70, right: 100, width: 180, height: 180, borderRadius: '50%', background: 'rgba(255,255,255,0.04)' }} />
        <div style={{ position: 'absolute', top: 20, right: 180, width: 80, height: 80, borderRadius: '50%', background: 'rgba(255,255,255,0.03)' }} />

        <div style={{ position: 'relative', zIndex: 1 }}>
          <p style={{ fontSize: 14, opacity: 0.85, margin: 0, fontWeight: 500 }}>{greeting}</p>
          <h2 style={{ fontSize: 28, fontWeight: 700, margin: '6px 0 24px', letterSpacing: '-0.02em', fontFamily: "var(--font-display)" }}>
            Current Balance is {fmt(overview.totalBalance)}
          </h2>
          <div className="balance-actions">
            {quickActions.map(a => (
              <button
                key={a.label}
                onClick={a.onClick}
                data-order={a.mobileOrder}
                style={{
                  display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8, padding: '10px 18px', whiteSpace: 'nowrap',
                  borderRadius: 'var(--radius-md)',
                  background: 'rgba(255,255,255,0.15)',
                  backdropFilter: 'blur(12px)',
                  border: '1px solid rgba(255,255,255,0.2)',
                  color: 'white', fontSize: 13, fontWeight: 600,
                  cursor: 'pointer', transition: 'all var(--transition-fast)',
                  fontFamily: 'inherit',
                }}
                onMouseEnter={e => e.currentTarget.style.background = 'rgba(255,255,255,0.28)'}
                onMouseLeave={e => e.currentTarget.style.background = 'rgba(255,255,255,0.15)'}
              >
                {a.icon(14)} {a.label}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Summary Stat Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 20 }}>
        <SummaryCard
          label="Total Income"
          value={`+${fmt(overview.monthIncome)}`}
          subtext="This month"
          accentColor="var(--color-success)"
          icon={ICONS.arrowUp(20)}
        />
        <SummaryCard
          label="Total Expense"
          value={`-${fmt(overview.monthExpense)}`}
          subtext="This month"
          accentColor="var(--color-danger)"
          icon={ICONS.arrowDown(20)}
        />
        <SummaryCard
          label="Net Balance"
          value={fmt(overview.totalBalance)}
          subtext="All wallets"
          accentColor="var(--color-primary)"
          icon={ICONS.money(20)}
        />
        <Card style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 8 }}>
          <span style={{ fontSize: 11, fontWeight: 700, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.06em' }}>Budget Usage</span>
          <BudgetGauge percent={overview.budgetUsage} size={110} />
        </Card>
      </div>

      {/* Charts Row */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: 20 }}>
        <Card>
          <h3 style={{ fontSize: 15, fontWeight: 700, color: 'var(--color-text-primary)', margin: '0 0 20px', display: 'flex', alignItems: 'center', gap: 10 }}>
            <span style={{ width: 3, height: 18, borderRadius: 2, background: 'var(--color-primary)' }} />
            Expense by Category
          </h3>
          <div style={{ display: 'flex', alignItems: 'center', gap: 24 }}>
            <MiniDonut segments={chartCategories} size={130} />
            <div style={{ display: 'flex', flexDirection: 'column', gap: 10, flex: 1 }}>
              {chartCategories.map(c => (
                <div key={c.name} style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                  <span style={{ width: 8, height: 8, borderRadius: 4, background: c.color, flexShrink: 0 }} />
                  <span style={{ fontSize: 12, color: 'var(--color-text-secondary)', flex: 1 }}>{c.name}</span>
                  <span style={{ fontSize: 12, fontWeight: 700, color: 'var(--color-text-primary)' }}>{fmt(c.value)}</span>
                </div>
              ))}
            </div>
          </div>
        </Card>
        <Card>
          <h3 style={{ fontSize: 15, fontWeight: 700, color: 'var(--color-text-primary)', margin: '0 0 20px', display: 'flex', alignItems: 'center', gap: 10 }}>
            <span style={{ width: 3, height: 18, borderRadius: 2, background: 'var(--color-success)' }} />
            Income vs Expense
          </h3>
          <div style={{ display: 'flex', justifyContent: 'center' }}>
            <MiniBarChart data={comparisonData} size={{ w: 280, h: 130 }} />
          </div>
          <div style={{ display: 'flex', gap: 20, justifyContent: 'center', marginTop: 12 }}>
            <span style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 11, color: 'var(--color-text-tertiary)', fontWeight: 600 }}>
              <span style={{ width: 8, height: 8, borderRadius: 4, background: 'var(--color-success)' }} /> Income
            </span>
            <span style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 11, color: 'var(--color-text-tertiary)', fontWeight: 600 }}>
              <span style={{ width: 8, height: 8, borderRadius: 4, background: 'var(--color-danger)' }} /> Expense
            </span>
          </div>
        </Card>
      </div>

      {/* Recent Transactions */}
      <Card padding={0}>
        <div style={{ padding: '20px 24px', borderBottom: '1px solid var(--color-border)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <h3 style={{ fontSize: 15, fontWeight: 700, color: 'var(--color-text-primary)', margin: 0 }}>Recent Transactions</h3>
          <Link to="/transactions" style={{ fontSize: 13, fontWeight: 600, color: 'var(--color-primary)', textDecoration: 'none' }}>View All</Link>
        </div>
        <div style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 13 }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--color-border)' }}>
                {['Date', 'Type', 'Category', 'Amount', 'Wallet'].map(h => (
                  <th key={h} style={{ padding: '12px 24px', textAlign: h === 'Amount' ? 'right' : h === 'Wallet' ? 'center' : 'left', fontSize: 10, fontWeight: 700, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.08em' }}>{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {recentTxns.map(t => (
                <tr
                  key={t.id}
                  style={{ borderBottom: '1px solid var(--color-border)', cursor: 'pointer' }}
                  className="hover-row transition-colors"
                  onClick={() => setSelectedTxn(t)}
                >
                  <td style={{ padding: '14px 24px', color: 'var(--color-text-tertiary)', fontSize: 12, fontWeight: 600 }}>
                    {new Date(t.timestamp * 1000).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
                  </td>
                  <td style={{ padding: '14px 24px' }}><Badge type={t.type as any} /></td>
                  <td style={{ padding: '14px 24px', fontWeight: 600, color: 'var(--color-text-primary)' }}>
                    {subcatMap.get(t.subcategoryId) || t.subcategoryId}
                  </td>
                  <td style={{
                    padding: '14px 24px', textAlign: 'right', fontWeight: 700, fontSize: 14,
                    color: t.type === 'Income' ? 'var(--color-success)' : t.type === 'Transfer' ? 'var(--color-primary)' : 'var(--color-danger)',
                  }}>
                    {t.type === 'Income' ? '+' : t.type === 'Transfer' ? '' : '-'}{fmt(t.amount)}
                  </td>
                  <td style={{ padding: '14px 24px' }}>
                    <WalletFlow srcId={t.srcId} dstId={t.dstId} contactName={t.contactName} />
                  </td>
                </tr>
              ))}
              {recentTxns.length === 0 && (
                <tr>
                  <td colSpan={5} style={{ padding: 48, textAlign: 'center', color: 'var(--color-text-tertiary)' }}>
                    No recent transactions found
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </Card>

      {/* Transaction Detail Slide-in */}
      {selectedTxn && (
        <TransactionDetails
          txn={selectedTxn}
          wallets={wallets ?? []}
          contacts={contacts ?? []}
          categories={allCategories ?? []}
          subcategories={subcategories ?? []}
          onClose={() => setSelectedTxn(null)}
          onEdit={(t) => navigate(`/transactions?edit=${t.id}`)}
        />
      )}

      {showStatementModal && <StatementModal onClose={() => setShowStatementModal(false)} />}
    </div>
  )
}

/* Summary Card with accent left border */
function SummaryCard({ label, value, subtext, accentColor, icon }: {
  label: string; value: string; subtext: string; accentColor: string; icon: React.ReactNode
}) {
  return (
    <Card style={{ borderLeft: `4px solid ${accentColor}`, display: 'flex', flexDirection: 'column', gap: 12 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <span style={{ fontSize: 11, fontWeight: 700, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.06em' }}>{label}</span>
        <div style={{
          width: 36, height: 36, borderRadius: 'var(--radius-sm)',
          background: accentColor + '15', color: accentColor,
          display: 'flex', alignItems: 'center', justifyContent: 'center',
        }}>{icon}</div>
      </div>
      <div>
        <div style={{ fontSize: 24, fontWeight: 700, color: accentColor, letterSpacing: '-0.02em' }}>{value}</div>
        <p style={{ fontSize: 11, color: 'var(--color-text-tertiary)', marginTop: 4, fontWeight: 500 }}>{subtext}</p>
      </div>
    </Card>
  )
}


function StatementModal({ onClose }: { onClose: () => void }) {
  const durations = [
    { label: 'This Week', value: 'one_week' },
    { label: 'Current Month', value: 'this_month' },
    { label: 'Last 30 Days', value: 'one_month' },
    { label: 'Last 6 Months', value: 'half_year' },
    { label: 'Current Year', value: 'this_year' },
    { label: 'Last 1 Year', value: 'one_year' },
    { label: 'All Time', value: 'all_time' },
  ]

  const handlePreview = (duration: string) => {
    window.open(`/statement?duration=${duration}`, '_blank')
    onClose()
  }

  return (
    <Modal title="Generate Statement" onClose={onClose} width={400}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
        <p style={{ fontSize: 14, color: 'var(--color-text-secondary)', marginBottom: 8 }}>
          Select a time period for your financial statement.
        </p>
        {durations.map(d => (
          <button
            key={d.value}
            onClick={() => handlePreview(d.value)}
            style={{
              padding: '14px 20px', borderRadius: 'var(--radius-md)',
              border: '1px solid var(--color-border)', background: 'var(--color-surface)',
              cursor: 'pointer', fontSize: 14, fontWeight: 600, color: 'var(--color-text-primary)',
              textAlign: 'left', transition: 'all var(--transition-fast)', fontFamily: 'inherit',
            }}
            onMouseEnter={e => { e.currentTarget.style.background = 'var(--color-primary-subtle)'; e.currentTarget.style.borderColor = 'var(--color-primary)' }}
            onMouseLeave={e => { e.currentTarget.style.background = 'var(--color-surface)'; e.currentTarget.style.borderColor = 'var(--color-border)' }}
          >
            {d.label}
          </button>
        ))}
      </div>
    </Modal>
  )
}
