import { formatDate } from '../lib/formatter'
import { useMemo, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getChartData, listCategories, listSubcategories , getProfile } from '../api/endpoints'
import { useTransactions } from '../hooks/useTransactions'
import { fmt } from '../lib/formatter'

import TopBar from '../components/layout/TopBar'
import Card from '../components/ui/Card'
import Badge from '../components/ui/Badge'
import Modal from '../components/ui/Modal'
import DateRangePicker from '../components/ui/DateRangePicker'
import MiniDonut from '../components/charts/MiniDonut'
import MiniBarChart from '../components/charts/MiniBarChart'
import BudgetGauge from '../components/charts/BudgetGauge'
import { ICONS } from '../components/ui/Icons'
import WalletFlow from '../components/ui/WalletFlow'
import { useWallets } from '../hooks/useWallets'
import { useContacts } from '../hooks/useContacts'
import TransactionDetails from '../components/ui/TransactionDetails'
import MetricChip from '../components/ui/MetricChip'
import SectionHeader from '../components/ui/SectionHeader'
import Eyebrow from '../components/ui/Eyebrow'
import WalletCard, { inferVariant } from '../components/ui/WalletCard'
import TxnDialog, { TxnType } from '../components/ui/TxnDialog'
import DeleteTxnDialog from '../components/ui/DeleteTxnDialog'

export default function Dashboard() {
  const { data: profile } = useQuery({ queryKey: ['profile'], queryFn: getProfile })
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
  const [addTxnType, setAddTxnType] = useState<TxnType | null>(null)
  const [editTxn, setEditTxn] = useState<any>(null)
  const [deleteTxn, setDeleteTxn] = useState<any>(null)

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
    { label: 'Add Expense', icon: ICONS.shoppingCart, onClick: () => setAddTxnType('Expense'), mobileOrder: 1 },
    { label: 'Add Income', icon: ICONS.trendingUp, onClick: () => setAddTxnType('Income'), mobileOrder: 3 },
    { label: 'Transfer', icon: ICONS.swapHoriz, onClick: () => setAddTxnType('Transfer'), mobileOrder: 2 },
    { label: 'Statement', icon: ICONS.receiptLong, onClick: () => setShowStatementModal(true), mobileOrder: 4 },
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

      {/* Bento metric strip */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: 16 }}>
        <MetricChip
          label="Income · this month"
          value={`+${fmt(overview.monthIncome)}`}
          accent="var(--color-success)"
          icon={ICONS.trendingUp(16)}
        />
        <MetricChip
          label="Expense · this month"
          value={`−${fmt(overview.monthExpense)}`}
          accent="var(--color-danger)"
          icon={ICONS.trendingDown(16)}
        />
        <MetricChip
          label="Net · all wallets"
          value={fmt(overview.totalBalance)}
          accent="var(--color-primary)"
          icon={ICONS.wallet(16)}
        />
        <Card style={{ display: 'flex', alignItems: 'center', gap: 14, padding: 14, borderLeft: '4px solid var(--color-warning)', borderRadius: 'var(--radius-md)' }}>
          <BudgetGauge percent={overview.budgetUsage} size={64} />
          <div style={{ display: 'flex', flexDirection: 'column', gap: 4, minWidth: 0 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
              <span style={{ color: 'var(--color-warning)', display: 'flex' }}>{ICONS.pieChart(14)}</span>
              <Eyebrow>Budget usage</Eyebrow>
            </div>
            <span style={{ fontSize: 20, fontWeight: 700, color: 'var(--color-warning)', fontFamily: 'var(--font-display)', letterSpacing: '-0.02em' }}>
              {Math.round(overview.budgetUsage)}%
            </span>
          </div>
        </Card>
      </div>

      {/* Wallets carousel */}
      <div>
        <SectionHeader
          title="My Wallets"
          action={<Link to="/wallets" style={{ fontSize: 12, fontWeight: 600, color: 'var(--color-primary)', textDecoration: 'none' }}>Manage →</Link>}
        />
        <div className="wallet-carousel" style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(260px, 1fr))', gap: 16 }}>
          {(() => {
            const counts: Record<string, number> = {}
            return (wallets ?? []).slice(0, 4).map(w => {
              const variant = inferVariant(w.type, w.name, w.shortName)
              const idx = counts[variant] ?? 0
              counts[variant] = idx + 1
              return (
                <WalletCard
                  key={w.id}
                  variant={variant}
                  paletteIndex={idx}
                  name={w.name}
                  shortName={w.shortName}
                  balance={w.balance}
                  onClick={() => navigate('/wallets')}
                />
              )
            })
          })()}
        </div>
      </div>

      {/* Charts Row */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: 20 }}>
        <Card>
          <SectionHeader title="Expense by Category" />
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
          <SectionHeader title="Income vs Expense" accent="var(--color-success)" />
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
          <table style={{ width: '100%', minWidth: 640, borderCollapse: 'collapse', fontSize: 13 }}>
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
                    {formatDate(t.timestamp * 1000, { month: 'short', day: 'numeric' }, profile?.timezone)}
                  </td>
                  <td style={{ padding: '14px 24px' }}><Badge type={t.type as any} /></td>
                  <td style={{ padding: '14px 24px', fontWeight: 600, color: 'var(--color-text-primary)' }}>
                    {subcatMap.get(t.subcategoryId) || t.subcategoryId}
                  </td>
                  <td style={{
                    padding: '14px 24px', textAlign: 'right', fontWeight: 700, fontSize: 14,
                    color: t.type === 'Income' ? 'var(--color-success)' : t.type === 'Transfer' ? 'var(--color-primary)' : 'var(--color-danger)',
                    whiteSpace: 'nowrap',
                  }}>
                    {t.type === 'Income' ? '+' : t.type === 'Transfer' ? '' : '-'}{fmt(t.amount)}
                  </td>
                  <td style={{ padding: '14px 24px' }}>
                    <WalletFlow srcId={t.srcId} dstId={t.dstId} contactName={t.contactName} type={t.type as any} />
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
          onEdit={(t) => { setSelectedTxn(null); setEditTxn(t) }}
          onDelete={(t) => { setSelectedTxn(null); setDeleteTxn(t) }}
        />
      )}

      {deleteTxn && <DeleteTxnDialog txn={deleteTxn} onClose={() => setDeleteTxn(null)} />}

      {showStatementModal && <StatementModal onClose={() => setShowStatementModal(false)} />}

      {(addTxnType || editTxn) && (
        <TxnDialog
          txn={editTxn || undefined}
          initialType={addTxnType ?? undefined}
          onClose={() => { setAddTxnType(null); setEditTxn(null) }}
        />
      )}
    </div>
  )
}

