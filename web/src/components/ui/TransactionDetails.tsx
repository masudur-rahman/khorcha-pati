import { Transaction, Wallet, Contact, TxnCategory, TxnSubcategory } from '../../types'
import { fmt } from '../../lib/formatter'
import { ICONS } from './Icons'
import CategoryIcon, { categoryAccent } from './CategoryIcon'
import ActionButton from './ActionButton'
import Modal from './Modal'

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
  const txnDate = new Date(txn.timestamp * 1000)
  const dateStr = txnDate.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
  const timeStr = txnDate.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })

  const accentColor = txn.type === 'Income' ? 'var(--color-success)' : txn.type === 'Transfer' ? 'var(--color-primary)' : 'var(--color-danger)'
  const heroGradient = txn.type === 'Income'
    ? 'linear-gradient(135deg, #006844 0%, #2BAE66 55%, #36B37E 100%)'
    : txn.type === 'Transfer'
      ? 'var(--hero-gradient)'
      : 'var(--hero-gradient-danger)'

  const TypeIcon = txn.type === 'Income' ? ICONS.trendingUp : txn.type === 'Transfer' ? ICONS.swapHoriz : ICONS.trendingDown
  const sign = txn.type === 'Income' ? '+' : txn.type === 'Transfer' ? '' : '-'

  const catMap = new Map(categories.map(c => [c.id, c.name]))
  const subcatMap = new Map(subcategories.map(s => [s.id, s.name]))
  const walletMap = new Map(wallets.map(w => [w.shortName, w.name]))
  const contactMap = new Map(contacts.map(c => [c.nickName, c.fullName]))

  const catId = txn.subcategoryId?.split('-')[0] ?? ''
  type PartyKind = 'wallet' | 'contact' | 'income' | 'expense'
  type Party = { label: string; kind: PartyKind }
  const resolveParty = (id: string): Party | null => {
    if (!id) return null
    const w = walletMap.get(id)
    if (w) return { label: w, kind: 'wallet' }
    const c = contactMap.get(id)
    return { label: c || id, kind: 'contact' }
  }
  let from = resolveParty(txn.srcId)
  let to = resolveParty(txn.dstId)
  // Use contactName as the missing-side contact when it aligns with the txn direction.
  if (txn.contactName) {
    const personParty: Party = { label: contactMap.get(txn.contactName) || txn.contactName, kind: 'contact' }
    if (!to && txn.type === 'Expense') to = personParty
    else if (!from && txn.type === 'Income') from = personParty
  }
  // Fill the still-missing side with a soft "world" node so every txn reads
  // from → to (destination wallet always right, source always left).
  if (!from && txn.type === 'Income') from = { label: 'Income', kind: 'income' }
  if (!to && txn.type === 'Expense') to = { label: 'Expense', kind: 'expense' }
  const categoryName = catMap.get(catId) || catId
  const subcategoryName = subcatMap.get(txn.subcategoryId) || txn.subcategoryId

  return (
    <Modal
      title="Transaction Details"
      onClose={onClose}
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            {/* Hero card — gradient with glass actions bottom-right (mirrors WalletCard) */}
            <div style={{
              position: 'relative',
              background: heroGradient,
              borderRadius: 16,
              padding: '20px 20px 22px',
              color: 'white',
              overflow: 'hidden',
              boxShadow: '0 10px 30px rgba(23,43,77,0.14)',
              minHeight: 132,
              display: 'flex', flexDirection: 'column', gap: 16,
            }}>
              <span aria-hidden style={{ position: 'absolute', top: -40, right: -40, width: 150, height: 150, borderRadius: '50%', background: 'rgba(255,255,255,0.10)' }} />
              <span aria-hidden style={{ position: 'absolute', bottom: -56, right: 20, width: 130, height: 130, borderRadius: '50%', background: 'rgba(255,255,255,0.06)' }} />

              <div style={{ position: 'relative', display: 'flex', alignItems: 'center', gap: 14 }}>
                <div style={{
                  width: 48, height: 48, borderRadius: 14,
                  background: 'rgba(255,255,255,0.20)', backdropFilter: 'blur(4px)',
                  border: '1px solid rgba(255,255,255,0.28)',
                  display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0,
                }}>
                  {TypeIcon(24)}
                </div>
                <span style={{
                  padding: '3px 11px', borderRadius: 999, fontSize: 10, fontWeight: 700,
                  letterSpacing: '0.08em', textTransform: 'uppercase',
                  background: 'rgba(255,255,255,0.22)', border: '1px solid rgba(255,255,255,0.28)',
                  backdropFilter: 'blur(4px)',
                }}>{txn.type}</span>
              </div>

              <div style={{ position: 'relative' }}>
                <div style={{
                  fontFamily: 'var(--font-display)', fontSize: 32, fontWeight: 800,
                  letterSpacing: '-0.02em', lineHeight: 1.05, textShadow: '0 1px 4px rgba(0,0,0,0.18)',
                }}>
                  {sign}{fmt(txn.amount)}
                </div>
                <div style={{ fontSize: 12, fontWeight: 600, opacity: 0.92, marginTop: 6 }}>
                  {dateStr}
                  <span style={{ opacity: 0.7, margin: '0 6px' }}>·</span>
                  {timeStr}
                </div>
              </div>

              <div style={{ position: 'absolute', bottom: 14, right: 14, display: 'flex', gap: 6 }}>
                <ActionButton actionType="edit" variant="glass" icon={ICONS.edit(14)} onClick={() => onEdit?.(txn)} title="Edit Transaction" />
                <ActionButton actionType="delete" variant="glass" icon={ICONS.trash(14)} onClick={() => onDelete?.(txn)} title="Delete Transaction" />
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
              <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginTop: 12 }}>
                <div style={{
                  width: 40, height: 40, borderRadius: 10,
                  background: categoryAccent(catId) + '1A', color: categoryAccent(catId),
                  display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0,
                }}>
                  <CategoryIcon catId={catId} size={20} />
                </div>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10, flex: 1, minWidth: 0 }}>
                  <Field label="Category" value={categoryName} />
                  <Field label="Subcategory" value={subcategoryName} />
                </div>
              </div>
            </div>

            {/* Movement card */}
            {(from || to) && (
              <div style={{
                background: 'var(--color-surface)',
                borderRadius: 14,
                padding: '12px 14px',
                border: '1px solid var(--color-border)',
                borderLeft: `3px solid ${accentColor}`,
              }}>
                <div className="movement-grid" style={{
                  display: 'flex',
                  gap: 12, alignItems: 'center', justifyContent: 'center'
                }}>
                  {from && <MovementCell role="Source" party={from} />}
                  {from && to && (
                    <div className="movement-arrow" style={{ color: 'var(--color-text-tertiary)', fontWeight: 700, fontSize: 16, lineHeight: 1 }}>→</div>
                  )}
                  {to && <MovementCell role="Destination" party={to} />}
                </div>
              </div>
            )}

            {/* Remarks card */}
            <div style={{
              padding: '12px 14px',
              borderRadius: 14,
              background: 'rgba(0,0,0,0.02)',
              border: '1px dashed var(--color-border)',
            }}>
              <SectionLabel icon={ICONS.file(12)} text="Remarks" />
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
    </Modal>
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

