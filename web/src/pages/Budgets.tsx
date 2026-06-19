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
import CategoryIcon, { categoryAccent } from '../components/ui/CategoryIcon'
import { ICONS } from '../components/ui/Icons'

export default function Budgets() {
  const { searchTerm } = useSearch()
  const { data: budgets, isLoading } = useBudgets()
  const { data: alerts } = useBudgetAlerts()
  const { data: categories } = useQuery({ queryKey: ['categories'], queryFn: () => listCategories() })
  const [showAdd, setShowAdd] = useState(false)
  const [editing, setEditing] = useState<import('../types').BudgetStatus | null>(null)

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
        <div style={{ display: 'grid', gridTemplateColumns: 'minmax(0, 2fr) minmax(0, 1fr)', gap: 18 }} className="budget-overview">
          <Card style={{ display: 'flex', flexDirection: 'column', gap: 10, padding: 22, borderLeft: '3px solid var(--color-primary)' }}>
            <Eyebrow color="var(--color-primary)">{monthLabel}</Eyebrow>
            <h3 style={{ fontSize: 22, fontWeight: 700, color: 'var(--color-text-primary)', margin: 0, fontFamily: 'var(--font-display)', letterSpacing: '-0.02em' }}>
              Monthly Overview
            </h3>
            <p style={{ fontSize: 13, color: 'var(--color-text-secondary)', margin: '2px 0 0', lineHeight: 1.5 }}>
              You've spent <strong style={{ color: 'var(--color-text-primary)' }}>{totals.percent.toFixed(0)}%</strong> of your monthly limit.
              {totals.remaining >= 0
                ? ` You're on track to save ${fmt(totals.remaining)}.`
                : ` You're over by ${fmt(Math.abs(totals.remaining))}.`}
            </p>
            <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginTop: 14, flexWrap: 'wrap' }}>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                <Eyebrow>Total spent</Eyebrow>
                <div style={{ display: 'flex', alignItems: 'baseline', gap: 8 }}>
                  <span style={{ fontSize: 28, fontWeight: 700, color: 'var(--color-text-primary)', fontFamily: 'var(--font-display)', letterSpacing: '-0.02em', lineHeight: 1 }}>
                    {fmt(totals.spent)}
                  </span>
                  <span style={{ fontSize: 13, color: 'var(--color-text-tertiary)' }}>/ {fmt(totals.limit)}</span>
                </div>
              </div>
              <div style={{ marginLeft: 'auto' }}>
                <Button onClick={() => setShowAdd(true)} icon={ICONS.savings(16)}>Set Budget</Button>
              </div>
            </div>
          </Card>

          <Card style={{
            background: 'var(--hero-gradient)', color: 'white', overflow: 'hidden',
            display: 'flex', flexDirection: 'column', justifyContent: 'center', alignItems: 'center',
            position: 'relative', gap: 6, padding: 22,
          }}>
            <Eyebrow color="rgba(255,255,255,0.85)">Remaining Balance</Eyebrow>
            <div style={{ width: 130, height: 130, display: 'flex', alignItems: 'center', justifyContent: 'center', marginTop: 2 }}>
              <BudgetGauge
                percent={Math.min(100, Math.max(0, totals.percent))}
                size={130}
                color="white"
                textColor="white"
                trackColor="rgba(255,255,255,0.22)"
              />
            </div>
            <span style={{ fontSize: 24, fontWeight: 700, fontFamily: 'var(--font-display)', letterSpacing: '-0.02em', lineHeight: 1 }}>
              {fmt(Math.abs(totals.remaining))}
            </span>
            <span style={{ fontSize: 11, opacity: 0.85, fontWeight: 600, marginTop: 2 }}>
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
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: 18 }}>
          {filteredBudgets.map(b => <BudgetCard key={b.categoryId} budget={b} onEdit={() => setEditing(b)} />)}
          <CreateBudgetCard onClick={() => setShowAdd(true)} />
        </div>
      )}

      {showAdd && <SetBudgetDialog categories={categories ?? []} onClose={() => setShowAdd(false)} />}
      {editing && <SetBudgetDialog categories={categories ?? []} existing={editing} onClose={() => setEditing(null)} />}
    </div>
  )
}

