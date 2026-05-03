import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTransactions } from '../../hooks/useTransactions'
import { useWallets } from '../../hooks/useWallets'
import { useContacts } from '../../hooks/useContacts'
import { useSearch } from '../../context/SearchContext'
import { fmt } from '../../lib/formatter'
import { ICONS } from '../ui/Icons'

const PER_SECTION = 5

interface SearchResultsProps {
  anchorTop: number
  onClose: () => void
}

export default function SearchResults({ anchorTop, onClose }: SearchResultsProps) {
  const { searchTerm, setSearchTerm } = useSearch()
  const navigate = useNavigate()
  const debounced = useDebounced(searchTerm, 200)
  const term = debounced.trim().toLowerCase()

  const { data: txnResp } = useTransactions()
  const { data: wallets } = useWallets()
  const { data: contacts } = useContacts()

  const txns = txnResp?.data ?? []

  const txnMatches = useMemo(() => {
    if (!term) return []
    return txns.filter(t =>
      (t.remarks || '').toLowerCase().includes(term) ||
      (t.subcategoryId || '').toLowerCase().includes(term) ||
      (t.contactName || '').toLowerCase().includes(term) ||
      (t.srcId || '').toLowerCase().includes(term) ||
      (t.dstId || '').toLowerCase().includes(term) ||
      String(t.amount).includes(term)
    )
  }, [txns, term])

  const walletMatches = useMemo(() => {
    if (!term || !wallets) return []
    return wallets.filter(w =>
      w.name.toLowerCase().includes(term) ||
      (w.shortName || '').toLowerCase().includes(term)
    )
  }, [wallets, term])

  const contactMatches = useMemo(() => {
    if (!term || !contacts) return []
    return contacts.filter(c =>
      (c.nickName || '').toLowerCase().includes(term) ||
      (c.fullName || '').toLowerCase().includes(term)
    )
  }, [contacts, term])

  const settingMatches = useMemo(() => {
    if (!term) return []
    return SETTINGS_SECTIONS.filter(s => s.toLowerCase().includes(term))
  }, [term])

  if (!term) return null

  const total = txnMatches.length + walletMatches.length + contactMatches.length + settingMatches.length

  const go = (path: string) => {
    setSearchTerm('')
    onClose()
    navigate(path)
  }

  return (
    <div
      style={{
        position: 'fixed',
        top: anchorTop,
        right: 16,
        width: 'min(420px, calc(100vw - 32px))',
        maxHeight: '70vh',
        background: 'var(--color-surface)',
        borderRadius: 'var(--radius-lg)',
        border: '1px solid var(--color-border)',
        boxShadow: '0 16px 48px rgba(0,0,0,0.14)',
        zIndex: 250,
        overflow: 'hidden',
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      <div style={{ padding: '10px 16px', borderBottom: '1px solid var(--color-border)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <span style={{ fontSize: 11, fontWeight: 700, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.08em' }}>
          {total} {total === 1 ? 'result' : 'results'}
        </span>
        <span style={{ fontSize: 11, color: 'var(--color-text-tertiary)' }}>"{searchTerm}"</span>
      </div>
      <div style={{ overflowY: 'auto', flex: 1 }}>
        {total === 0 && (
          <div style={{ padding: 24, textAlign: 'center' }}>
            <p style={{ fontSize: 13, color: 'var(--color-text-tertiary)', margin: 0 }}>No results across pages</p>
          </div>
        )}

        <Section title="Transactions" count={txnMatches.length} icon={ICONS.transactions(13)}>
          {txnMatches.slice(0, PER_SECTION).map(t => {
            const date = new Date(t.timestamp * 1000).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
            const color = t.type === 'Income' ? 'var(--color-success)' : t.type === 'Transfer' ? 'var(--color-primary)' : 'var(--color-danger)'
            const sign = t.type === 'Income' ? '+' : t.type === 'Transfer' ? '' : '-'
            return (
              <Row key={t.id} onClick={() => go(`/transactions?show=${t.id}`)}>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--color-text-primary)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {t.remarks || t.subcategoryId}
                  </div>
                  <div style={{ fontSize: 11, color: 'var(--color-text-tertiary)', marginTop: 2 }}>
                    {date} · {t.subcategoryId}
                  </div>
                </div>
                <div style={{ fontSize: 13, fontWeight: 700, color, flexShrink: 0, marginLeft: 12 }}>
                  {sign}{fmt(t.amount)}
                </div>
              </Row>
            )
          })}
          {txnMatches.length > PER_SECTION && (
            <MoreRow onClick={() => go('/transactions')}>
              +{txnMatches.length - PER_SECTION} more in Transactions
            </MoreRow>
          )}
        </Section>

        <Section title="Wallets" count={walletMatches.length} icon={ICONS.wallet(13)}>
          {walletMatches.slice(0, PER_SECTION).map(w => (
            <Row key={w.shortName} onClick={() => go('/wallets')}>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--color-text-primary)' }}>{w.name}</div>
                <div style={{ fontSize: 11, color: 'var(--color-text-tertiary)', marginTop: 2 }}>{w.shortName}</div>
              </div>
              <div style={{ fontSize: 13, fontWeight: 700, color: 'var(--color-text-secondary)', flexShrink: 0, marginLeft: 12 }}>
                {fmt(w.balance ?? 0)}
              </div>
            </Row>
          ))}
        </Section>

        <Section title="Contacts" count={contactMatches.length} icon={ICONS.user(13)}>
          {contactMatches.slice(0, PER_SECTION).map(c => (
            <Row key={c.nickName} onClick={() => go('/wallets')}>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--color-text-primary)' }}>{c.fullName || c.nickName}</div>
                <div style={{ fontSize: 11, color: 'var(--color-text-tertiary)', marginTop: 2 }}>{c.nickName}</div>
              </div>
            </Row>
          ))}
        </Section>

        <Section title="Settings" count={settingMatches.length} icon={ICONS.settings(13)}>
          {settingMatches.map(s => (
            <Row key={s} onClick={() => go('/settings')}>
              <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--color-text-primary)' }}>{s}</div>
            </Row>
          ))}
        </Section>
      </div>
    </div>
  )
}

