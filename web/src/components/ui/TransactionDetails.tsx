import { useEffect } from 'react'
import { Transaction, Wallet, Contact, TxnCategory, TxnSubcategory } from '../../types'
import { fmt } from '../../lib/formatter'
import { ICONS } from './Icons'
import Badge from './Badge'

interface TransactionDetailsProps {
  txn: Transaction
  wallets: Wallet[]
  contacts: Contact[]
  categories: TxnCategory[]
  subcategories: TxnSubcategory[]
  onClose: () => void
  onEdit?: (txn: Transaction) => void
  onDelete?: (txn: Transaction) => void
}

export default function TransactionDetails({
  txn,
  wallets,
  contacts,
  categories,
  subcategories,
  onClose,
  onEdit,
  onDelete,
}: TransactionDetailsProps) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onClose])

  const txnDate = new Date(txn.timestamp * 1000)
  const dateStr = txnDate.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
  const timeStr = txnDate.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })

  const accentColor = txn.type === 'Income' ? 'var(--color-success)' : txn.type === 'Transfer' ? 'var(--color-primary)' : 'var(--color-danger)'
  const heroBg = txn.type === 'Income' ? 'var(--color-success-subtle)' : txn.type === 'Transfer' ? 'var(--color-primary-subtle)' : 'var(--color-danger-subtle)'

  const TypeIcon = txn.type === 'Income' ? ICONS.arrowUp : txn.type === 'Transfer' ? ICONS.transfer : ICONS.arrowDown
  const sign = txn.type === 'Income' ? '+' : txn.type === 'Transfer' ? '' : '-'

  const catMap = new Map(categories.map(c => [c.id, c.name]))
  const subcatMap = new Map(subcategories.map(s => [s.id, s.name]))
  const walletMap = new Map(wallets.map(w => [w.shortName, w.name]))
  const contactMap = new Map(contacts.map(c => [c.nickName, c.fullName]))

  const catId = txn.subcategoryId?.split('-')[0] ?? ''
  type PartyKind = 'wallet' | 'contact' | 'unknown'
  const resolveParty = (id: string): { label: string; kind: PartyKind } | null => {
    if (!id) return null
    const w = walletMap.get(id)
    if (w) return { label: w, kind: 'wallet' }
    const c = contactMap.get(id)
    if (c) return { label: c || id, kind: 'contact' }
    return { label: id, kind: 'unknown' }
  }
  const from = resolveParty(txn.srcId)
  const to = resolveParty(txn.dstId)
  const personLabel = contactMap.get(txn.contactName) || txn.contactName
  const personIsDuplicate = !!personLabel && (personLabel === from?.label || personLabel === to?.label)
  const categoryName = catMap.get(catId) || catId
  const subcategoryName = subcatMap.get(txn.subcategoryId) || txn.subcategoryId

  return (
    <>
      <div
        onClick={onClose}
        style={{
          position: 'fixed', inset: 0,
          background: 'rgba(9, 9, 11, 0.5)',
          backdropFilter: 'blur(6px)',
          zIndex: 300,
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          padding: 16,
          animation: 'fadeIn 0.2s ease-out',
        }}
      >
        <div
          onClick={e => e.stopPropagation()}
          role="dialog"
          aria-modal="true"
          style={{
            width: 'min(560px, 100%)',
            maxHeight: '88vh',
            background: 'var(--color-surface)',
            borderRadius: 20,
            boxShadow: '0 24px 64px rgba(0,0,0,0.18)',
            display: 'flex', flexDirection: 'column',
            overflow: 'hidden',
            animation: 'modalIn 0.25s cubic-bezier(0.4, 0, 0.2, 1)',
          }}
        >
          <style>{`
            @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
            @keyframes modalIn { from { opacity: 0; transform: translateY(8px) scale(0.98); } to { opacity: 1; transform: translateY(0) scale(1); } }
            .td-action-btn:hover { filter: brightness(0.95); transform: translateY(-1px); }
            .td-action-btn:active { transform: translateY(0); }
            .td-close:hover { background: var(--color-bg) !important; color: var(--color-text-primary) !important; }
          `}</style>

          {/* Header */}
          <div style={{ padding: '14px 20px', display: 'flex', alignItems: 'center', justifyContent: 'space-between', borderBottom: '1px solid var(--color-border)' }}>
            <span style={{ fontSize: 12, fontWeight: 700, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.08em' }}>Transaction Details</span>
            <button
              onClick={onClose}
              className="td-close"
              aria-label="Close"
              style={{
                width: 30, height: 30, borderRadius: 8, border: 'none',
                background: 'transparent', cursor: 'pointer', display: 'flex',
                alignItems: 'center', justifyContent: 'center', color: 'var(--color-text-tertiary)',
                transition: 'all 0.15s',
              }}
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
            </button>
          </div>

          {/* Body */}
          <div style={{ padding: 20, overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: 14 }}>
            {/* Hero card */}
            <div style={{
              background: heroBg,
              borderRadius: 16,
              padding: '22px 20px',
              display: 'flex', alignItems: 'center', gap: 18,
              border: `1px solid ${accentColor}22`,
            }}>
              <div style={{
                width: 56, height: 56, borderRadius: 16,
                background: 'var(--color-surface)',
                color: accentColor,
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                boxShadow: '0 4px 12px rgba(0,0,0,0.06)',
                flexShrink: 0,
              }}>
                {TypeIcon(28)}
              </div>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ fontSize: 30, fontWeight: 800, color: accentColor, letterSpacing: '-0.02em', lineHeight: 1.1 }}>
                  {sign}{fmt(txn.amount)}
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginTop: 8, flexWrap: 'wrap' }}>
                  <Badge type={txn.type as any} />
                  <span style={{ fontSize: 12, color: 'var(--color-text-secondary)', fontWeight: 600 }}>
                    {dateStr}
                    <span style={{ color: 'var(--color-text-tertiary)', margin: '0 6px' }}>·</span>
                    {timeStr}
                  </span>
                </div>
              </div>
            </div>

            {/* Classification card */}
            <div style={{
              background: 'var(--color-bg)',
              borderRadius: 14,
              padding: 14,
              border: '1px solid var(--color-border)',
            }}>
              <SectionLabel icon={ICONS.dashboard(13)} text="Classification" />
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10, marginTop: 10 }}>
                <Field label="Category" value={categoryName} />
                <Field label="Subcategory" value={subcategoryName} />
              </div>
            </div>

            {/* Movement card */}
            {(from || to || (personLabel && !personIsDuplicate)) && (
              <div style={{
                background: 'var(--color-surface)',
                borderRadius: 14,
                padding: 14,
                border: '1px solid var(--color-border)',
                borderLeft: `3px solid ${accentColor}`,
              }}>
                <SectionLabel icon={ICONS.transfer(13)} text="Movement" />
                {(from || to) && (
                  <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginTop: 12 }}>
                    {from && <PartyField label="From" party={from} />}
                    {from && to && (
                      <div style={{ color: 'var(--color-text-tertiary)', fontWeight: 700, fontSize: 16, flexShrink: 0 }}>→</div>
                    )}
                    {to && <PartyField label="To" party={to} />}
                  </div>
                )}
                {personLabel && !personIsDuplicate && (
                  <div style={{
                    marginTop: from || to ? 12 : 0,
                    paddingTop: from || to ? 12 : 0,
                    borderTop: from || to ? '1px dashed var(--color-border)' : 'none',
                  }}>
                    <Field label="Person" value={personLabel} icon={ICONS.user(13)} />
                  </div>
                )}
              </div>
            )}

            {/* Remarks card */}
            <div style={{
              padding: 14,
              borderRadius: 14,
              background: 'rgba(0,0,0,0.02)',
              border: '1px dashed var(--color-border)',
            }}>
              <SectionLabel icon={ICONS.file(13)} text="Remarks" />
              <div style={{
                fontSize: 13,
                color: txn.remarks ? 'var(--color-text-primary)' : 'var(--color-text-tertiary)',
                lineHeight: 1.5,
                fontStyle: txn.remarks ? 'normal' : 'italic',
                marginTop: 8,
                whiteSpace: 'pre-wrap',
                wordBreak: 'break-word',
              }}>
                {txn.remarks || 'No remarks provided'}
              </div>
            </div>
          </div>

          {/* Actions */}
          <div style={{ padding: '14px 20px', borderTop: '1px solid var(--color-border)', display: 'flex', gap: 10, background: 'var(--color-surface)' }}>
            <button
              className="td-action-btn"
              onClick={() => onEdit?.(txn)}
              style={{
                flex: 1, height: 44, borderRadius: 12, border: 'none',
                background: 'var(--color-primary)', color: 'white', fontWeight: 700,
                fontSize: 14, cursor: 'pointer', display: 'flex', alignItems: 'center',
                justifyContent: 'center', gap: 8, transition: 'all 0.2s', fontFamily: 'inherit',
              }}
            >
              {ICONS.edit(16)} Edit
            </button>
            <button
              className="td-action-btn"
              onClick={() => onDelete?.(txn)}
              aria-label="Delete"
              style={{
                width: 44, height: 44, borderRadius: 12, border: '1px solid var(--color-border)',
                background: 'var(--color-surface)', color: 'var(--color-danger)', cursor: 'pointer',
                display: 'flex', alignItems: 'center', justifyContent: 'center', transition: 'all 0.2s',
              }}
            >
              {ICONS.trash(18)}
            </button>
          </div>
        </div>
      </div>
    </>
  )
}