function BudgetCard({ budget, onEdit }: { budget: import('../types').BudgetStatus; onEdit: () => void }) {
  const del = useDeleteBudget()
  const catAccent = categoryAccent(budget.categoryId)
  const status = budget.percent >= 100
    ? { label: budget.percent > 100 ? 'Over limit' : 'Limit reached', color: 'var(--color-danger)' }
    : budget.percent > 80
      ? { label: 'Approaching limit', color: 'var(--color-warning)' }
      : budget.percent > 50
        ? { label: 'On track', color: 'var(--color-success)' }
        : { label: 'Plenty left', color: 'var(--color-success)' }

  return (
    <Card padding={0} style={{
      display: 'flex', flexDirection: 'column', overflow: 'hidden',
      background: `linear-gradient(135deg, ${catAccent}14 0%, ${catAccent}05 45%, var(--color-surface) 100%)`,
      borderColor: `${catAccent}33`,
    }}>
      <div style={{ padding: 18, flex: 1, display: 'flex', flexDirection: 'column', gap: 16 }}>
        <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12, minWidth: 0 }}>
            <div style={{
              width: 38, height: 38, borderRadius: 'var(--radius-md)',
              background: catAccent + '18', color: catAccent,
              display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0,
            }}>
              <CategoryIcon catId={budget.categoryId} size={18} />
            </div>
            <h3 style={{
              fontSize: 15, fontWeight: 700, color: 'var(--color-text-primary)', margin: 0,
              overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
            }}>{budget.categoryName}</h3>
          </div>
          <div style={{ display: 'flex', gap: 4, flexShrink: 0 }}>
            <button
              onClick={onEdit}
              aria-label="Edit budget"
              style={{ width: 28, height: 28, borderRadius: 'var(--radius-sm)', background: 'transparent', cursor: 'pointer', color: 'var(--color-text-tertiary)', display: 'flex', alignItems: 'center', justifyContent: 'center', transition: 'all var(--transition-fast)', border: 'none' }}
              onMouseEnter={e => { e.currentTarget.style.background = 'var(--color-primary-subtle)'; e.currentTarget.style.color = 'var(--color-primary)' }}
              onMouseLeave={e => { e.currentTarget.style.background = 'transparent'; e.currentTarget.style.color = 'var(--color-text-tertiary)' }}
            >{ICONS.edit(14)}</button>
            <button
              onClick={() => del.mutate(budget.categoryId)}
              aria-label="Delete budget"
              style={{ width: 28, height: 28, borderRadius: 'var(--radius-sm)', background: 'transparent', cursor: 'pointer', color: 'var(--color-text-tertiary)', display: 'flex', alignItems: 'center', justifyContent: 'center', transition: 'all var(--transition-fast)', border: 'none' }}
              onMouseEnter={e => { e.currentTarget.style.background = 'var(--color-danger-subtle)'; e.currentTarget.style.color = 'var(--color-danger)' }}
              onMouseLeave={e => { e.currentTarget.style.background = 'transparent'; e.currentTarget.style.color = 'var(--color-text-tertiary)' }}
            >{ICONS.trash(14)}</button>
          </div>
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
          <BudgetGauge percent={budget.percent} size={72} color={catAccent} endColor="#ef4444" />
          <div style={{ flex: 1, minWidth: 0 }}>
            <Eyebrow>Spent</Eyebrow>
            <p style={{
              fontSize: 18, fontWeight: 700, margin: '4px 0 0',
              color: 'var(--color-text-primary)', fontFamily: 'var(--font-display)',
              letterSpacing: '-0.02em',
            }}>{fmt(budget.spent)}</p>
            <p style={{ fontSize: 11, color: 'var(--color-text-tertiary)', margin: '4px 0 0', fontWeight: 500 }}>
              of {fmt(budget.amount)} limit
            </p>
          </div>
        </div>
      </div>

      <div style={{
        padding: '12px 18px', display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        borderTop: `1px solid ${catAccent}1f`,
        background: `color-mix(in srgb, var(--color-surface) 75%, ${catAccent} 25%)`,
      }}>
        <span style={{
          display: 'inline-flex', alignItems: 'center', gap: 6,
          padding: '3px 9px', borderRadius: 999,
          fontSize: 11, fontWeight: 700, letterSpacing: '0.03em',
          color: status.color,
          background: `color-mix(in srgb, ${status.color} 14%, transparent)`,
          border: `1px solid color-mix(in srgb, ${status.color} 28%, transparent)`,
        }}>
          <span style={{ width: 6, height: 6, borderRadius: 999, background: status.color }} />
          {status.label}
        </span>
        <span style={{
          fontSize: 12, fontWeight: 700,
          color: budget.remaining >= 0 ? 'var(--color-text-secondary)' : 'var(--color-danger)',
          fontFamily: 'var(--font-mono)',
        }}>
          {budget.remaining >= 0 ? `${fmt(budget.remaining)} left` : `Over by ${fmt(Math.abs(budget.remaining))}`}
        </span>
      </div>
    </Card>
  )
}