function Section({ title, count, icon, children }: { title: string; count: number; icon: React.ReactNode; children: React.ReactNode }) {
  if (count === 0) return null
  return (
    <div>
      <div style={{
        padding: '8px 16px',
        background: 'var(--color-bg)',
        display: 'flex',
        alignItems: 'center',
        gap: 6,
        color: 'var(--color-text-tertiary)',
      }}>
        {icon}
        <span style={{ fontSize: 10, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.08em' }}>
          {title} · {count}
        </span>
      </div>
      <div>{children}</div>
    </div>
  )
}

function Row({ children, onClick }: { children: React.ReactNode; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className="search-result-row"
      style={{
        width: '100%',
        background: 'transparent',
        border: 'none',
        borderBottom: '1px solid var(--color-border)',
        padding: '10px 16px',
        display: 'flex',
        alignItems: 'center',
        cursor: 'pointer',
        textAlign: 'left',
        fontFamily: 'inherit',
        transition: 'background 0.15s',
      }}
      onMouseEnter={e => (e.currentTarget.style.background = 'var(--color-bg)')}
      onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}
    >
      {children}
    </button>
  )
}

function MoreRow({ children, onClick }: { children: React.ReactNode; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      style={{
        width: '100%',
        background: 'transparent',
        border: 'none',
        borderBottom: '1px solid var(--color-border)',
        padding: '8px 16px',
        fontSize: 12,
        fontWeight: 600,
        color: 'var(--color-primary)',
        cursor: 'pointer',
        textAlign: 'left',
        fontFamily: 'inherit',
      }}
    >
      {children}
    </button>
  )
}

function useDebounced<T>(value: T, delay: number): T {
  const [v, setV] = useState(value)
  useEffect(() => {
    const id = setTimeout(() => setV(value), delay)
    return () => clearTimeout(id)
  }, [value, delay])
  return v
}

const SETTINGS_SECTIONS = [
  'Profile',
  'Personal Information',
  'Contact',
  'Security',
  'Timezone',
  'Dark Mode',
  'Logout',
]
