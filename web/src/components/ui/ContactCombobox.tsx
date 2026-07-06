import { useMemo, useState, type KeyboardEvent } from 'react'
import type { Contact } from '../../types'

interface Props {
  label: string
  contacts: Contact[]
  value: string
  onChange: (v: string) => void
  error?: string
}

const labelStyle: React.CSSProperties = {
  fontSize: 10, fontWeight: 700, textTransform: 'uppercase',
  letterSpacing: '0.08em', color: 'var(--color-text-tertiary)', marginLeft: 4,
}

const fieldBase: React.CSSProperties = {
  width: '100%', background: 'var(--color-bg)', borderRadius: 12,
  padding: '12px 16px', fontSize: 14, fontWeight: 500,
  color: 'var(--color-text-primary)', outline: 'none', transition: 'all 0.15s ease',
}

// ContactCombobox is a single-select, tokenized contact picker: type to search
// (best match preselected, Enter commits), free text allowed. Once committed the
// value is atomic — no character editing; Backspace/× clears the whole name.
export default function ContactCombobox({ label, contacts, value, onChange, error }: Props) {
  const [query, setQuery] = useState('')
  const [open, setOpen] = useState(false)
  const [highlight, setHighlight] = useState(0)

  const matches = useMemo(() => {
    const q = query.trim().toLowerCase()
    const rank = (c: Contact) => {
      const nick = c.nickName.toLowerCase()
      const full = (c.fullName || '').toLowerCase()
      if (!q) return 1
      if (nick === q || full === q) return 100
      if (nick.startsWith(q) || full.startsWith(q)) return 60
      if (nick.includes(q) || full.includes(q)) return 30
      return 0
    }
    return contacts
      .map(c => ({ c, s: rank(c) }))
      .filter(x => x.s > 0)
      .sort((a, b) => b.s - a.s)
      .slice(0, 6)
      .map(x => x.c)
  }, [contacts, query])

  // Committed value is the stable key: a known contact's nickName, or free text.
  const commitContact = (c: Contact) => commit(c.nickName)
  const commit = (v: string) => {
    onChange(v)
    setQuery('')
    setOpen(false)
    setHighlight(0)
  }
  const clear = () => { onChange(''); setQuery(''); setOpen(false); setHighlight(0) }

  // Token/chip mode — a value is committed and can only be cleared via the ×.
  // Show the contact's full name, but the underlying value stays the nickName.
  if (value) {
    const display = contacts.find(c => c.nickName === value)?.fullName || value
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 6, width: '100%' }}>
        <span style={labelStyle}>{label}</span>
        <div style={{
          ...fieldBase, border: error ? '1px solid var(--color-danger)' : '1px solid var(--color-border)',
          display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 8, cursor: 'default',
        }}>
          <span style={{
            display: 'inline-flex', alignItems: 'center', gap: 6, maxWidth: '100%',
            padding: '2px 10px', borderRadius: 999, fontWeight: 600,
            background: 'var(--color-primary-subtle)', color: 'var(--color-primary)',
            overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
          }}>
            {display}
          </span>
          <button
            type="button"
            aria-label="Clear contact"
            onClick={clear}
            style={{ border: 'none', background: 'transparent', cursor: 'pointer', color: 'var(--color-text-tertiary)', fontSize: 18, lineHeight: 1, padding: 0 }}
          >
            ×
          </button>
        </div>
        {error && <span style={{ fontSize: 11, color: 'var(--color-danger)', marginLeft: 4 }}>{error}</span>}
      </div>
    )
  }

  const onKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault(); setOpen(true); setHighlight(h => Math.min(h + 1, matches.length - 1))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault(); setHighlight(h => Math.max(h - 1, 0))
    } else if (e.key === 'Enter') {
      e.preventDefault()
      if (open && matches.length) commitContact(matches[Math.min(highlight, matches.length - 1)])
      else if (query.trim()) commit(query.trim())
    } else if (e.key === 'Escape') {
      setOpen(false)
    }
  }

  return (
    <label style={{ display: 'flex', flexDirection: 'column', gap: 6, width: '100%', position: 'relative' }}>
      <span style={labelStyle}>{label}</span>
      <input
        value={query}
        onChange={e => { setQuery(e.target.value); setOpen(true); setHighlight(0) }}
        onKeyDown={onKeyDown}
        onFocus={e => { setOpen(true); e.currentTarget.style.border = '1px solid var(--color-primary)'; e.currentTarget.style.boxShadow = '0 0 0 4px var(--color-primary-subtle)' }}
        onBlur={e => { e.currentTarget.style.border = error ? '1px solid var(--color-danger)' : '1px solid var(--color-border)'; e.currentTarget.style.boxShadow = 'none'; if (query.trim()) commit(query.trim()) }}
        placeholder="Search or type a name…"
        style={{ ...fieldBase, border: error ? '1px solid var(--color-danger)' : '1px solid var(--color-border)' }}
      />
      {open && matches.length > 0 && (
        <div style={{
          position: 'absolute', zIndex: 20, top: '100%', left: 0, right: 0, marginTop: 4,
          background: 'var(--color-surface)', border: '1px solid var(--color-border)',
          borderRadius: 12, boxShadow: '0 8px 24px rgba(0,0,0,0.12)', overflow: 'hidden',
        }}>
          {matches.map((c, i) => (
            <button
              key={c.nickName}
              type="button"
              onMouseDown={e => { e.preventDefault(); commitContact(c) }}
              onMouseMove={() => setHighlight(i)}
              style={{
                display: 'flex', flexDirection: 'column', alignItems: 'flex-start', gap: 2,
                width: '100%', padding: '10px 16px', border: 'none',
                background: i === highlight ? 'var(--color-hover)' : 'transparent',
                cursor: 'pointer', textAlign: 'left',
              }}
            >
              <span style={{ fontSize: 14, fontWeight: 600, color: 'var(--color-text-primary)' }}>{c.fullName || c.nickName}</span>
              <span style={{ fontSize: 11, color: 'var(--color-text-tertiary)' }}>{c.nickName}</span>
            </button>
          ))}
        </div>
      )}
      {error && <span style={{ fontSize: 11, color: 'var(--color-danger)', marginLeft: 4 }}>{error}</span>}
    </label>
  )
}