function StatementModal({ onClose }: { onClose: () => void }) {
  const [startDate, setStartDate] = useState('')
  const [endDate, setEndDate] = useState('')

  const durations = [
    { label: 'This Week', value: 'one_week' },
    { label: 'This Month', value: 'this_month' },
    { label: 'Last 30 Days', value: 'one_month' },
    { label: 'This Year', value: 'this_year' },
    { label: 'All Time', value: 'all_time' },
  ]

  const handlePreview = (duration: string) => {
    window.open(`/statement?duration=${duration}`, '_blank')
    onClose()
  }

  const handleCustomPreview = () => {
    window.open(`/statement?start=${startDate}&end=${endDate}`, '_blank')
    onClose()
  }

  return (
    <Modal title="Generate Statement" onClose={onClose} width={760}>
      <div className="statement-modal-grid">
        
        {/* Left Column: Quick Select */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <h4 style={{ fontSize: 11, fontWeight: 700, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', margin: 0 }}>Quick Select</h4>
          <div className="quick-select-container">
            {durations.map(d => (
              <button
                key={d.value}
                className="quick-select-btn"
                onClick={() => handlePreview(d.value)}
                style={{
                  border: '1px solid var(--color-border)', background: 'var(--color-surface)',
                  cursor: 'pointer', fontWeight: 600, color: 'var(--color-text-primary)',
                  transition: 'all var(--transition-fast)', fontFamily: 'inherit',
                  display: 'flex', justifyContent: 'space-between', alignItems: 'center'
                }}
                onMouseEnter={e => { e.currentTarget.style.background = 'var(--color-primary-subtle)'; e.currentTarget.style.borderColor = 'var(--color-primary)' }}
                onMouseLeave={e => { e.currentTarget.style.background = 'var(--color-surface)'; e.currentTarget.style.borderColor = 'var(--color-border)' }}
              >
                {d.label}
                <svg className="quick-select-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" style={{ opacity: 0.5 }}>
                  <polyline points="9 18 15 12 9 6"></polyline>
                </svg>
              </button>
            ))}
          </div>
        </div>

        {/* Right Column: Custom Range */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <h4 style={{ fontSize: 11, fontWeight: 700, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', margin: 0 }}>Custom Range</h4>
          
          <DateRangePicker 
            startDate={startDate} 
            endDate={endDate} 
            onChange={(start, end) => { setStartDate(start); setEndDate(end) }} 
          />

          <button 
            onClick={handleCustomPreview} 
            disabled={!startDate || !endDate}
            style={{ 
              width: '100%', padding: '14px', 
              background: (!startDate || !endDate) ? 'var(--color-border)' : 'var(--color-primary)', 
              border: 'none', borderRadius: 12, 
              cursor: (!startDate || !endDate) ? 'not-allowed' : 'pointer', 
              fontWeight: 600, color: 'white', fontSize: 14, fontFamily: 'inherit',
              transition: 'background var(--transition-fast)',
              marginTop: 'auto'
            }}
          >
            Generate Custom Statement
          </button>
        </div>

      </div>
    </Modal>
  )
}

