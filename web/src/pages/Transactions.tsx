import { formatDate } from '../lib/formatter'
import { useState, useMemo, useEffect, useRef } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useTransactions } from '../hooks/useTransactions'
import { useSearch } from '../context/SearchContext'
import { useWallets } from '../hooks/useWallets'
import { useContacts } from '../hooks/useContacts'
import { useQuery } from '@tanstack/react-query'
import { listCategories, listSubcategories , getProfile } from '../api/endpoints'
import type { Transaction } from '../types'
import { fmt } from '../lib/formatter'

import TopBar from '../components/layout/TopBar'
import Card from '../components/ui/Card'
import Badge from '../components/ui/Badge'
import Button from '../components/ui/Button'
import ActionButton from '../components/ui/ActionButton'
import DeleteTxnDialog from '../components/ui/DeleteTxnDialog'
import { ICONS } from '../components/ui/Icons'
import WalletFlow from '../components/ui/WalletFlow'
import TransactionDetails from '../components/ui/TransactionDetails'
import MetricChip from '../components/ui/MetricChip'
import TxnDialog, { TXN_TYPE_OPTIONS, TxnType } from '../components/ui/TxnDialog'
import { useFitPageSize } from '../hooks/useFitPageSize'

const typeOptions = TXN_TYPE_OPTIONS