function CreateBudgetCard({ onClick }: { onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      style={{
        minHeight: 180,
        borderRadius: 16,
        border: '2px dashed var(--color-border)',
        background:
          'linear-gradient(135deg, var(--color-surface) 0%, var(--color-bg) 100%)',
        color: 'var(--color-text-tertiary)',
        position: 'relative',
        overflow: 'hidden',
        padding: '18px 20px',
        cursor: 'pointer',
        fontFamily: 'inherit',
        textAlign: 'left',
        transition: 'all var(--transition-fast)',
        display: 'flex',
        flexDirection: 'column',
      }}
      onMouseEnter={e => {
        e.currentTarget.style.borderColor = 'var(--color-primary)'
        e.currentTarget.style.color = 'var(--color-primary)'
        e.currentTarget.style.background =
          'linear-gradient(135deg, var(--color-primary-subtle) 0%, var(--color-surface) 100%)'
        e.currentTarget.style.transform = 'translateY(-4px)'
        e.currentTarget.style.boxShadow = '0 14px 32px rgba(0, 82, 204, 0.16)'
      }}
      onMouseLeave={e => {
        e.currentTarget.style.borderColor = 'var(--color-border)'
        e.currentTarget.style.color = 'var(--color-text-tertiary)'
        e.currentTarget.style.background =
          'linear-gradient(135deg, var(--color-surface) 0%, var(--color-bg) 100%)'
        e.currentTarget.style.transform = 'translateY(0)'
        e.currentTarget.style.boxShadow = 'none'
      }}
    >
      {/* Faint decorative blobs */}
      <span aria-hidden style={{
        position: 'absolute', top: -28, right: -28, width: 120, height: 120,
        borderRadius: '50%', background: 'currentColor', opacity: 0.04,
      }} />
      <span aria-hidden style={{
        position: 'absolute', bottom: -36, left: -16, width: 100, height: 100,
        borderRadius: '50%', background: 'currentColor', opacity: 0.03,
      }} />

      {/* Top row — eyebrow + dashed category chip */}
      <div style={{ position: 'relative', display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <span style={{
          fontFamily: 'var(--font-display)', fontWeight: 700, fontSize: 12, letterSpacing: '0.18em', opacity: 0.55,
        }}>NEW BUDGET</span>
        <span style={{
          padding: '3px 10px', borderRadius: 999, fontSize: 9, fontWeight: 700,
          letterSpacing: '0.08em', border: '1px dashed currentColor', opacity: 0.55,
        }}>CATEGORY</span>
      </div>

      {/* Center plus medallion */}
      <div style={{ position: 'relative', flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <span style={{
          width: 52, height: 52, borderRadius: '50%',
          border: '2px dashed currentColor',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          background: 'color-mix(in srgb, var(--color-surface) 70%, transparent)',
          backdropFilter: 'blur(2px)',
          transition: 'all var(--transition-fast)',
        }}>
          <svg width={22} height={22} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.5} strokeLinecap="round">
            <line x1="12" y1="5" x2="12" y2="19" />
            <line x1="5" y1="12" x2="19" y2="12" />
          </svg>
        </span>
      </div>

      {/* Bottom — title + subtitle */}
      <div style={{ position: 'relative', display: 'flex', flexDirection: 'column', gap: 4 }}>
        <span style={{ fontSize: 14, fontWeight: 700, letterSpacing: '-0.01em' }}>Create Budget</span>
        <span style={{ fontSize: 10, fontWeight: 600, letterSpacing: '0.06em', textTransform: 'uppercase', opacity: 0.7 }}>
          Add a category or goal
        </span>
      </div>
    </button>
  )
}

function SetBudgetDialog({ categories, existing, onClose }: { categories: import('../types').TxnCategory[], existing?: import('../types').BudgetStatus, onClose: () => void }) {
  const setBudget = useSetBudget()
  const isEdit = !!existing
  const [categoryId, setCategoryId] = useState(existing?.categoryId ?? '')
  const [amount, setAmount] = useState(existing ? String(existing.amount) : '')
  const [alertAt, setAlertAt] = useState(existing ? String(existing.alertAt) : '80')

  return (
    <Modal title={isEdit ? 'Edit Budget' : 'Set Budget'} onClose={onClose} width={460}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <Select label="Category" value={categoryId} onChange={e => setCategoryId(e.target.value)} disabled={isEdit}
          options={[{ value: '', label: 'Overall Budget' }, ...categories.map(c => ({ value: c.id, label: c.name }))]} />
        <Input label="Monthly Limit" type="number" placeholder="0.00" value={amount} onChange={e => setAmount(e.target.value)} />
        <Input label="Alert Threshold (%)" type="number" placeholder="80" value={alertAt} onChange={e => setAlertAt(e.target.value)} />
        <div style={{ display: 'flex', gap: 12, justifyContent: 'flex-end', marginTop: 12 }}>
          <Button variant="secondary" onClick={onClose}>Cancel</Button>
          <Button onClick={() => setBudget.mutate({ categoryId, amount: parseFloat(amount), alertAt: parseInt(alertAt) }, { onSuccess: onClose })} disabled={!amount} style={{ padding: '12px 32px' }}>
            {isEdit ? 'Update Budget' : 'Save Budget'}
          </Button>
        </div>
      </div>
    </Modal>
  )
}
