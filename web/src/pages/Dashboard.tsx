import { useMemo, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getChartData, listCategories, listSubcategories } from '../api/endpoints'
import { useTransactions } from '../hooks/useTransactions'
import { fmt } from '../lib/formatter'

import TopBar from '../components/layout/TopBar'
import Card from '../components/ui/Card'
import StatCard from '../components/ui/StatCard'
import Badge from '../components/ui/Badge'
import Button from '../components/ui/Button'
import Modal from '../components/ui/Modal'
import MiniDonut from '../components/charts/MiniDonut'
import MiniBarChart from '../components/charts/MiniBarChart'
import BudgetGauge from '../components/charts/BudgetGauge'
import { ICONS } from '../components/ui/Icons'

export default function Dashboard() {
  const navigate = useNavigate()
  const { data: charts, isLoading: isChartsLoading } = useQuery({
    queryKey: ['chartData'],
    queryFn: () => getChartData(),
  })
  const { data: resp } = useTransactions()
  const txns = resp?.data ?? []
  const { data: allCategories, isLoading: isCatsLoading } = useQuery({ queryKey: ['categories'], queryFn: () => listCategories() })
  const { data: subcategories, isLoading: isSubsLoading } = useQuery({ queryKey: ['subcategories'], queryFn: () => listSubcategories() })
  const [showStatementModal, setShowStatementModal] = useState(false)

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
    if (categorySpends.length <= 5) {
      return categorySpends.map((c, i) => ({
        value: c.amount,
        color: i === 0 ? 'var(--color-primary)' : i === 1 ? '#f59e0b' : i === 2 ? 'var(--color-danger)' : i === 3 ? '#8b5cf6' : 'var(--color-text-tertiary)',
        name: catMap.get(c.categoryId) || c.categoryName || c.categoryId
      }))
    }

    const top4 = categorySpends.slice(0, 4).map((c, i) => ({
      value: c.amount,
      color: i === 0 ? 'var(--color-primary)' : i === 1 ? '#f59e0b' : i === 2 ? 'var(--color-danger)' : i === 3 ? '#8b5cf6' : 'var(--color-text-tertiary)',
      name: catMap.get(c.categoryId) || c.categoryName || c.categoryId
    }))

    const othersAmount = categorySpends.slice(4).reduce((sum, c) => sum + c.amount, 0)
    top4.push({
      value: othersAmount,
      color: 'var(--color-text-tertiary)',
      name: 'Others'
    })

    return top4
  }, [charts?.categories, catMap])

  const comparisonData = useMemo(() => {
    return (charts?.comparison || []).map(d => ({
      label: d.month,
      income: d.income,
      expense: d.expense
    }))
  }, [charts?.comparison])

  if (isLoading) return <p className="text-gray-500">Loading...</p>
  if (!charts) return null

  const overview = charts.overview
  const recentTxns = [...txns].sort((a, b) => b.timestamp - a.timestamp).slice(0, 5)

  const hour = new Date().getHours()
  const greeting = hour < 12 ? 'Good morning' : hour < 17 ? 'Good afternoon' : 'Good evening'

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
      <TopBar title="Dashboard" subtitle="Summary of your financial activity" />

      {/* Greeting + Quick Actions */}
      <div style={{
        background: `linear-gradient(135deg, var(--color-primary) 0%, oklch(0.55 0.14 165) 100%)`,
        borderRadius: 20, padding: '28px 32px', color: 'white', position: 'relative', overflow: 'hidden',
      }}>
        <div style={{ position: 'absolute', top: -40, right: -20, width: 200, height: 200, borderRadius: '50%', background: 'rgba(255,255,255,0.08)' }} />
        <div style={{ position: 'absolute', bottom: -60, right: 80, width: 150, height: 150, borderRadius: '50%', background: 'rgba(255,255,255,0.05)' }} />
        <div style={{ position: 'relative', zIndex: 1 }}>
          <p style={{ fontSize: 14, opacity: 0.8, margin: 0 }}>{greeting}, Tracker</p>
          <h2 style={{ fontSize: 22, fontWeight: 700, margin: '4px 0 16px', letterSpacing: '-0.01em' }}>
            Current Balance is {fmt(overview.totalBalance)}
          </h2>
          <div style={{ display: 'flex', gap: 10, flexWrap: 'wrap' }}>
            {[
              { label: 'Add Expense', icon: ICONS.arrowDown, onClick: () => navigate('/transactions?add=Expense') },
              { label: 'Add Income', icon: ICONS.arrowUp, onClick: () => navigate('/transactions?add=Income') },
              { label: 'Transfer', icon: ICONS.transfer, onClick: () => navigate('/transactions?add=Transfer') },
              { label: 'Statement', icon: ICONS.file, onClick: () => setShowStatementModal(true) },
            ].map(a => (
              <button 
                key={a.label} 
                onClick={a.onClick}
                style={{
                  display: 'flex', alignItems: 'center', gap: 6, padding: '8px 14px', borderRadius: 10,
                  background: 'rgba(255,255,255,0.15)', backdropFilter: 'blur(10px)',
                  border: '1px solid rgba(255,255,255,0.2)', color: 'white', fontSize: 12, fontWeight: 600,
                  cursor: 'pointer', transition: 'all 0.15s',
                }}
                onMouseEnter={e => e.currentTarget.style.background = 'rgba(255,255,255,0.25)'}
                onMouseLeave={e => e.currentTarget.style.background = 'rgba(255,255,255,0.15)'}
              >
                {a.icon(14)} {a.label}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Stat Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: 16 }}>
        <StatCard label="Total Balance" value={fmt(overview.totalBalance)} icon={ICONS.money(18)} accentColor="var(--color-primary)" />
        <StatCard label="Month Income" value={fmt(overview.monthIncome)} icon={ICONS.arrowUp(18)} accentColor="var(--color-success)" />
        <StatCard label="Month Expense" value={fmt(overview.monthExpense)} icon={ICONS.arrowDown(18)} accentColor="var(--color-danger)" />
        <Card style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 8 }}>
          <span style={{ fontSize: 12, fontWeight: 600, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.06em' }}>Budget Usage</span>
          <BudgetGauge percent={overview.budgetUsage} size={110} />
        </Card>
      </div>

      {/* Charts Row */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: 16 }}>
        <Card>
          <h3 style={{ fontSize: 15, fontWeight: 700, color: 'var(--color-text-primary)', margin: '0 0 20px', display: 'flex', alignItems: 'center', gap: 8 }}>
            <span style={{ width: 3, height: 18, borderRadius: 2, background: 'var(--color-primary)' }} />
            Expense by Category
          </h3>
          <div style={{ display: 'flex', alignItems: 'center', gap: 24 }}>
            <MiniDonut segments={chartCategories} size={130} />
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8, flex: 1 }}>
              {chartCategories.map(c => (
                <div key={c.name} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <span style={{ width: 8, height: 8, borderRadius: 3, background: c.color, flexShrink: 0 }} />
                  <span style={{ fontSize: 12, color: 'var(--color-text-secondary)', flex: 1 }}>{c.name}</span>
                  <span style={{ fontSize: 12, fontWeight: 600, color: 'var(--color-text-primary)' }}>{fmt(c.value)}</span>
                </div>
              ))}
            </div>
          </div>
        </Card>
        <Card>
          <h3 style={{ fontSize: 15, fontWeight: 700, color: 'var(--color-text-primary)', margin: '0 0 20px', display: 'flex', alignItems: 'center', gap: 8 }}>
            <span style={{ width: 3, height: 18, borderRadius: 2, background: 'var(--color-success)' }} />
            Income vs Expense
          </h3>
          <div style={{ display: 'flex', justifyContent: 'center' }}>
            <MiniBarChart data={comparisonData} size={{ w: 280, h: 130 }} />
          </div>
          <div style={{ display: 'flex', gap: 16, justifyContent: 'center', marginTop: 8 }}>
            <span style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 11, color: 'var(--color-text-tertiary)' }}>
              <span style={{ width: 8, height: 8, borderRadius: 3, background: 'var(--color-success)' }} /> Income
            </span>
            <span style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 11, color: 'var(--color-text-tertiary)' }}>
              <span style={{ width: 8, height: 8, borderRadius: 3, background: 'var(--color-danger)' }} /> Expense
            </span>
          </div>
        </Card>
      </div>

      {/* Recent Transactions */}
      <Card padding={0}>
        <div style={{ padding: '20px 24px', borderBottom: '1px solid var(--color-border)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <h3 style={{ fontSize: 15, fontWeight: 700, color: 'var(--color-text-primary)', margin: 0 }}>Recent Transactions</h3>
          <Link to="/transactions" style={{ fontSize: 12, fontWeight: 600, color: 'var(--color-primary)', textDecoration: 'none' }}>View All</Link>
        </div>
        <div style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 13 }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--color-border)' }}>
                {['Date', 'Type', 'Category', 'Amount', 'Wallet'].map(h => (
                  <th key={h} style={{ padding: '12px 24px', textAlign: h === 'Amount' ? 'right' : 'left', fontSize: 10, fontWeight: 700, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.08em' }}>{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {recentTxns.map(t => (
                <tr key={t.id} style={{ borderBottom: '1px solid var(--color-border)' }}
                  className="hover-row transition-colors">
                  <td style={{ padding: '14px 24px', color: 'var(--color-text-tertiary)', fontSize: 12, fontWeight: 600 }}>
                    {new Date(t.timestamp * 1000).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
                  </td>
                  <td style={{ padding: '14px 24px' }}>
                    <Badge type={t.type as any} />
                  </td>
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
                    <span style={{ fontSize: 11, fontWeight: 700, background: 'var(--color-bg)', padding: '3px 8px', borderRadius: 6, color: 'var(--color-text-tertiary)' }}>
                      {t.srcId || t.dstId || t.contactName || '—'}
                    </span>
                  </td>
                </tr>
              ))}
              {recentTxns.length === 0 && (
                <tr>
                  <td colSpan={5} style={{ padding: '40px', textAlign: 'center', color: 'var(--color-text-tertiary)' }}>
                    No recent transactions found
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </Card>

      {showStatementModal && (
        <StatementModal onClose={() => setShowStatementModal(false)} />
      )}
    </div>
  )
}

function StatementModal({ onClose }: { onClose: () => void }) {
  const durations = [
    { label: 'This Week', value: 'one_week' },
    { label: 'Current Month', value: 'this_month' },
    { label: 'Last 30 Days', value: 'one_month' },
    { label: 'Last 6 Months', value: 'half_year' },
    { label: 'Current Year', value: 'this_year' },
    { label: 'All Time', value: 'all-time' },
  ]

  const handlePreview = (duration: string) => {
    window.open(`/statement?duration=${duration}`, '_blank')
    onClose()
  }

  return (
    <Modal title="Generate Statement" onClose={onClose} width={400}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <p style={{ fontSize: 14, color: 'var(--color-text-secondary)', marginBottom: 8 }}>
          Select a time period for your financial statement preview.
        </p>
        {durations.map(d => (
          <Button 
            key={d.value} 
            variant="secondary" 
            onClick={() => handlePreview(d.value)}
            style={{ justifyContent: 'flex-start', padding: '14px 20px' }}
          >
            {d.label}
          </Button>
        ))}
      </div>
    </Modal>
  )
}