export default function Transactions() {
  const { data: profile } = useQuery({ queryKey: ['profile'], queryFn: getProfile })
  const [searchParams, setSearchParams] = useSearchParams()
  const { searchTerm } = useSearch()
  const { data: resp, isLoading } = useTransactions()
  const txns = resp?.data ?? []
  const { data: wallets } = useWallets()
  const { data: contacts } = useContacts()
  const { data: categories } = useQuery({ queryKey: ['categories'], queryFn: () => listCategories() })
  const { data: subcategories } = useQuery({ queryKey: ['subcategories'], queryFn: () => listSubcategories() })

  const [filterType, setFilterType] = useState<string>('')
  const [showAdd, setShowAdd] = useState(false)
  const [initialType, setInitialType] = useState<TxnType | undefined>()
  const [initialContact, setInitialContact] = useState<string | undefined>()
  const [editTxn, setEditTxn] = useState<Transaction | null>(null)
  const [deleteTxn, setDeleteTxn] = useState<Transaction | null>(null)
  const [selectedTxn, setSelectedTxn] = useState<Transaction | null>(null)
  const [page, setPage] = useState(0)
  const filterBarRef = useRef<HTMLDivElement>(null)
  const firstCardRef = useRef<HTMLButtonElement>(null)
  const firstRowRef = useRef<HTMLTableRowElement>(null)
  const tableWrapRef = useRef<HTMLDivElement>(null)
  const paginationRef = useRef<HTMLDivElement>(null)
  const pageSize = useFitPageSize({ topRef: filterBarRef, cardRef: firstCardRef, rowRef: firstRowRef, wrapRef: tableWrapRef, paginationRef })

  useEffect(() => {
    const addType = searchParams.get('add') as TxnType
    if (addType && typeOptions.includes(addType)) {
      setInitialType(addType)
      const c = searchParams.get('contact')
      if (c) setInitialContact(c)
      setShowAdd(true)
      searchParams.delete('add')
      searchParams.delete('contact')
      setSearchParams(searchParams, { replace: true })
    }

    const editId = searchParams.get('edit')
    if (editId) {
      const txn = txns.find(t => t.id === parseInt(editId))
      if (txn) {
        setEditTxn(txn)
        searchParams.delete('edit')
        setSearchParams(searchParams, { replace: true })
      }
    }

    const showId = searchParams.get('show')
    if (showId) {
      const txn = txns.find(t => t.id === parseInt(showId))
      if (txn) {
        setSelectedTxn(txn)
        searchParams.delete('show')
        setSearchParams(searchParams, { replace: true })
      }
    }
  }, [searchParams, setSearchParams, txns])

  const subcatMap = useMemo(() => {
    const m = new Map<string, string>()
    subcategories?.forEach(s => m.set(s.id, s.name))
    return m
  }, [subcategories])

  if (isLoading) return <p style={{ color: 'var(--color-text-tertiary)', padding: 40 }}>Loading...</p>

  const filtered = txns
    .filter(t => {
      const matchesType = !filterType || t.type === filterType
      const matchesSearch = !searchTerm ||
        t.remarks?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        t.subcategoryId.toLowerCase().includes(searchTerm.toLowerCase()) ||
        subcatMap.get(t.subcategoryId)?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        t.amount.toString().includes(searchTerm) ||
        t.contactName?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        t.srcId?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        t.dstId?.toLowerCase().includes(searchTerm.toLowerCase())
      return matchesType && matchesSearch
    })
    .sort((a, b) => b.timestamp - a.timestamp)

  const totalPages = Math.max(1, Math.ceil(filtered.length / pageSize))
  const safePage = Math.min(page, totalPages - 1)
  const paginated = filtered.slice(safePage * pageSize, (safePage + 1) * pageSize)

  const totals = {
    income: txns.filter(t => t.type === 'Income').reduce((s, t) => s + t.amount, 0),
    expense: txns.filter(t => t.type === 'Expense').reduce((s, t) => s + t.amount, 0),
    transfers: txns.filter(t => t.type === 'Transfer').reduce((s, t) => s + t.amount, 0),
  }

  return (
    <div className="txn-page" style={{ display: 'flex', flexDirection: 'column', gap: 28 }}>
      <TopBar title="Transactions" subtitle="Detailed history of your financial movements" />

      {/* Summary chips */}
      <div className="txn-metrics" style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: 16 }}>
        <MetricChip
          label="Total Income"
          value={`+${fmt(totals.income)}`}
          accent="var(--color-success)"
          icon={ICONS.trendingUp(16)}
          hint={`${txns.filter(t => t.type === 'Income').length} transactions`}
        />
        <MetricChip
          label="Total Expense"
          value={`−${fmt(totals.expense)}`}
          accent="var(--color-danger)"
          icon={ICONS.trendingDown(16)}
          hint={`${txns.filter(t => t.type === 'Expense').length} transactions`}
        />
        <MetricChip
          label="Transfers"
          value={fmt(totals.transfers)}
          accent="var(--color-primary)"
          icon={ICONS.swapHoriz(16)}
          hint={`${txns.filter(t => t.type === 'Transfer').length} transactions`}
        />
      </div>

      {/* Filter Bar — sticks to the top on mobile so it never scrolls away */}
      <div ref={filterBarRef} className="txn-filter-bar" style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 12 }}>
        <div style={{ display: 'flex', background: 'var(--color-surface)', border: '1px solid var(--color-border)', borderRadius: 'var(--radius-md)', padding: 4, gap: 2 }}>
          {['', ...typeOptions].map(f => (
            <button
              key={f || 'all'}
              onClick={() => { setFilterType(f); setPage(0) }}
              style={{
                padding: '8px 20px', borderRadius: 'var(--radius-sm)', fontSize: 13, fontWeight: 600,
                border: 'none', cursor: 'pointer', transition: 'all var(--transition-fast)',
                background: filterType === f ? 'var(--color-primary)' : 'transparent',
                color: filterType === f ? 'white' : 'var(--color-text-tertiary)',
                fontFamily: 'inherit',
              }}>
              {f || 'All'}
            </button>
          ))}
        </div>
        {/* Hidden on mobile — the floating + button covers this there */}
        <span className="hidden md:inline-flex">
          <Button onClick={() => setShowAdd(true)} icon={ICONS.addCircle(16)}>Add Transaction</Button>
        </span>
      </div>

      {/* Table */}
      <Card padding={0}>
        {/* Desktop: zebra table with tinted sticky header; frame is bounded to the
            measured height so the header/pagination stay fixed (page doesn't scroll) */}
        <div ref={tableWrapRef} className="zebra-table-wrap hidden md:block">
          <table className="zebra-table" style={{ width: '100%', minWidth: 720, tableLayout: 'fixed', borderCollapse: 'collapse', fontSize: 13 }}>
            <thead>
              <tr>
                {[
                  { h: 'Date', cls: '', w: 96 },
                  { h: 'Type', cls: '', w: 120 },
                  { h: 'Category', cls: '', w: undefined },
                  { h: 'Amount', cls: '', w: 130 },
                  { h: 'Wallet', cls: '', w: 200 },
                  { h: 'Remarks', cls: 'hidden lg:table-cell', w: undefined },
                  { h: '', cls: '', w: 96 },
                ].map(({ h, cls, w }) => (
                  <th key={h || 'actions'} className={cls} style={{ width: w, padding: '14px 24px', textAlign: h === 'Amount' ? 'right' : h === 'Wallet' ? 'center' : 'left', fontSize: 10, fontWeight: 700, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.08em' }}>{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {paginated.map((t, i) => (
                <tr key={t.id} ref={i === 0 ? firstRowRef : undefined} style={{ borderBottom: '1px solid var(--color-border)' }}
                  onClick={() => setSelectedTxn(t)}>
                  <td style={{ padding: '14px 24px', color: 'var(--color-text-tertiary)', fontSize: 12, fontWeight: 600, whiteSpace: 'nowrap' }}>
                    {formatDate(t.timestamp * 1000, { month: 'short', day: 'numeric' }, profile?.timezone)}
                  </td>
                  <td style={{ padding: '14px 24px' }}><Badge type={t.type as any} /></td>
                  <td style={{ padding: '14px 24px', fontWeight: 600, color: 'var(--color-text-primary)' }}>
                    {subcatMap.get(t.subcategoryId) || t.subcategoryId}
                  </td>
                  <td style={{
                    padding: '14px 24px', textAlign: 'right', fontWeight: 700, fontSize: 14,
                    fontFamily: 'var(--font-mono)', whiteSpace: 'nowrap',
                    color: t.type === 'Income' ? 'var(--color-success)' : t.type === 'Transfer' ? 'var(--color-primary)' : 'var(--color-danger)',
                  }}>
                    {t.type === 'Income' ? '+' : t.type === 'Transfer' ? '' : '−'}{fmt(t.amount)}
                  </td>
                  <td style={{ padding: '14px 24px' }}>
                    <WalletFlow srcId={t.srcId} dstId={t.dstId} contactName={t.contactName} type={t.type as any} />
                  </td>
                  <td className="hidden lg:table-cell" style={{ padding: '14px 24px', color: 'var(--color-text-tertiary)', fontSize: 12, fontStyle: 'italic', maxWidth: 120, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {t.remarks || '—'}
                  </td>
                  <td style={{ padding: '14px 24px', textAlign: 'right' }}>
                    <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
                      <ActionButton
                        actionType="edit"
                        icon={ICONS.edit(14)}
                        onClick={e => { e.stopPropagation(); setEditTxn(t) }}
                      />
                      <ActionButton
                        actionType="delete"
                        icon={ICONS.trash(14)}
                        onClick={e => { e.stopPropagation(); setDeleteTxn(t) }}
                      />
                    </div>
                  </td>
                </tr>
              ))}
              {filtered.length === 0 && (
                <tr><td colSpan={7} style={{ padding: 60, textAlign: 'center', color: 'var(--color-text-tertiary)' }}>No transactions found</td></tr>
              )}
            </tbody>
          </table>
        </div>

        {/* Mobile: stacked, tappable summary cards (natural height, vertical scroll) */}
        <div className="txn-card-list flex flex-col md:hidden">
          {paginated.map((t, i) => (
            <button key={t.id} ref={i === 0 ? firstCardRef : undefined} className="txn-card" onClick={() => setSelectedTxn(t)}>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 10 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, minWidth: 0 }}>
                  <Badge type={t.type as any} />
                  <span style={{ fontWeight: 700, fontSize: 14, color: 'var(--color-text-primary)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {subcatMap.get(t.subcategoryId) || t.subcategoryId}
                  </span>
                </div>
                <span style={{
                  fontWeight: 700, fontSize: 15, fontFamily: 'var(--font-mono)', whiteSpace: 'nowrap', flexShrink: 0,
                  color: t.type === 'Income' ? 'var(--color-success)' : t.type === 'Transfer' ? 'var(--color-primary)' : 'var(--color-danger)',
                }}>
                  {t.type === 'Income' ? '+' : t.type === 'Transfer' ? '' : '−'}{fmt(t.amount)}
                </span>
              </div>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 10 }}>
                <div style={{ minWidth: 0, overflow: 'hidden' }}>
                  <WalletFlow srcId={t.srcId} dstId={t.dstId} contactName={t.contactName} type={t.type as any} />
                </div>
                <span style={{ fontSize: 12, color: 'var(--color-text-tertiary)', fontWeight: 600, whiteSpace: 'nowrap', flexShrink: 0 }}>
                  {formatDate(t.timestamp * 1000, { month: 'short', day: 'numeric' }, profile?.timezone)}
                </span>
              </div>
            </button>
          ))}
          {filtered.length === 0 && (
            <div style={{ padding: 40, textAlign: 'center', color: 'var(--color-text-tertiary)' }}>No transactions found</div>
          )}
        </div>

        {/* Pagination — measured as the bottom boundary of the row area */}
        {totalPages > 1 && (
          <div ref={paginationRef} style={{ padding: '16px 24px', borderTop: '1px solid var(--color-border)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <p style={{ fontSize: 12, color: 'var(--color-text-tertiary)', fontWeight: 500 }}>
              Showing {safePage * pageSize + 1}–{Math.min((safePage + 1) * pageSize, filtered.length)} of {filtered.length}
            </p>
            <div style={{ display: 'flex', gap: 6 }}>
              <button
                disabled={safePage === 0}
                onClick={() => setPage(safePage - 1)}
                style={{
                  padding: '8px 14px', borderRadius: 'var(--radius-sm)', fontSize: 12, fontWeight: 600,
                  border: '1px solid var(--color-border)', background: 'var(--color-surface)',
                  cursor: safePage === 0 ? 'not-allowed' : 'pointer', color: 'var(--color-text-secondary)',
                  opacity: safePage === 0 ? 0.5 : 1, fontFamily: 'inherit',
                }}
              >Previous</button>
              <button
                disabled={safePage >= totalPages - 1}
                onClick={() => setPage(safePage + 1)}
                style={{
                  padding: '8px 14px', borderRadius: 'var(--radius-sm)', fontSize: 12, fontWeight: 600,
                  border: '1px solid var(--color-border)', background: 'var(--color-surface)',
                  cursor: safePage >= totalPages - 1 ? 'not-allowed' : 'pointer', color: 'var(--color-text-secondary)',
                  opacity: safePage >= totalPages - 1 ? 0.5 : 1, fontFamily: 'inherit',
                }}
              >Next</button>
            </div>
          </div>
        )}
      </Card>

      {/* Transaction Detail Slide-in */}
      {selectedTxn && (
        <TransactionDetails
          txn={selectedTxn}
          wallets={wallets ?? []}
          contacts={contacts ?? []}
          categories={categories ?? []}
          subcategories={subcategories ?? []}
          onClose={() => setSelectedTxn(null)}
          onEdit={(t) => { setSelectedTxn(null); setEditTxn(t) }}
          onDelete={(t) => { setSelectedTxn(null); setDeleteTxn(t) }}
        />
      )}

      {(showAdd || editTxn) && (
        <TxnDialog
          txn={editTxn || undefined}
          initialType={initialType}
          initialContact={initialContact}
          onClose={() => { setShowAdd(false); setEditTxn(null); setInitialType(undefined); setInitialContact(undefined) }}
        />
      )}
      {deleteTxn && <DeleteTxnDialog txn={deleteTxn} onClose={() => setDeleteTxn(null)} />}
    </div>
  )
}