function MovementCell({ role, party }: {
  role: 'Source' | 'Destination'
  party: { label: string; kind: 'wallet' | 'contact' | 'income' | 'expense' }
}) {
  // Real entities (wallet/contact) read solid; the "world" income/expense node is
  // a soft, tinted placeholder so it stays visually secondary to the wallet.
  const isWorld = party.kind === 'income' || party.kind === 'expense'
  const isContact = party.kind === 'contact'
  const style = {
    wallet: { tone: 'var(--color-text-tertiary)', bg: 'var(--color-bg)', icon: ICONS.wallet(14) },
    contact: { tone: 'var(--color-primary)', bg: 'var(--color-primary-subtle)', icon: ICONS.user(14) },
    income: { tone: 'var(--color-success)', bg: 'var(--color-success-subtle)', icon: ICONS.trendingUp(14) },
    expense: { tone: 'var(--color-danger)', bg: 'var(--color-danger-subtle)', icon: ICONS.trendingDown(14) },
  }[party.kind]

  const badge = (
    <span style={{
      width: 22, height: 22, borderRadius: 6,
      // On the tinted contact pill the icon chip flips to surface so it still reads.
      background: isContact ? 'var(--color-surface)' : style.bg,
      color: style.tone,
      display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0,
    }}>
      {style.icon}
    </span>
  )

  return (
    <div style={{ minWidth: 0, display: 'flex', flex: 1, justifyContent: role === 'Source' ? 'flex-end' : 'flex-start' }}>
      {/* Contact node wears a tinted pill so a person stands out from plain wallet nodes. */}
      <div style={{
        minWidth: 0, display: 'flex', alignItems: 'center', gap: 8,
        ...(isContact ? { background: 'var(--color-primary-subtle)', padding: '4px 10px', borderRadius: 999 } : {}),
      }}>
        {role === 'Destination' && badge}
        <div style={{
          fontSize: 14, fontWeight: 600,
          color: isWorld ? style.tone : isContact ? 'var(--color-primary)' : 'var(--color-text-primary)',
          whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis',
        }}>
          {party.label}
        </div>
        {role === 'Source' && badge}
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