function SectionLabel({ icon, text }: { icon: React.ReactNode; text: string }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 6, color: 'var(--color-text-tertiary)' }}>
      {icon}
      <span style={{ fontSize: 10, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.08em' }}>{text}</span>
    </div>
  )
}

function PartyField({ label, party }: { label: string; party: { label: string; kind: 'wallet' | 'contact' | 'unknown' } }) {
  const isContact = party.kind === 'contact'
  const tone = isContact ? 'var(--color-primary)' : 'var(--color-text-tertiary)'
  const icon = isContact ? ICONS.user(13) : ICONS.wallet(13)
  return (
    <div style={{ flex: 1, minWidth: 0 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 5, color: 'var(--color-text-tertiary)' }}>
        <span style={{ display: 'flex' }}>{icon}</span>
        <span style={{ fontSize: 10, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em' }}>{label}</span>
      </div>
      <div style={{
        fontSize: 13, fontWeight: 600,
        color: 'var(--color-text-primary)',
        marginTop: 4,
        display: 'flex', alignItems: 'center', gap: 6,
        minWidth: 0,
      }}>
        <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{party.label}</span>
        {isContact && (
          <span style={{
            flexShrink: 0,
            fontSize: 9, fontWeight: 700, letterSpacing: '0.06em', textTransform: 'uppercase',
            padding: '2px 7px', borderRadius: 999,
            background: 'var(--color-primary-subtle)', color: tone,
          }}>
            Contact
          </span>
        )}
      </div>
    </div>
  )
}

function Field({ label, value, icon }: { label: string; value: string; icon?: React.ReactNode }) {
  return (
    <div style={{ flex: 1, minWidth: 0 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 5, color: 'var(--color-text-tertiary)' }}>
        {icon}
        <span style={{ fontSize: 10, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em' }}>{label}</span>
      </div>
      <div style={{
        fontSize: 13, fontWeight: 600, color: 'var(--color-text-primary)',
        marginTop: 4, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis',
      }}>
        {value}
      </div>
    </div>
  )
}
