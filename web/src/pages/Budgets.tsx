import { useState, useMemo } from 'react'
import { useBudgets, useBudgetAlerts, useSetBudget, useDeleteBudget } from '../hooks/useBudgets'
import { useQuery } from '@tanstack/react-query'
import { listCategories } from '../api/endpoints'
import { useSearch } from '../context/SearchContext'
import { fmt } from '../lib/formatter'

import TopBar from '../components/layout/TopBar'
import Card from '../components/ui/Card'
import Button from '../components/ui/Button'
import Modal from '../components/ui/Modal'
import Input from '../components/ui/Input'
import Select from '../components/ui/Select'
import BudgetGauge from '../components/charts/BudgetGauge'
import SectionHeader from '../components/ui/SectionHeader'
import Eyebrow from '../components/ui/Eyebrow'
import { ICONS } from '../components/ui/Icons'

export default function Budgets() {
  const { searchTerm } = useSearch()
  const { data: budgets, isLoading } = useBudgets()
  const { data: alerts } = useBudgetAlerts()
  const { data: categories } = useQuery({ queryKey: ['categories'], queryFn: listCategories })
  const [showAdd, setShowAdd] = useState(false)

  const filteredBudgets = useMemo(() =>
    (budgets ?? []).filter(b => !searchTerm || b.categoryName.toLowerCase().includes(searchTerm.toLowerCase())),
    [budgets, searchTerm])

  const totals = useMemo(() => {
    const list = budgets ?? []
    const spent = list.reduce((s, b) => s + b.spent, 0)
    const limit = list.reduce((s, b) => s + b.amount, 0)
    const remaining = limit - spent
    const percent = limit > 0 ? (spent / limit) * 100 : 0
    return { spent, limit, remaining, percent }
  }, [budgets])

  if (isLoading) return <p style={{ color: 'var(--color-text-tertiary)', padding: 40 }}>Loading...</p>

  const monthLabel = new Date().toLocaleDateString('en-US', { month: 'long', year: 'numeric' })
  const daysLeft = (() => {
    const now = new Date()
    const last = new Date(now.getFullYear(), now.getMonth() + 1, 0).getDate()
    return last - now.getDate()
  })()

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 28 }}>
      <TopBar title="Budgets" subtitle="Plan your monthly spending limits" />

      {/* Monthly Overview band */}
      {(budgets?.length ?? 0) > 0 && (
        <div style={{ display: 'grid', gridTemplateColumns: 'minmax(0, 2fr) minmax(0, 1fr)', gap: 20 }} className="budget-overview">
          <Card style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <Eyebrow color="var(--color-primary)">{monthLabel}</Eyebrow>
            <h3 style={{ fontSize: 22, fontWeight: 700, color: 'var(--color-text-primary)', margin: 0, fontFamily: 'var(--font-display)', letterSpacing: '-0.02em' }}>
              Monthly Overview
            </h3>
            <p style={{ fontSize: 13, color: 'var(--color-text-secondary)', margin: 0, lineHeight: 1.55 }}>
              You've spent <strong style={{ color: 'var(--color-text-primary)' }}>{totals.percent.toFixed(0)}%</strong> of your monthly limit.
              {totals.remaining >= 0
                ? ` You're on track to save ${fmt(totals.remaining)}.`
                : ` You're over by ${fmt(Math.abs(totals.remaining))}.`}
            </p>
            <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginTop: 4 }}>
              <Eyebrow>Total spent</Eyebrow>
              <span style={{ fontSize: 26, fontWeight: 700, color: 'var(--color-text-primary)', fontFamily: 'var(--font-display)', letterSpacing: '-0.02em' }}>
                {fmt(totals.spent)}
              </span>
              <span style={{ fontSize: 13, color: 'var(--color-text-tertiary)' }}>/ {fmt(totals.limit)}</span>
              <div style={{ marginLeft: 'auto' }}>
                <Button onClick={() => setShowAdd(true)} icon={ICONS.savings(16)}>Set Budget</Button>
              </div>
            </div>
          </Card>

          <Card style={{
            background: 'var(--hero-gradient)', color: 'white', overflow: 'hidden',
            display: 'flex', flexDirection: 'column', justifyContent: 'center', alignItems: 'center',
            position: 'relative', gap: 8, padding: 24,
          }}>
            <Eyebrow color="rgba(255,255,255,0.85)">Remaining Balance</Eyebrow>
            <div style={{ width: 130, height: 130, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <BudgetGauge percent={Math.min(100, Math.max(0, totals.percent))} size={130} />
            </div>
            <span style={{ fontSize: 24, fontWeight: 700, fontFamily: 'var(--font-display)', letterSpacing: '-0.02em' }}>
              {fmt(Math.abs(totals.remaining))}
            </span>
            <span style={{ fontSize: 11, opacity: 0.85, fontWeight: 600 }}>
              {(100 - totals.percent).toFixed(0)}% left · {daysLeft}d remaining
            </span>
          </Card>
        </div>
      )}

      <SectionHeader
        title="Category Budgets"
        action={
          (budgets?.length ?? 0) === 0 ? (
            <Button onClick={() => setShowAdd(true)} icon={ICONS.plus(16)}>Set Budget</Button>
          ) : null
        }
      />

      {/* Alerts */}
      {alerts && alerts.length > 0 && !searchTerm && (
        <div style={{ background: 'var(--color-danger-subtle)', borderRadius: 'var(--radius-xl)', padding: 24, border: '1px solid var(--color-danger)' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 16 }}>
            <span style={{ color: 'var(--color-danger)' }}>{ICONS.alert(20)}</span>
            <h3 style={{ fontSize: 13, fontWeight: 700, color: 'var(--color-danger)', textTransform: 'uppercase', letterSpacing: '0.06em' }}>Spending Alerts</h3>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(240px, 1fr))', gap: 12 }}>
            {alerts.map(a => (
              <div key={a.categoryId} style={{ background: 'var(--color-surface)', padding: 16, borderRadius: 'var(--radius-lg)', border: '1px solid var(--color-border)' }}>
                <p style={{ fontSize: 13, fontWeight: 700, color: 'var(--color-text-primary)' }}>{a.categoryName}</p>
                <p style={{ fontSize: 11, fontWeight: 600, color: 'var(--color-danger)', textTransform: 'uppercase', marginTop: 4 }}>
                  {a.percent.toFixed(0)}% used • {(a.budgetAmount - a.spent) < 0 ? `Over by ${fmt(Math.abs(a.budgetAmount - a.spent))}` : `${fmt(a.budgetAmount - a.spent)} left`}
                </p>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Budget cards */}
      {filteredBudgets.length === 0 ? (
        <Card style={{ padding: 60, textAlign: 'center', border: '2px dashed var(--color-border)' }}>
          <p style={{ color: 'var(--color-text-tertiary)', fontWeight: 600 }}>
            {searchTerm ? 'No budgets match your search' : 'No budgets set. Create one to start tracking your goals.'}
          </p>
        </Card>
      ) : (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))', gap: 20 }}>
          {filteredBudgets.map(b => <BudgetCard key={b.categoryId} budget={b} />)}
        </div>
      )}

      {showAdd && <SetBudgetDialog categories={categories ?? []} onClose={() => setShowAdd(false)} />}
    </div>
  )
}

