import { useState, useMemo, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useTransactions, useDeleteTransaction } from '../hooks/useTransactions'
import { useSearch } from '../context/SearchContext'
import { useWallets } from '../hooks/useWallets'
import { useContacts } from '../hooks/useContacts'
import { useQuery } from '@tanstack/react-query'
import { listCategories, listSubcategories } from '../api/endpoints'
import type { Transaction } from '../types'
import { fmt } from '../lib/formatter'

import TopBar from '../components/layout/TopBar'
import Card from '../components/ui/Card'
import Badge from '../components/ui/Badge'
import Button from '../components/ui/Button'
import Modal from '../components/ui/Modal'
import { ICONS } from '../components/ui/Icons'
import WalletFlow from '../components/ui/WalletFlow'
import TransactionDetails from '../components/ui/TransactionDetails'
import MetricChip from '../components/ui/MetricChip'
import TxnDialog, { TXN_TYPE_OPTIONS, TxnType } from '../components/ui/TxnDialog'

const typeOptions = TXN_TYPE_OPTIONS
const PAGE_SIZE = 15

export default function Transactions() {
  const [searchParams, setSearchParams] = useSearchParams()
  const { searchTerm } = useSearch()
  const { data: resp, isLoading } = useTransactions()
  const txns = resp?.data ?? []
  const { data: wallets } = useWallets()
  const { data: contacts } = useContacts()
  const { data: categories } = useQuery({ queryKey: ['categories'], queryFn: listCategories })
  const { data: subcategories } = useQuery({ queryKey: ['subcategories'], queryFn: () => listSubcategories() })

  const [filterType, setFilterType] = useState<string>('')
  const [showAdd, setShowAdd] = useState(false)
  const [initialType, setInitialType] = useState<TxnType | undefined>()
  const [initialContact, setInitialContact] = useState<string | undefined>()
  const [editTxn, setEditTxn] = useState<Transaction | null>(null)
  const [deleteTxn, setDeleteTxn] = useState<Transaction | null>(null)
  const [selectedTxn, setSelectedTxn] = useState<Transaction | null>(null)
  const [page, setPage] = useState(0)

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

  const totalPages = Math.ceil(filtered.length / PAGE_SIZE)
  const paginated = filtered.slice(page * PAGE_SIZE, (page + 1) * PAGE_SIZE)

  const totals = {
    income: txns.filter(t => t.type === 'Income').reduce((s, t) => s + t.amount, 0),
    expense: txns.filter(t => t.type === 'Expense').reduce((s, t) => s + t.amount, 0),
    transfers: txns.filter(t => t.type === 'Transfer').reduce((s, t) => s + t.amount, 0),
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 28 }}>
      <TopBar title="Transactions" subtitle="Detailed history of your financial movements" />

      {/* Summary chips */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: 16 }}>
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

      {/* Filter Bar */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 12 }}>
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
        <Button onClick={() => setShowAdd(true)} icon={ICONS.addCircle(16)}>Add Transaction</Button>
      </div>

      {/* Table */}
      <Card padding={0}>
        <div style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 13 }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--color-border)' }}>
                {[
                  { h: 'Date', cls: '' },
                  { h: 'Type', cls: '' },
                  { h: 'Category', cls: '' },
                  { h: 'Amount', cls: '' },
                  { h: 'Wallets', cls: 'hidden md:table-cell' },
                  { h: 'Remarks', cls: 'hidden lg:table-cell' },
                  { h: '', cls: '' },
                ].map(({ h, cls }) => (
                  <th key={h || 'actions'} className={cls} style={{ padding: '14px 24px', textAlign: h === 'Amount' ? 'right' : h === 'Wallets' ? 'center' : 'left', fontSize: 10, fontWeight: 700, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.08em' }}>{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {paginated.map(t => (
                <tr key={t.id} style={{ borderBottom: '1px solid var(--color-border)' }}
                  className="hover-row transition-colors"
                  onClick={() => setSelectedTxn(t)}>
                  <td style={{ padding: '14px 24px', color: 'var(--color-text-tertiary)', fontSize: 12, fontWeight: 600, whiteSpace: 'nowrap' }}>
                    {new Date(t.timestamp * 1000).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
                  </td>
                  <td style={{ padding: '14px 24px' }}><Badge type={t.type as any} /></td>
                  <td style={{ padding: '14px 24px', fontWeight: 600, color: 'var(--color-text-primary)' }}>
                    {subcatMap.get(t.subcategoryId) || t.subcategoryId}
                  </td>
                  <td style={{
                    padding: '14px 24px', textAlign: 'right', fontWeight: 700, fontSize: 14,
                    fontFamily: 'var(--font-mono)',
                    color: t.type === 'Income' ? 'var(--color-success)' : t.type === 'Transfer' ? 'var(--color-primary)' : 'var(--color-danger)',
                  }}>
                    {t.type === 'Income' ? '+' : t.type === 'Transfer' ? '' : '−'}{fmt(t.amount)}
                  </td>
                  <td className="hidden md:table-cell" style={{ padding: '14px 24px' }}>
                    <WalletFlow srcId={t.srcId} dstId={t.dstId} contactName={t.contactName} type={t.type as any} />
                  </td>
                  <td className="hidden lg:table-cell" style={{ padding: '14px 24px', color: 'var(--color-text-tertiary)', fontSize: 12, fontStyle: 'italic', maxWidth: 120, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {t.remarks || '—'}
                  </td>
                  <td style={{ padding: '14px 24px', textAlign: 'right' }}>
                    <div style={{ display: 'flex', gap: 4, justifyContent: 'flex-end' }}>
                      <button
                        onClick={e => { e.stopPropagation(); setEditTxn(t) }}
                        style={{ width: 32, height: 32, borderRadius: 'var(--radius-sm)', border: 'none', background: 'transparent', cursor: 'pointer', color: 'var(--color-primary)', display: 'flex', alignItems: 'center', justifyContent: 'center', transition: 'all var(--transition-fast)' }}
                        onMouseEnter={e => e.currentTarget.style.background = 'var(--color-primary-subtle)'}
                        onMouseLeave={e => e.currentTarget.style.background = 'transparent'}
                      >{ICONS.edit(14)}</button>
                      <button
                        onClick={e => { e.stopPropagation(); setDeleteTxn(t) }}
                        style={{ width: 32, height: 32, borderRadius: 'var(--radius-sm)', border: 'none', background: 'transparent', cursor: 'pointer', color: 'var(--color-danger)', display: 'flex', alignItems: 'center', justifyContent: 'center', transition: 'all var(--transition-fast)' }}
                        onMouseEnter={e => e.currentTarget.style.background = 'var(--color-danger-subtle)'}
                        onMouseLeave={e => e.currentTarget.style.background = 'transparent'}
                      >{ICONS.trash(14)}</button>
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

        {/* Pagination */}
        {totalPages > 1 && (
          <div style={{ padding: '16px 24px', borderTop: '1px solid var(--color-border)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <p style={{ fontSize: 12, color: 'var(--color-text-tertiary)', fontWeight: 500 }}>
              Showing {page * PAGE_SIZE + 1}–{Math.min((page + 1) * PAGE_SIZE, filtered.length)} of {filtered.length}
            </p>
            <div style={{ display: 'flex', gap: 6 }}>
              <button
                disabled={page === 0}
                onClick={() => setPage(p => p - 1)}
                style={{
                  padding: '8px 14px', borderRadius: 'var(--radius-sm)', fontSize: 12, fontWeight: 600,
                  border: '1px solid var(--color-border)', background: 'var(--color-surface)',
                  cursor: page === 0 ? 'not-allowed' : 'pointer', color: 'var(--color-text-secondary)',
                  opacity: page === 0 ? 0.5 : 1, fontFamily: 'inherit',
                }}
              >Previous</button>
              <button
                disabled={page >= totalPages - 1}
                onClick={() => setPage(p => p + 1)}
                style={{
                  padding: '8px 14px', borderRadius: 'var(--radius-sm)', fontSize: 12, fontWeight: 600,
                  border: '1px solid var(--color-border)', background: 'var(--color-surface)',
                  cursor: page >= totalPages - 1 ? 'not-allowed' : 'pointer', color: 'var(--color-text-secondary)',
                  opacity: page >= totalPages - 1 ? 0.5 : 1, fontFamily: 'inherit',
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
      {deleteTxn && <DeleteDialog txn={deleteTxn} onClose={() => setDeleteTxn(null)} />}
    </div>
  )
}


function DeleteDialog({ txn, onClose }: { txn: Transaction; onClose: () => void }) {
  const del = useDeleteTransaction()
  return (
    <Modal onClose={onClose} width={400}>
      <div style={{ textAlign: 'center' }}>
        <div style={{ width: 64, height: 64, background: 'var(--color-danger-subtle)', color: 'var(--color-danger)', borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 20px' }}>
          {ICONS.trash(32)}
        </div>
        <h2 style={{ fontSize: 20, fontWeight: 700, color: 'var(--color-text-primary)', margin: '0 0 10px' }}>Delete Transaction?</h2>
        <p style={{ fontSize: 14, color: 'var(--color-text-tertiary)', lineHeight: 1.6, margin: '0 0 32px' }}>
          Are you sure you want to delete this <span style={{ fontWeight: 700, color: 'var(--color-text-secondary)' }}>{txn.type}</span> for <span style={{ fontWeight: 700, color: 'var(--color-text-primary)' }}>{fmt(txn.amount)}</span>? This cannot be undone.
        </p>
        <div style={{ display: 'flex', gap: 12, justifyContent: 'center' }}>
          <Button variant="secondary" onClick={onClose} style={{ padding: '12px 24px' }}>Cancel</Button>
          <Button variant="danger" onClick={() => del.mutate(txn.id, { onSuccess: onClose })} style={{ padding: '12px 24px' }}>Confirm Delete</Button>
        </div>
      </div>
    </Modal>
  )
}