function BudgetCard({ budget }: { budget: import('../types').BudgetStatus }) {
  const del = useDeleteBudget()
  const accentColor = budget.percent > 100 ? 'var(--color-danger)' : budget.percent > 80 ? 'var(--color-warning)' : 'var(--color-success)'

  return (
    <Card padding={0} style={{ display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <div style={{ padding: 24, flex: 1, display: 'flex', flexDirection: 'column', gap: 20 }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <div style={{
              width: 42, height: 42, borderRadius: 'var(--radius-md)',
              background: accentColor + '15', color: accentColor,
              display: 'flex', alignItems: 'center', justifyContent: 'center',
            }}>{ICONS.pieChart(20)}</div>
            <h3 style={{ fontSize: 17, fontWeight: 700, color: 'var(--color-text-primary)' }}>{budget.categoryName}</h3>
          </div>
          <button
            onClick={() => del.mutate(budget.categoryId)}
            style={{ width: 32, height: 32, borderRadius: 'var(--radius-sm)', background: 'var(--color-surface)', cursor: 'pointer', color: 'var(--color-text-tertiary)', display: 'flex', alignItems: 'center', justifyContent: 'center', transition: 'all var(--transition-fast)', border: '1px solid var(--color-border)' }}
            onMouseEnter={e => e.currentTarget.style.background = 'var(--color-danger-subtle)'}
            onMouseLeave={e => e.currentTarget.style.background = 'var(--color-surface)'}
          >{ICONS.trash(16)}</button>
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: 24 }}>
          <BudgetGauge percent={budget.percent} size={100} />
          <div style={{ flex: 1 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
              <span style={{ fontSize: 11, fontWeight: 700, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Usage</span>
              <span style={{ fontSize: 14, fontWeight: 800, color: accentColor }}>{budget.percent.toFixed(0)}%</span>
            </div>
            <div style={{ height: 8, background: 'var(--color-bg)', borderRadius: 4, overflow: 'hidden' }}>
              <div style={{ height: '100%', width: `${Math.min(budget.percent, 100)}%`, background: accentColor, borderRadius: 4, transition: 'width 0.5s ease' }} />
            </div>
          </div>
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
          <div style={{ background: 'var(--color-bg)', padding: '12px 16px', borderRadius: 'var(--radius-md)' }}>
            <p style={{ fontSize: 10, fontWeight: 700, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.06em', marginBottom: 4 }}>Spent</p>
            <p style={{ fontSize: 16, fontWeight: 800, color: 'var(--color-text-primary)', fontFamily: "var(--font-display)" }}>{fmt(budget.spent)}</p>
          </div>
          <div style={{ background: 'var(--color-bg)', padding: '12px 16px', borderRadius: 'var(--radius-md)' }}>
            <p style={{ fontSize: 10, fontWeight: 700, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.06em', marginBottom: 4 }}>Limit</p>
            <p style={{ fontSize: 16, fontWeight: 800, color: 'var(--color-text-primary)', fontFamily: "var(--font-display)" }}>{fmt(budget.amount)}</p>
          </div>
        </div>
      </div>

      <div style={{
        padding: '14px 24px', display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        background: budget.remaining >= 0 ? 'var(--color-success-subtle)' : 'var(--color-danger-subtle)',
        borderTop: '1px solid var(--color-border)',
      }}>
        <span style={{ fontSize: 10, fontWeight: 800, textTransform: 'uppercase', color: budget.remaining >= 0 ? 'var(--color-success)' : 'var(--color-danger)', letterSpacing: '0.05em' }}>
          {budget.remaining >= 0 ? 'Available Balance' : 'Budget Exceeded'}
        </span>
        <span style={{ fontSize: 15, fontWeight: 800, color: budget.remaining >= 0 ? 'var(--color-success)' : 'var(--color-danger)', fontFamily: "var(--font-display)" }}>
          {fmt(Math.abs(budget.remaining))}
        </span>
      </div>
    </Card>
  )
}

function SetBudgetDialog({ categories, onClose }: { categories: import('../types').TxnCategory[], onClose: () => void }) {
  const setBudget = useSetBudget()
  const [categoryId, setCategoryId] = useState('')
  const [amount, setAmount] = useState('')
  const [alertAt, setAlertAt] = useState('80')

  return (
    <Modal title="Set Budget" onClose={onClose} width={460}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <Select label="Category" value={categoryId} onChange={e => setCategoryId(e.target.value)}
          options={[{ value: '', label: 'Overall Budget' }, ...categories.map(c => ({ value: c.id, label: c.name }))]} />
        <Input label="Monthly Limit" type="number" placeholder="0.00" value={amount} onChange={e => setAmount(e.target.value)} />
        <Input label="Alert Threshold (%)" type="number" placeholder="80" value={alertAt} onChange={e => setAlertAt(e.target.value)} />
        <div style={{ display: 'flex', gap: 12, justifyContent: 'flex-end', marginTop: 12 }}>
          <Button variant="secondary" onClick={onClose}>Cancel</Button>
          <Button onClick={() => setBudget.mutate({ categoryId, amount: parseFloat(amount), alertAt: parseInt(alertAt) }, { onSuccess: onClose })} disabled={!amount} style={{ padding: '12px 32px' }}>Save Budget</Button>
        </div>
      </div>
    </Modal>
  )
}
